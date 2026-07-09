package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "JS Kern.md",
		Width:     1180,
		Height:    760,
		MinWidth:  940,
		MinHeight: 620,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: app.assetHandler(),
		},
		BackgroundColour: &options.RGBA{R: 15, G: 15, B: 17, A: 1},
		OnStartup:        app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: productInfo.AppID,
			OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
				_ = app.queueLaunchArgs(data.Args, data.WorkingDirectory)
			},
		},
		EnableDefaultContextMenu: false,
		Windows: &windows.Options{
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
			BackdropType:         windows.Mica,
			DisablePinchZoom:     true,
			Theme:                windows.SystemDefault,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
