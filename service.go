package main

import (
	"path/filepath"
	"strings"

	"github.com/Tinnci/idcopy-studio/idcopy"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// IDCopyService exposes the reusable Go layout engine to the Wails frontend.
type IDCopyService struct {
	app *application.App
}

func (s *IDCopyService) Catalog() idcopy.Catalog {
	return idcopy.GetCatalog()
}

func (s *IDCopyService) Preview(request idcopy.LayoutRequest) (idcopy.LayoutPlan, error) {
	return idcopy.Plan(request)
}

func (s *IDCopyService) Generate(request idcopy.LayoutRequest) (idcopy.ExportResult, error) {
	return idcopy.Export(request)
}

func (s *IDCopyService) Save(request idcopy.LayoutRequest, suggestedFilename string) (idcopy.SaveResult, error) {
	filename := strings.TrimSpace(filepath.Base(suggestedFilename))
	if filename == "" || filename == "." {
		filename = "id-copy.pdf"
	}
	if !strings.EqualFold(filepath.Ext(filename), ".pdf") {
		filename += ".pdf"
	}
	dialog := s.app.Dialog.SaveFile().
		SetFilename(filename).
		AddFilter("PDF document", "*.pdf")
	if currentWindow := s.app.Window.Current(); currentWindow != nil {
		dialog.AttachToWindow(currentWindow)
	}
	path, err := dialog.PromptForSingleSelection()
	if err != nil {
		return idcopy.SaveResult{}, err
	}
	if path == "" {
		return idcopy.SaveResult{Canceled: true}, nil
	}
	return idcopy.SavePDF(path, request)
}
