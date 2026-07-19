package idcopy

import (
	"errors"
	"fmt"
	"math"
)

const epsilon = 0.0001

type layoutGeometry struct {
	groupWidth  float64
	groupHeight float64
	columns     int
	rows        int
	capacity    int
	sides       []string
}

func Plan(request LayoutRequest) (LayoutPlan, error) {
	if request.Copies < 1 || request.Copies > 100 {
		return LayoutPlan{}, errors.New("copies must be between 1 and 100")
	}
	if request.MarginMM < 0 || request.MarginMM > 50 {
		return LayoutPlan{}, errors.New("margin must be between 0 and 50 mm")
	}
	if request.GapMM < 0 || request.GapMM > 50 {
		return LayoutPlan{}, errors.New("gap must be between 0 and 50 mm")
	}
	if request.ScalePercent == 0 {
		request.ScalePercent = 100
	}
	if request.ScalePercent < 25 || request.ScalePercent > 400 {
		return LayoutPlan{}, errors.New("scale must be between 25% and 400%")
	}

	paper, err := resolvePaper(request)
	if err != nil {
		return LayoutPlan{}, err
	}
	document, err := resolveDocument(request)
	if err != nil {
		return LayoutPlan{}, err
	}
	document.Width *= request.ScalePercent / 100
	document.Height *= request.ScalePercent / 100

	sides, err := resolveSides(request.Sides)
	if err != nil {
		return LayoutPlan{}, err
	}
	arrangement := request.Arrangement
	if arrangement == "" {
		arrangement = "paired"
	}
	if arrangement != "paired" && arrangement != "stacked" && arrangement != "grid" && arrangement != "separate-pages" {
		return LayoutPlan{}, fmt.Errorf("unsupported arrangement %q", arrangement)
	}

	orientation := request.Orientation
	if orientation == "" {
		orientation = "auto"
	}
	if orientation != "auto" && orientation != "portrait" && orientation != "landscape" {
		return LayoutPlan{}, fmt.Errorf("unsupported orientation %q", orientation)
	}

	if orientation == "auto" {
		portraitPaper := portrait(paper)
		landscapePaper := SizeMM{Width: portraitPaper.Height, Height: portraitPaper.Width}
		portraitGeometry := geometry(portraitPaper, document, sides, arrangement, request.MarginMM, request.GapMM)
		landscapeGeometry := geometry(landscapePaper, document, sides, arrangement, request.MarginMM, request.GapMM)
		if portraitGeometry.capacity == 0 && landscapeGeometry.capacity == 0 {
			return LayoutPlan{}, errors.New("document layout does not fit on the selected paper")
		}
		if landscapeGeometry.capacity > portraitGeometry.capacity {
			paper = landscapePaper
			orientation = "landscape"
		} else {
			paper = portraitPaper
			orientation = "portrait"
		}
	} else if orientation == "portrait" {
		paper = portrait(paper)
	} else {
		paper = landscape(paper)
	}

	geometry := geometry(paper, document, sides, arrangement, request.MarginMM, request.GapMM)
	if geometry.capacity == 0 {
		return LayoutPlan{}, errors.New("document layout does not fit on the selected paper")
	}

	plan := LayoutPlan{Paper: paper, Document: document, Orientation: orientation}
	if arrangement == "separate-pages" {
		for copyIndex := 0; copyIndex < request.Copies; copyIndex++ {
			for _, side := range sides {
				page := len(plan.Items) + 1
				plan.Items = append(plan.Items, PlacedItem{
					Page: page, Copy: copyIndex + 1, Side: side,
					X:     (paper.Width - document.Width) / 2,
					Y:     (paper.Height - document.Height) / 2,
					Width: document.Width, Height: document.Height,
				})
			}
		}
		plan.Pages = len(plan.Items)
		return plan, nil
	}

	for copyIndex := 0; copyIndex < request.Copies; copyIndex++ {
		groupIndex := copyIndex % geometry.capacity
		page := copyIndex/geometry.capacity + 1
		column := groupIndex % geometry.columns
		row := groupIndex / geometry.columns
		groupX := request.MarginMM + float64(column)*(geometry.groupWidth+request.GapMM)
		groupY := request.MarginMM + float64(row)*(geometry.groupHeight+request.GapMM)

		for sideIndex, side := range sides {
			x, y := groupX, groupY
			switch arrangement {
			case "paired":
				x += float64(sideIndex) * (document.Width + request.GapMM)
			case "stacked":
				y += float64(sideIndex) * (document.Height + request.GapMM)
			case "grid":
				flatIndex := copyIndex*len(sides) + sideIndex
				page = flatIndex/geometry.capacity + 1
				itemIndex := flatIndex % geometry.capacity
				x = request.MarginMM + float64(itemIndex%geometry.columns)*(document.Width+request.GapMM)
				y = request.MarginMM + float64(itemIndex/geometry.columns)*(document.Height+request.GapMM)
			}
			plan.Items = append(plan.Items, PlacedItem{
				Page: page, Copy: copyIndex + 1, Side: side,
				X: x, Y: y, Width: document.Width, Height: document.Height,
			})
		}
	}
	for _, item := range plan.Items {
		if item.Page > plan.Pages {
			plan.Pages = item.Page
		}
	}
	return plan, nil
}

func resolvePaper(request LayoutRequest) (SizeMM, error) {
	if request.PaperID == "custom" {
		if err := validateSize(request.CustomPaper, "custom paper", 50, 2000); err != nil {
			return SizeMM{}, err
		}
		return request.CustomPaper, nil
	}
	paper, ok := findPaper(request.PaperID)
	if !ok {
		return SizeMM{}, fmt.Errorf("unknown paper size %q", request.PaperID)
	}
	return SizeMM{Width: paper.Width, Height: paper.Height}, nil
}

func resolveDocument(request LayoutRequest) (SizeMM, error) {
	if request.DocumentID == "custom" {
		if err := validateSize(request.CustomDocument, "custom document", 5, 1000); err != nil {
			return SizeMM{}, err
		}
		return request.CustomDocument, nil
	}
	document, ok := findDocument(request.DocumentID)
	if !ok {
		return SizeMM{}, fmt.Errorf("unknown document size %q", request.DocumentID)
	}
	return SizeMM{Width: document.Width, Height: document.Height}, nil
}

func validateSize(size SizeMM, name string, minimum, maximum float64) error {
	if size.Width < minimum || size.Height < minimum || size.Width > maximum || size.Height > maximum {
		return fmt.Errorf("%s dimensions must be between %.0f and %.0f mm", name, minimum, maximum)
	}
	return nil
}

func resolveSides(value string) ([]string, error) {
	switch value {
	case "", "both":
		return []string{"front", "back"}, nil
	case "front":
		return []string{"front"}, nil
	case "back":
		return []string{"back"}, nil
	default:
		return nil, fmt.Errorf("unsupported sides %q", value)
	}
}

func geometry(paper, document SizeMM, sides []string, arrangement string, margin, gap float64) layoutGeometry {
	availableWidth := paper.Width - 2*margin
	availableHeight := paper.Height - 2*margin
	if availableWidth <= 0 || availableHeight <= 0 {
		return layoutGeometry{sides: sides}
	}

	groupWidth, groupHeight := document.Width, document.Height
	switch arrangement {
	case "paired":
		groupWidth = float64(len(sides))*document.Width + float64(len(sides)-1)*gap
	case "stacked":
		groupHeight = float64(len(sides))*document.Height + float64(len(sides)-1)*gap
	case "grid":
		// Each side is an independent grid item.
	case "separate-pages":
		if document.Width <= availableWidth+epsilon && document.Height <= availableHeight+epsilon {
			return layoutGeometry{groupWidth: document.Width, groupHeight: document.Height, columns: 1, rows: 1, capacity: 1, sides: sides}
		}
		return layoutGeometry{sides: sides}
	}

	columns := int(math.Floor((availableWidth + gap + epsilon) / (groupWidth + gap)))
	rows := int(math.Floor((availableHeight + gap + epsilon) / (groupHeight + gap)))
	if columns < 1 || rows < 1 {
		return layoutGeometry{groupWidth: groupWidth, groupHeight: groupHeight, sides: sides}
	}
	capacity := columns * rows
	if arrangement == "grid" {
		capacity = columns * rows
	}
	return layoutGeometry{groupWidth: groupWidth, groupHeight: groupHeight, columns: columns, rows: rows, capacity: capacity, sides: sides}
}

func portrait(size SizeMM) SizeMM {
	if size.Width <= size.Height {
		return size
	}
	return SizeMM{Width: size.Height, Height: size.Width}
}

func landscape(size SizeMM) SizeMM {
	if size.Width >= size.Height {
		return size
	}
	return SizeMM{Width: size.Height, Height: size.Width}
}
