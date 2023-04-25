package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	"os"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (c *P2Papp) QueueFile(rendezvous string, path string) {
	c.fmtPrintln("[*] QueueFile", rendezvous, path)

	if rendezvous == "" || path == "" {
		return
	}

	if !c.checkRend(rendezvous) {
		c.fmtPrintln("rendezvous not found: ", rendezvous)
		return
	}
	c.queueFilesMutex.Lock()
	c.fmtPrintln("pass lock")

	if _, ok := c.queueFiles[rendezvous]; !ok {
		c.queueFiles[rendezvous] = make([]string, 0)
	}
	c.fmtPrintln("queueFiles: ", c.queueFiles)
	//check if file is already in queue

	c.queueFiles[rendezvous] = append(c.queueFiles[rendezvous], path)
	c.fmtPrintln("queueFiles: ", c.queueFiles)
	c.queueFilesMutex.Unlock()
	c.movequeue <- rendezvous

}
func (c *P2Papp) MoveQueue() {
	c.fmtPrintln("[*] MoveQueue")
	go func() {

		for {
			select {
			case rendezvous := <-c.movequeue:
				c.fmtPrintln("[*] c.movequeue channel triggered")

				c.queueFilesMutex.Lock()

				files := c.queueFiles[rendezvous]
				c.fmtPrintln("rendezvous: ", rendezvous, " files: ", files, " len: ", len(files))
				if len(files) > 0 {
					c.SendFile(rendezvous, files[0])
					if len(files) > 1 {

						c.queueFiles[rendezvous] = files[1:]
						c.queueFilesMutex.Unlock()
						c.movequeue <- rendezvous
					} else {
						c.queueFiles[rendezvous] = make([]string, 0)
						c.queueFilesMutex.Unlock()
					}

				}

			case <-c.ctx.Done():
				return

			}
		}
	}()
}

func (c *P2Papp) SendFile(rendezvous string, path string) {
	c.fmtPrintln("[*] SendFile", rendezvous, path)

	x, _ := c.Get(rendezvous)
	c.fmtPrintln("x: ", x)
	if x != nil {

		file, err := os.Open(path)
		// calculate file hash
		hashString := c.getHash(file)
		if err != nil {
			c.fmtPrintln(err)
			return
		}

		fileInfo, err := file.Stat()
		if err != nil {
			log.Println(err)
			return
		}

		fileSize := fillString(fmt.Sprintf("%d", fileInfo.Size()), 10)
		fileName := fillString(fmt.Sprintf("%s", fileInfo.Name()), 64)
		c.fmtPrintln("sending file ", fileName, " size: ", fileSize, " to ", rendezvous)

		fromrendezvous := fillString(rendezvous, 64)

		totalsent := 0
		progress := 0
		last := 0
		c.EmitEvent("progressFile", rendezvous, "me", 0, fileInfo.Name())
		c.writeDataRendFunc(c.fileproto, rendezvous, func(stream network.Stream) {
			if stream == nil {
				return
			}
			n, err := stream.Write([]byte(fromrendezvous))

			c.fmtPrintln("write fromrendezvous: ", n, err)
			n, err = stream.Write([]byte(fileSize))
			c.fmtPrintln("write fileSize: ", n, err)
			n, err = stream.Write([]byte(fileName))
			c.fmtPrintln("write fileName: ", n, err)
			n, err = stream.Write([]byte(hashString))
			c.fmtPrintln("write hash: ", n, err)
			for {

				sendBuffer := make([]byte, 1024)

				n, err := file.Read(sendBuffer)

				if err == io.EOF {
					break
				} else {
					tries := 0
				retry:
					if stream == nil {
						return
					}
					n_write, err := stream.Write([]byte(sendBuffer)[:n])

					if err == nil {
						totalsent += n_write
						aux := (float64(totalsent)) / (float64(fileInfo.Size()))
						progress = int(aux * 100)
						if progress%7 == 0 && progress != last || progress == 100 {
							c.EmitEvent("progressFile", rendezvous, "me", progress, fileInfo.Name())

						}
						last = progress
					} else {
						c.fmtPrintln("restart stream with ", stream.Conn().RemotePeer().String(), " err: ", err)
						peerid := stream.Conn().RemotePeer()

						stream = c.streamStart(peerid, c.fileproto)

						tries++
						if tries > 3 {
							break
						}

						goto retry

					}

				}
			}
			file.Close()
			ret := 100
			if totalsent != int(fileInfo.Size()) {
				ret = -1
			}
			c.fmtPrintln("totalsent: ", totalsent, " fileSize: ", fileInfo.Size(), "ret", ret)
			stream.Close()
			c.EmitEvent("progressFile", rendezvous, "me", ret, fileInfo.Name())

		})
	}
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

	downloadDir := getDefaultDownloadDir()
	createDirIfNotExist(downloadDir) //create download dir if it does not exist

	fromrendezvousbuffer := make([]byte, 64)
	n, err := stream.Read(fromrendezvousbuffer)

	fromrendezvous := strings.Trim(string(fromrendezvousbuffer), ":")
	c.fmtPrintln("fromrendezvous: ", fromrendezvous, " err: ", err, " n: ", n)

	fileSizeBuffer := make([]byte, 10)
	n, err = stream.Read(fileSizeBuffer)
	c.fmtPrintln("fileSizeBuffer: ", fileSizeBuffer, " err: ", err, " n: ", n)

	fileSize, _ := strconv.Atoi(strings.Trim(string(fileSizeBuffer), ":"))

	fileNameBuffer := make([]byte, 64)
	n, err = stream.Read(fileNameBuffer)
	hashoffile := make([]byte, 64)
	n, err = stream.Read(hashoffile)
	fileName := strings.Trim(string(fileNameBuffer), ":")
	c.fmtPrintln("fileName: ", fileName, " err: ", err, " n: ", n)
	temporaryfilename := fmt.Sprintf("Uncompleted_%s.tmp", fileName)

	filepath := fmt.Sprintf("%s/%s", downloadDir, temporaryfilename)

	c.fmtPrintln("Receiving file: ", fileName, " of size: ", fileSize, " bytes from rendezvous ", fromrendezvous, " from peer ", stream.Conn().RemotePeer(), "with hash:", string(hashoffile))

	var pa struct {
		Path     string `json:"path"`
		Filename string `json:"filename"`
	}
	pa.Path = filepath
	pa.Filename = fileName
	var newFile *os.File
	runtime.EventsEmit(c.ctx, "receiveFile", fromrendezvous, stream.Conn().RemotePeer().String(), pa)
	newFile, err = os.Create(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer newFile.Close()

	var receivedBytes int

	last := 0
	receiveBuffer := make([]byte, 1024)
	for {
		if (fileSize - receivedBytes) < 1024 {
			receiveBuffer = make([]byte, (fileSize - receivedBytes))
		}
		n, err := stream.Read(receiveBuffer)
		if err != nil {
			if err != io.EOF {
				c.fmtPrintln("error downloading file: ", err)
				c.fmtPrintln(receiveBuffer)
				break
			}
		}
		receivedBytes += n
		newFile.Write(receiveBuffer[:n])
		if receivedBytes == fileSize {
			break
		}
		aux := (float64(receivedBytes)) / (float64(fileSize))
		progress := int(aux * 100)
		if progress%7 == 0 && progress != last || progress == 100 {
			c.EmitEvent("progressFile", fromrendezvous, stream.Conn().RemotePeer().String(), progress, fileName)
		}
		last = progress

	}

	stream.Close()

	c.fmtPrintln("Received file: ", fileName, " of size: ", fileSize, " bytes from rendezvous ", fromrendezvous, " from peer ", stream.Conn().RemotePeer(), " receivedBytes: ", receivedBytes)
	ret := 100
	//get hash of file
	hashString := c.getHash(newFile)
	newFile.Close()
	c.fmtPrintln("hashString: ", hashString, " hashoffile: ", string(hashoffile))
	if hashString != string(hashoffile) {
		ret = -1
	} else {

		err = os.Rename(filepath, fmt.Sprintf("%s/%s", downloadDir, fileName))

	}
	c.EmitEvent("progressFile", fromrendezvous, stream.Conn().RemotePeer().String(), ret, fileName)

	return
}

func (c *P2Papp) getHash(file *os.File) string {
	file.Seek(0, 0)
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	//set file pointer to the beginning
	file.Seek(0, 0)
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)
	return hashString

}
