package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/schollz/progressbar/v3"
)

func sendFile(rendezvous string, path string) {

	log.Println("sendFile ", path)
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	fileSize := fillString(fmt.Sprintf("%d", fileInfo.Size()), 10)
	fileName := fillString(fmt.Sprintf("%s", fileInfo.Name()), 64)

	writeDataRend([]byte(fileSize), string(fileproto), rendezvous, false)
	writeDataRend([]byte(fileName), string(fileproto), rendezvous, false)

	start := time.Now()
	bar := progressbar.Default(100)
	totalSend := 0
	last := 0
	for {

		sendBuffer := make([]byte, 1024)

		read, err := file.Read(sendBuffer)

		totalSend += read
		if err == io.EOF {

			break
		} else {

			writeDataRend(sendBuffer, string(fileproto), rendezvous, false)
		}
		//progress bar indicating download progress aproximately every 10 % of the file
		aux := (float64(totalSend)) / (float64(fileInfo.Size()))
		progress := int(aux * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}
	elapsed := time.Since(start)
	file.Close()
	log.Println("\r File has been sent successfully! in ", elapsed, "at an average speed of ", float64(fileInfo.Size())/float64(elapsed.Milliseconds()), " Mbytes/sec")
}

func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

func receiveFilehandler(stream network.Stream) {

	downloadDir := getDefaultDownloadDir()
	createDirIfNotExist(downloadDir) //create download dir if it does not exist

	fileSizeBuffer := make([]byte, 10)
	stream.Read(fileSizeBuffer)

	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))

	fileNameBuffer := make([]byte, 64)
	stream.Read(fileNameBuffer)
	fileName := strings.Trim(string(fileNameBuffer), ":")

	log.Println("Receiving file: ", fileName, " of size: ", fileSize, " bytes")

	newFile, err := os.Create(fmt.Sprintf("%s/%s", downloadDir, fileName))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newFile.Close()

	var receivedBytes int
	bar := progressbar.Default(100)
	last := 0
	receiveBuffer := make([]byte, 1024)
	for {
		if (fileSize - receivedBytes) < 1024 {
			receiveBuffer = make([]byte, (fileSize - receivedBytes))
		}
		n, err := stream.Read(receiveBuffer)
		if err != nil {
			break
		}
		receivedBytes += n
		newFile.Write(receiveBuffer[:n])
		if receivedBytes == fileSize {
			break
		}
		aux := (float64(receivedBytes)) / (float64(fileSize))
		progress := int(aux * 100)
		if progress%1 == 0 && progress != last {
			bar.Add(progress - last)
		}
		last = progress

	}

	log.Println("File has been received successfully!")
	stream.Close()
	newFile.Close()
}
