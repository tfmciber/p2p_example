package main

import (
	"fmt"
	filepath "path/filepath"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (c *P2Papp) EmitEvent(event string, data ...interface{}) {

	c.fmtPrintln("emit event: ", event, " data ", data)
	runtime.EventsEmit(c.ctx, event, data...)

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
		pathFilenames = append(pathFilenames, PathFilename{Path: f, Filename: filename, Progress: -2})

		i += 1
	}

	return pathFilenames
}

func (c *P2Papp) DataChanged() {

	go func() {
		for {
			select {
			case <-c.chatadded:
				c.fmtPrintln("chat added")
				runtime.EventsEmit(c.ctx, "updateChats", c.GetData())
			case <-c.useradded:

				runtime.EventsEmit(c.ctx, "updateUsers", c.ListUsers())

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
	timemow := time.Now().Format("2006-01-02 15:04:05")
	output = fmt.Sprintf("[%s] - %s", timemow, output)

	fmt.Println(output)
	runtime.EventsEmit(c.ctx, "receiveCommands", output)
}

func (c *P2Papp) AddDm(peerid peer.ID) {

	if _, ok := c.direcmessages[peerid.String()]; !ok {

		c.direcmessages[peerid.String()] = DmData{Status: true}
		//convert []peer.ID to []string

		c.saveData("direcmessages", c.direcmessages)
		runtime.EventsEmit(c.ctx, "directMessage", c.direcmessages)

	}

}
