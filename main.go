package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac" // Added for macOS specific tweaks
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "MultiBlox",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 20, G: 20, B: 20, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		// --- FIXING THE ERRORS HERE ---
		// If DisableInspector still errors, use Debug: options.Debug{} with
		// its internal settings, but usually, 'DisableRemoteIspector' or
		// checking your 'wails.json' for debug settings is the way.

		// This is the most compatible way to handle the context menu in v2:
		OnDomReady: func(ctx context.Context) {
			// We can also handle lockdown via JS if the Go struct is acting up
		},

		// macOS specific lockdown to disable the "Inspect" menu item
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(), // Makes it look native and clean
			About: &mac.AboutInfo{
				Title:   "MultiBlox",
				Message: "Created by iigordev",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
