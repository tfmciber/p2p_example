package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a host:port string")
		return
	}

	CONNECT := arguments[1]
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		if text[0:4] == "send" {
			file, err := os.Open(text[5 : len(text)-1])
			if err != nil {
				fmt.Println(err)
				continue
			}
			sendFile(file, c)
			continue
		}
		if text[0:8] == "receive" {
			receiveFile(text[9:len(text)-1], c)
			continue
		}
		fmt.Fprintf(c, text+"\n")
	}
}

func sendFile(file *os.File, conn net.Conn) {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(fmt.Sprintf("%d", fileInfo.Size()), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Fprint(conn, fileSize)
	fmt.Fprint(conn, fileName)

	sendBuffer := make([]byte, 1024)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		conn.Write(sendBuffer)
	}
	fmt.Println("File has been sent successfully!")
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
