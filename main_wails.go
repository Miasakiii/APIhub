//go:build wails

package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewWailsApp()

	err := wails.Run(&options.App{
		Title:  "APIHub",
		Width:  1280,
		Height: 800,
		MinWidth: 800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  app.startup,
		OnDomReady: app.domReady,
		OnShutdown: app.shutdown,
		OnBeforeClose: func(ctx context.Context) bool {
			return app.onBeforeClose()
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			Theme:                windows.SystemDefault,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
			About: &mac.AboutInfo{
				Title:   "APIHub",
				Message: "Personal LLM API Monitoring Dashboard",
			},
		},
		Linux: &linux.Options{
			ProgramName: "APIHub",
		},
	})

	if err != nil {
		log.Fatal("Failed to start Wails app:", err)
	}
}
