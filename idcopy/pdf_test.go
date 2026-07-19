package idcopy

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testPNGDataURL(t *testing.T, width, height int, fill color.Color) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fill)
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
}

func testJPEGDataURL(t *testing.T, width, height int, fill color.Color) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fill)
		}
	}
	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(encoded.Bytes())
}

func TestGeneratePDF(t *testing.T) {
	request := baseRequest()
	request.Front.DataURL = testPNGDataURL(t, 160, 100, color.NRGBA{R: 220, G: 30, B: 30, A: 255})
	request.Back.DataURL = testPNGDataURL(t, 160, 100, color.NRGBA{R: 30, G: 60, B: 220, A: 255})
	request.CutMarks = true
	pdf, plan, err := GeneratePDF(request)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Pages != 1 {
		t.Fatalf("expected one page, got %d", plan.Pages)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) || !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Fatal("generated data is not a complete PDF")
	}
	text := string(pdf)
	for _, token := range []string{"/MediaBox", "/ImFront", "/ImBack", "/Count 1"} {
		if !strings.Contains(text, token) {
			t.Fatalf("PDF is missing %s", token)
		}
	}
	if _, err := exec.LookPath("pdfinfo"); err == nil {
		command := exec.Command("pdfinfo", "-")
		command.Stdin = bytes.NewReader(pdf)
		info, commandErr := command.CombinedOutput()
		if commandErr != nil {
			t.Fatalf("pdfinfo rejected generated PDF: %v\n%s", commandErr, info)
		}
		if !strings.Contains(string(info), "Pages:") || !strings.Contains(string(info), "A4") {
			t.Fatalf("unexpected pdfinfo output:\n%s", info)
		}
	}
}

func TestJPEGAndUnicodeMetadata(t *testing.T) {
	request := baseRequest()
	request.Sides = "front"
	request.Front.DataURL = testJPEGDataURL(t, 160, 100, color.NRGBA{R: 230, G: 210, B: 120, A: 255})
	request.PDF = PDFOptions{Title: "证件复印件", Author: "本地用户", Subject: "仅供测试"}
	pdf, _, err := GeneratePDF(request)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(pdf, []byte("/DCTDecode")) {
		t.Fatal("JPEG should be embedded with DCTDecode")
	}
	if _, err = exec.LookPath("pdfinfo"); err == nil {
		command := exec.Command("pdfinfo", "-")
		command.Stdin = bytes.NewReader(pdf)
		info, commandErr := command.CombinedOutput()
		if commandErr != nil {
			t.Fatalf("pdfinfo rejected metadata PDF: %v\n%s", commandErr, info)
		}
		if !strings.Contains(string(info), "证件复印件") {
			t.Fatalf("unicode title missing from pdfinfo output:\n%s", info)
		}
	}
}

func TestSavePDFAddsExtension(t *testing.T) {
	request := baseRequest()
	request.Sides = "front"
	request.Front.DataURL = testPNGDataURL(t, 80, 50, color.White)
	path := filepath.Join(t.TempDir(), "copy")
	result, err := SavePDF(path, request)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(result.Path) != ".pdf" {
		t.Fatalf("expected .pdf extension, got %s", result.Path)
	}
	if info, statErr := os.Stat(result.Path); statErr != nil || info.Size() == 0 {
		t.Fatalf("saved PDF missing or empty: %v", statErr)
	}
}

func TestRotateImage(t *testing.T) {
	data := testPNGDataURL(t, 20, 10, color.White)
	image, err := decodeImage(ImageInput{DataURL: data, Rotation: 90})
	if err != nil {
		t.Fatal(err)
	}
	if image.width != 10 || image.height != 20 {
		t.Fatalf("expected 10x20 rotated image, got %dx%d", image.width, image.height)
	}
}
