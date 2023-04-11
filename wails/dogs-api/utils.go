package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/libp2p/go-libp2p/core/peer"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func contains(s []peer.ID, e peer.ID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func PrintToVariable(args ...interface{}) (string, error) {
	var buffer bytes.Buffer

	n, err := fmt.Fprintln(&buffer, args...)
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", nil
	}

	return buffer.String()[:buffer.Len()-1], nil
}
func (c *P2Papp) fmtPrintln(args ...interface{}) {
	output, err := PrintToVariable(args...)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
	wailsruntime.EventsEmit(c.ctx, "receiveCommands", output)
}
func (c *P2Papp) OpenFileExplorer(path string) error {
	c.fmtPrintln("open file explorer")
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Run()
}

// func to get the default download directory os independent
func getDefaultDownloadDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // windows
	}
	downloadir := home + "/Downloads"
	return downloadir
}

// func to create a directory if it does not exist os independent
func createDirIfNotExist(dir string) {

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0777)
	}

}

// funct to create a csv file
func createFile(file string) {
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

// func to append string into csv
func appendToCSV(file string, data []string) {

	//if file not exist create it
	if _, err := os.Stat(file); os.IsNotExist(err) {
		createFile(file)
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	writer := csv.NewWriter(f)
	e := writer.Write(data)
	if e != nil {

	}
	writer.Flush()
	if e != nil {

	}

}
