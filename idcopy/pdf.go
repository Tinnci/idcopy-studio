package idcopy

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	pointsPerMM     = 72.0 / 25.4
	maxEncodedImage = 48 << 20
	maxImagePixels  = 60_000_000
	defaultPDFName  = "id-copy.pdf"
)

type rasterImage struct {
	width      int
	height     int
	compressed []byte
	filter     string
	colorSpace string
}

func GeneratePDF(request LayoutRequest) ([]byte, LayoutPlan, error) {
	plan, err := Plan(request)
	if err != nil {
		return nil, LayoutPlan{}, err
	}
	fit := request.Fit
	if fit == "" {
		fit = "cover"
	}
	if fit != "cover" && fit != "contain" && fit != "stretch" {
		return nil, LayoutPlan{}, fmt.Errorf("unsupported image fit %q", fit)
	}

	needed := map[string]bool{}
	for _, item := range plan.Items {
		needed[item.Side] = true
	}
	images := map[string]rasterImage{}
	for _, side := range []string{"front", "back"} {
		if !needed[side] {
			continue
		}
		input := request.Front
		if side == "back" {
			input = request.Back
		}
		if strings.TrimSpace(input.DataURL) == "" {
			return nil, LayoutPlan{}, fmt.Errorf("%s image is required", side)
		}
		decoded, decodeErr := decodeImage(input)
		if decodeErr != nil {
			return nil, LayoutPlan{}, fmt.Errorf("decode %s image: %w", side, decodeErr)
		}
		images[side] = decoded
	}

	pdf, err := buildPDF(plan, images, fit, request.CutMarks, request.PDF)
	if err != nil {
		return nil, LayoutPlan{}, err
	}
	return pdf, plan, nil
}

func SavePDF(path string, request LayoutRequest) (SaveResult, error) {
	pdf, plan, err := GeneratePDF(request)
	if err != nil {
		return SaveResult{}, err
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		path += ".pdf"
	}
	if err = os.WriteFile(path, pdf, 0o600); err != nil {
		return SaveResult{}, fmt.Errorf("save PDF: %w", err)
	}
	return SaveResult{Path: path, Pages: plan.Pages, Bytes: len(pdf)}, nil
}

func Export(request LayoutRequest) (ExportResult, error) {
	pdf, plan, err := GeneratePDF(request)
	if err != nil {
		return ExportResult{}, err
	}
	return ExportResult{
		Filename:   defaultPDFName,
		DataBase64: base64.StdEncoding.EncodeToString(pdf),
		Pages:      plan.Pages,
		Bytes:      len(pdf),
	}, nil
}

func decodeImage(input ImageInput) (rasterImage, error) {
	rotation := ((input.Rotation % 360) + 360) % 360
	if rotation != 0 && rotation != 90 && rotation != 180 && rotation != 270 {
		return rasterImage{}, errors.New("rotation must be 0, 90, 180 or 270 degrees")
	}

	payload := input.DataURL
	if comma := strings.IndexByte(payload, ','); comma >= 0 {
		header := payload[:comma]
		if !strings.Contains(header, ";base64") {
			return rasterImage{}, errors.New("image data URL must use base64 encoding")
		}
		payload = payload[comma+1:]
	}
	if len(payload) > base64.StdEncoding.EncodedLen(maxEncodedImage) {
		return rasterImage{}, errors.New("encoded image is too large")
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return rasterImage{}, errors.New("invalid base64 image")
	}
	if len(raw) == 0 || len(raw) > maxEncodedImage {
		return rasterImage{}, errors.New("image must be between 1 byte and 48 MiB")
	}

	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return rasterImage{}, errors.New("unsupported or corrupt image")
	}
	if format != "jpeg" && format != "png" {
		return rasterImage{}, fmt.Errorf("unsupported image format %q; use JPEG or PNG", format)
	}
	width, height := config.Width, config.Height
	if width < 1 || height < 1 || int64(width)*int64(height) > maxImagePixels {
		return rasterImage{}, errors.New("image dimensions are invalid or exceed 60 megapixels")
	}
	source, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return rasterImage{}, errors.New("unsupported or corrupt image")
	}
	if format == "jpeg" && rotation == 0 {
		switch source.(type) {
		case *image.Gray:
			return rasterImage{width: width, height: height, compressed: raw, filter: "DCTDecode", colorSpace: "DeviceGray"}, nil
		case *image.YCbCr:
			return rasterImage{width: width, height: height, compressed: raw, filter: "DCTDecode", colorSpace: "DeviceRGB"}, nil
		}
	}
	bounds := source.Bounds()

	outputWidth, outputHeight := width, height
	if rotation == 90 || rotation == 270 {
		outputWidth, outputHeight = height, width
	}

	var compressed bytes.Buffer
	compressor, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	if err != nil {
		return rasterImage{}, err
	}
	row := make([]byte, outputWidth*3)
	for y := 0; y < outputHeight; y++ {
		for x := 0; x < outputWidth; x++ {
			sx, sy := rotatedSourceCoordinate(x, y, width, height, rotation)
			r, g, b, a := source.At(bounds.Min.X+sx, bounds.Min.Y+sy).RGBA()
			// RGBA values are alpha-premultiplied. Composite transparent pixels onto white.
			r += 0xffff - a
			g += 0xffff - a
			b += 0xffff - a
			row[x*3] = byte(r >> 8)
			row[x*3+1] = byte(g >> 8)
			row[x*3+2] = byte(b >> 8)
		}
		if _, err = compressor.Write(row); err != nil {
			_ = compressor.Close()
			return rasterImage{}, err
		}
	}
	if err = compressor.Close(); err != nil {
		return rasterImage{}, err
	}
	return rasterImage{width: outputWidth, height: outputHeight, compressed: compressed.Bytes(), filter: "FlateDecode", colorSpace: "DeviceRGB"}, nil
}

func rotatedSourceCoordinate(x, y, width, height, rotation int) (int, int) {
	switch rotation {
	case 90:
		return y, height - 1 - x
	case 180:
		return width - 1 - x, height - 1 - y
	case 270:
		return width - 1 - y, x
	default:
		return x, y
	}
}

func buildPDF(plan LayoutPlan, images map[string]rasterImage, fit string, cutMarks bool, metadata PDFOptions) ([]byte, error) {
	objects := map[int][]byte{}
	imageObjectIDs := map[string]int{}
	objects[3] = pdfInfoObject(metadata)
	nextObjectID := 4
	for _, side := range []string{"front", "back"} {
		image, ok := images[side]
		if !ok {
			continue
		}
		imageObjectIDs[side] = nextObjectID
		objects[nextObjectID] = streamObject(fmt.Sprintf(
			"<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /%s /BitsPerComponent 8 /Filter /%s /Length %d >>",
			image.width, image.height, image.colorSpace, image.filter, len(image.compressed)), image.compressed)
		nextObjectID++
	}

	pageObjectIDs := make([]int, plan.Pages)
	contentObjectIDs := make([]int, plan.Pages)
	for index := range plan.Pages {
		pageObjectIDs[index] = nextObjectID
		contentObjectIDs[index] = nextObjectID + 1
		nextObjectID += 2
	}

	objects[1] = []byte("<< /Type /Catalog /Pages 2 0 R >>")
	kids := make([]string, 0, len(pageObjectIDs))
	for _, id := range pageObjectIDs {
		kids = append(kids, fmt.Sprintf("%d 0 R", id))
	}
	objects[2] = []byte(fmt.Sprintf("<< /Type /Pages /Count %d /Kids [%s] >>", plan.Pages, strings.Join(kids, " ")))

	pageWidth := plan.Paper.Width * pointsPerMM
	pageHeight := plan.Paper.Height * pointsPerMM
	resources := pdfResources(imageObjectIDs)
	for pageIndex := range plan.Pages {
		var content strings.Builder
		for _, item := range plan.Items {
			if item.Page != pageIndex+1 {
				continue
			}
			image, ok := images[item.Side]
			if !ok {
				return nil, fmt.Errorf("missing image for %s", item.Side)
			}
			content.WriteString(drawImage(item, plan.Paper, image, fit, pdfImageName(item.Side)))
			if cutMarks {
				content.WriteString(drawCutMarks(item, plan.Paper))
			}
		}
		contentBytes := []byte(content.String())
		objects[contentObjectIDs[pageIndex]] = streamObject(fmt.Sprintf("<< /Length %d >>", len(contentBytes)), contentBytes)
		objects[pageObjectIDs[pageIndex]] = []byte(fmt.Sprintf(
			"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %s %s] /Resources %s /Contents %d 0 R >>",
			pdfNumber(pageWidth), pdfNumber(pageHeight), resources, contentObjectIDs[pageIndex]))
	}

	return serializePDF(objects, nextObjectID-1)
}

func pdfInfoObject(metadata PDFOptions) []byte {
	entries := []string{"/Producer " + pdfTextString("ID Copy Studio")}
	if metadata.Title != "" {
		entries = append(entries, "/Title "+pdfTextString(metadata.Title))
	}
	if metadata.Author != "" {
		entries = append(entries, "/Author "+pdfTextString(metadata.Author))
	}
	if metadata.Subject != "" {
		entries = append(entries, "/Subject "+pdfTextString(metadata.Subject))
	}
	if metadata.Keywords != "" {
		entries = append(entries, "/Keywords "+pdfTextString(metadata.Keywords))
	}
	return []byte("<< " + strings.Join(entries, " ") + " >>")
}

func pdfTextString(value string) string {
	var result strings.Builder
	result.WriteString("<FEFF")
	for _, character := range value {
		if character <= 0xffff {
			fmt.Fprintf(&result, "%04X", character)
			continue
		}
		character -= 0x10000
		high := 0xd800 + (character >> 10)
		low := 0xdc00 + (character & 0x3ff)
		fmt.Fprintf(&result, "%04X%04X", high, low)
	}
	result.WriteByte('>')
	return result.String()
}

func pdfResources(imageObjectIDs map[string]int) string {
	keys := make([]string, 0, len(imageObjectIDs))
	for side := range imageObjectIDs {
		keys = append(keys, side)
	}
	sort.Strings(keys)
	var entries []string
	for _, side := range keys {
		entries = append(entries, fmt.Sprintf("/%s %d 0 R", pdfImageName(side), imageObjectIDs[side]))
	}
	return fmt.Sprintf("<< /XObject << %s >> >>", strings.Join(entries, " "))
}

func pdfImageName(side string) string {
	if side == "back" {
		return "ImBack"
	}
	return "ImFront"
}

func drawImage(item PlacedItem, paper SizeMM, image rasterImage, fit, imageName string) string {
	x := item.X * pointsPerMM
	y := (paper.Height - item.Y - item.Height) * pointsPerMM
	w := item.Width * pointsPerMM
	h := item.Height * pointsPerMM
	drawX, drawY, drawW, drawH := x, y, w, h
	imageAspect := float64(image.width) / float64(image.height)
	targetAspect := w / h
	if fit == "cover" {
		if imageAspect > targetAspect {
			drawW = h * imageAspect
			drawX = x - (drawW-w)/2
		} else {
			drawH = w / imageAspect
			drawY = y - (drawH-h)/2
		}
	} else if fit == "contain" {
		if imageAspect > targetAspect {
			drawH = w / imageAspect
			drawY = y + (h-drawH)/2
		} else {
			drawW = h * imageAspect
			drawX = x + (w-drawW)/2
		}
	}
	return fmt.Sprintf(
		"q\n%s %s %s %s re W n\n%s 0 0 %s %s %s cm\n/%s Do\nQ\n",
		pdfNumber(x), pdfNumber(y), pdfNumber(w), pdfNumber(h),
		pdfNumber(drawW), pdfNumber(drawH), pdfNumber(drawX), pdfNumber(drawY), imageName)
}

func drawCutMarks(item PlacedItem, paper SizeMM) string {
	x := item.X * pointsPerMM
	y := (paper.Height - item.Y - item.Height) * pointsPerMM
	w := item.Width * pointsPerMM
	h := item.Height * pointsPerMM
	offset := 1.2 * pointsPerMM
	length := 3.5 * pointsPerMM
	return fmt.Sprintf(
		"q 0.35 G 0.35 w\n"+
			"%s %s m %s %s l %s %s m %s %s l\n"+
			"%s %s m %s %s l %s %s m %s %s l\n"+
			"%s %s m %s %s l %s %s m %s %s l\n"+
			"%s %s m %s %s l %s %s m %s %s l S Q\n",
		pdfNumber(x-offset-length), pdfNumber(y), pdfNumber(x-offset), pdfNumber(y), pdfNumber(x), pdfNumber(y-offset-length), pdfNumber(x), pdfNumber(y-offset),
		pdfNumber(x+w+offset), pdfNumber(y), pdfNumber(x+w+offset+length), pdfNumber(y), pdfNumber(x+w), pdfNumber(y-offset-length), pdfNumber(x+w), pdfNumber(y-offset),
		pdfNumber(x-offset-length), pdfNumber(y+h), pdfNumber(x-offset), pdfNumber(y+h), pdfNumber(x), pdfNumber(y+h+offset), pdfNumber(x), pdfNumber(y+h+offset+length),
		pdfNumber(x+w+offset), pdfNumber(y+h), pdfNumber(x+w+offset+length), pdfNumber(y+h), pdfNumber(x+w), pdfNumber(y+h+offset), pdfNumber(x+w), pdfNumber(y+h+offset+length))
}

func streamObject(dictionary string, stream []byte) []byte {
	var result bytes.Buffer
	result.WriteString(dictionary)
	result.WriteString("\nstream\n")
	result.Write(stream)
	result.WriteString("\nendstream")
	return result.Bytes()
}

func serializePDF(objects map[int][]byte, objectCount int) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	offsets := make([]int64, objectCount+1)
	for id := 1; id <= objectCount; id++ {
		body, ok := objects[id]
		if !ok {
			return nil, fmt.Errorf("missing PDF object %d", id)
		}
		offsets[id] = int64(output.Len())
		fmt.Fprintf(&output, "%d 0 obj\n", id)
		output.Write(body)
		output.WriteString("\nendobj\n")
	}
	xrefOffset := output.Len()
	fmt.Fprintf(&output, "xref\n0 %d\n", objectCount+1)
	output.WriteString("0000000000 65535 f \n")
	for id := 1; id <= objectCount; id++ {
		fmt.Fprintf(&output, "%010d 00000 n \n", offsets[id])
	}
	fmt.Fprintf(&output, "trailer\n<< /Size %d /Root 1 0 R /Info 3 0 R >>\nstartxref\n%d\n%%%%EOF\n", objectCount+1, xrefOffset)
	return output.Bytes(), nil
}

func pdfNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', 3, 64)
}
