package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
)

func sendFile(rendezvous string, path string) {

	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(fmt.Sprintf("%d", fileInfo.Size()), 10)

	fileName := fillString(fmt.Sprintf("%s", fileInfo.Name()), 64)
	WriteDataRend([]byte(fileSize), "/file/1.1.0", rendezvous, false)
	WriteDataRend([]byte(fileName), "/file/1.1.0", rendezvous, false)
	start := time.Now()
	for {
		sendBuffer := make([]byte, 1024)
		_, err := file.Read(sendBuffer)
		if err == io.EOF {
			break
		} else {

			WriteDataRend(sendBuffer, "/file/1.1.0", rendezvous, false)
		}

	}
	elapsed := time.Since(start)
	fmt.Println("File has been sent successfully! in ", elapsed, "at an average speed of ", float64(fileInfo.Size())/float64(elapsed.Milliseconds()), " Mbytes/sec")
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

func ReceiveFilehandler(stream network.Stream) {

	downloadDir := GetDefaultDownloadDir()
	createDirIfNotExist(downloadDir) //create download dir if it does not exist

	fileSizeBuffer := make([]byte, 10)
	stream.Read(fileSizeBuffer)

	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))

	fileNameBuffer := make([]byte, 64)
	stream.Read(fileNameBuffer)
	fileName := strings.Trim(string(fileNameBuffer), ":")

	fmt.Println("Receiving file: ", fileName, " of size: ", fileSize, " bytes")

	newFile, err := os.Create(fmt.Sprintf("%s/%s", downloadDir, fileName))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newFile.Close()

	var receivedBytes int

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
	}

	fmt.Println("File has been received successfully!")
	stream.Close()

}
