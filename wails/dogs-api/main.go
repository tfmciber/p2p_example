package main

import (
	"context"
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
		Peers []peer.ID
		Timer uint
	}), preferquic: false, refresh: 15, trashchats: make(map[string]bool), messages: make(map[string][]Message), queueFiles: make(map[string][]string), rendezvousS: make(chan string, 1), updateDHT: make(chan bool), useradded: make(chan bool), reloadChat: make(chan string), chatadded: make(chan string), textproto: textproto, audioproto: audioproto, benchproto: benchproto, cmdproto: cmdproto, fileproto: fileproto, cancelRendezvous: make(map[string]context.CancelFunc, 0)}

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
		OnBeforeClose:    app.close,

		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
