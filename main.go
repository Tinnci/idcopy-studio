package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	service := &IDCopyService{}
	app := application.New(application.Options{
		Name:        "ID Copy Studio",
		Description: "Local international identity-document copy PDF generator",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	service.app = app

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "ID Copy Studio",
		Width:            1320,
		Height:           860,
		MinWidth:         980,
		MinHeight:        680,
		BackgroundColour: application.NewRGB(244, 241, 234),
		URL:              "/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 48,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
