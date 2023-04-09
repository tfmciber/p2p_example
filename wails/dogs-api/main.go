package main

import (
	"embed"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	var textproto = protocol.ID("/text/1.0.0")
	var audioproto = protocol.ID("/audio/1.0.0")
	var benchproto = protocol.ID("/bench/1.0.0")
	var cmdproto = protocol.ID("/cmd/1.0.0")
	var fileproto = protocol.ID("/file/1.0.0")

	app := &P2Papp{data: make(map[string]struct {
		peers []peer.ID
		timer uint
	}), preferquic: true, refresh: 15, rendezvousS: make(chan string, 1), useradded: make(chan bool), chatadded: make(chan string), textproto: textproto, audioproto: audioproto, benchproto: benchproto, cmdproto: cmdproto, fileproto: fileproto}

	err := wails.Run(&options.App{
		Title:  "P2P",
		Width:  1200,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		WindowStartState: options.Minimised,

		BackgroundColour: &options.RGBA{R: 99, G: 99, B: 99, A: 99},
		MinWidth:         1200,
		MinHeight:        720,
		OnStartup:        app.startup,

		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
