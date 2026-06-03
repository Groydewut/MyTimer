package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := &App{}
	err := wails.Run(&options.App{
		Title:  "Go Timer",
		Width:  500,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.OnStartup,
		Bind:      []interface{}{app}, // Привязка App к JS
		// ... остальные настройки
	})
	if err != nil {
		println("Error:", err.Error())
	}

}
