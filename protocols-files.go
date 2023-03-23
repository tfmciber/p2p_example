package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/schollz/progressbar/v3"
)

func (c *P2Papp) sendFile(rendezvous string, path string) {

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
	fromrendezvous := fillString(rendezvous, 64)

	c.writeDataRendChan(c.fileproto, rendezvous, func(stream network.Stream) {

		stream.Write([]byte(fromrendezvous))
		stream.Write([]byte(fileSize))
		stream.Write([]byte(fileName))
		for {

			sendBuffer := make([]byte, 1024)

			_, err := file.Read(sendBuffer)

			if err == io.EOF {

				break
			} else {

				stream.Write([]byte(sendBuffer))
			}
		}

	})

	file.Close()
	fmt.Println("\t File has been sent successfully!")
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

// falta añadir desde que rendezvous se ha recibido
func (c *P2Papp) receiveFilehandler(stream network.Stream) {

	fmt.Println("receiveFilehandler")
	downloadDir := getDefaultDownloadDir()
	createDirIfNotExist(downloadDir) //create download dir if it does not exist

	fromrendezvousbuffer := make([]byte, 64)
	stream.Read(fromrendezvousbuffer)
	fromrendezvous := strings.Trim(string(fromrendezvousbuffer), ":")

	fmt.Println(fromrendezvous)
	fileSizeBuffer := make([]byte, 10)
	stream.Read(fileSizeBuffer)
	fmt.Println(fileSizeBuffer)

	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))

	fileNameBuffer := make([]byte, 64)
	stream.Read(fileNameBuffer)
	fileName := strings.Trim(string(fileNameBuffer), ":")
	fmt.Println(fileName)

	fmt.Println("Receiving file: ", fileName, " of size: ", fileSize, " bytes from rendezvous ", fromrendezvous, " from peer ", stream.Conn().RemotePeer())

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
