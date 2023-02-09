package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func SendFileHandler() {

	go WriteData(fileChan, "/file/1.1.0")

}
func sendFile(path string) {
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
	fileChan <- []byte(fileSize)
	//time.Sleep(1 * time.Second)
	fileChan <- []byte(fileName)

	// sendBuffer := make([]byte, 1024)
	// for {
	// 	_, err = file.Read(sendBuffer)
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	fileChan <- sendBuffer
	// }
	fmt.Println("File has been sent successfully!")
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
terminar la lectura de archivos y migrar las funciones de lectura/escritura 
a stream.Read y write si se puede

func ReceiveFilehandler(stream network.Stream) {

	fmt.Print("ReceiveFilehandler")
	fileSizeBuffer := make([]byte, 10)
	stream.Read(fileSizeBuffer)

	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))
	fileNameBuffer := make([]byte, 64)
	stream.Read(fileNameBuffer)

	fileName := strings.Trim(string(fileNameBuffer), ":")
	fmt.Printf("archivo %s de %d \n", fileName, fileSize)
	newFile, err := os.Create(fileName)
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
}

func receiveFile(fileName string, conn net.Conn) {
	fileSizeBuffer := make([]byte, 10)
	conn.Read(fileSizeBuffer)
	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))

	fileNameBuffer := make([]byte, 64)
	conn.Read(fileNameBuffer)
	fileName = strings.Trim(string(fileNameBuffer), ":")

	newFile, err := os.Create(fileName)
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
		n, err := conn.Read(receiveBuffer)
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
}
