package main

import (
	"fmt"
	filepath "path/filepath"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (c *P2Papp) EmitEvent(event string, data ...interface{}) {
	runtime.EventsEmit(c.ctx, event, data)

}

func (c *P2Papp) SelectFiles() []PathFilename {

	var pathFilenames []PathFilename

	file, err := runtime.OpenMultipleFilesDialog(c.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		return nil
	}
	//get file sizes
	i := 0
	for _, f := range file {
		filename := filepath.Base(f)
		pathFilenames = append(pathFilenames, PathFilename{Path: f, Filename: filename})

		i += 1
	}

	return pathFilenames
}

func (c *P2Papp) DataChanged() {

	go func() {
		for {
			select {
			case <-c.chatadded:
				aux := c.ListChats()
				runtime.EventsEmit(c.ctx, "updateChats", aux)
			case <-c.useradded:
				aux := c.ListUsers()
				runtime.EventsEmit(c.ctx, "updateUsers", aux)
			case <-c.ctx.Done():

				return
			}
		}
	}()
}
func (c *P2Papp) fmtPrintln(args ...interface{}) {
	output, err := PrintToVariable(args...)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
	runtime.EventsEmit(c.ctx, "receiveCommands", output)
}
func (c *P2Papp) AddDm(peerid peer.ID) {

	if !contains(c.direcmessages, peerid) {
		c.direcmessages = append(c.direcmessages, peerid)
		//convert []peer.ID to []string
		var peerids []string
		for _, v := range c.direcmessages {
			peerids = append(peerids, v.String())
		}
		runtime.EventsEmit(c.ctx, "directMessage", peerids)

	}

}
