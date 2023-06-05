package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	filepath "path/filepath"
	"sync"
	"time"

	"os"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p/core/network"
)

func (c *P2Papp) QueueFile(rendezvous string, path string) {
	c.fmtPrintln("[*] QueueFile", rendezvous, path)

	if rendezvous == "" || path == "" {
		return
	}

	if !c.checkRend(rendezvous) && !c.checkUser(rendezvous) {
		c.fmtPrintln("rendezvous or user not found: ", rendezvous)
		return
	}
	c.queueFilesMutex.Lock()

	c.queueFiles[rendezvous] = append(c.queueFiles[rendezvous], path)
	c.queueFilesMutex.Unlock()
	c.MoveQueue(rendezvous)

}
func (c *P2Papp) MoveQueue(rendezvous string) {

	c.queueFilesMutex.Lock()
	defer c.queueFilesMutex.Unlock()

	if len(c.queueFiles[rendezvous]) > 0 {
		nexfile := c.queueFiles[rendezvous][0]
		c.queueFiles[rendezvous] = c.queueFiles[rendezvous][1:]
		c.SendFile(rendezvous, nexfile)

	}

}

func (c *P2Papp) SendFile(rendezvous string, path string) {
	c.fmtPrintln("[*] SendFile", rendezvous, path)

	x, _ := c.Get(rendezvous, true)

	file, err := os.Open(path)
	// calculate file hash

	fileInfo, err := file.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	totalprogress := -1
	c.fmtPrintln("x: ", x, "len", len(x))
	if x != nil {
		hashString := c.getHash(file)
		if err != nil {
			c.fmtPrintln(err)
			return
		}

		fileSize := fillString(fmt.Sprintf("%d", fileInfo.Size()), 10)
		fileName := fillString(fmt.Sprintf("%s", fileInfo.Name()), 64)
		c.fmtPrintln("sending file ", fileName, " size: ", fileSize, " to ", rendezvous)

		fromrendezvous := fillString(rendezvous, 64)

		totalsent := 0

		last := 0

		c.EmitEvent("progressFile", rendezvous, "me", 0, fileInfo.Name())

		c.writeFileRendFunc(c.fileproto, rendezvous, func(streams []network.Stream) {

			progress := 0
			var wg sync.WaitGroup
			for _, stream := range streams {
				wg.Add(1)

				go func(stream network.Stream) {
					defer wg.Done()
					n, err := stream.Write([]byte(fromrendezvous))

					c.fmtPrintln("write fromrendezvous: ", n, err)
					n, err = stream.Write([]byte(fileSize))
					c.fmtPrintln("write fileSize: ", n, err)
					n, err = stream.Write([]byte(fileName))
					c.fmtPrintln("write fileName: ", n, err)
					n, err = stream.Write([]byte(hashString))
					c.fmtPrintln("write hash: ", n, err)
				}(stream)
			}

			wg.Wait()
			c.fmtPrintln("sent header")
			for {

				sendBuffer := make([]byte, 1024)

				n, err := file.Read(sendBuffer)

				if err == io.EOF {

					break
				} else if err == nil {
					var wg2 sync.WaitGroup
					for _, stream := range streams {
						wg2.Add(1)
						go func(stream network.Stream) {
							defer wg2.Done()
							tries := 0
						retry:
							if stream == nil {
								return
							}
							n_write, err := stream.Write([]byte(sendBuffer)[:n])

							if err == nil {
								c.filemu.Lock()
								totalsent += n_write
								aux := (float64(totalsent)) / (float64(fileInfo.Size()))
								progress = int(aux*100) / len(x)

								if progress%7 == 0 && progress != last && progress != 100 {
									c.EmitEvent("progressFile", rendezvous, "me", progress, fileInfo.Name())

								}
								last = progress
								c.filemu.Unlock()
							} else {
								c.fmtPrintln("restart stream with ", stream.Conn().RemotePeer().String(), " err: ", err)
								peerid := stream.Conn().RemotePeer()

								stream = c.streamStart(peerid, c.fileproto)

								tries++
								if tries > 3 {
									return
								}

								goto retry

							}
						}(stream)
					}
					wg2.Wait()

				} else {
					c.fmtPrintln("error reading file: ", err)

					break
				}
			}
			totalprogress += progress

		})
		if totalsent != len(x)*int(fileInfo.Size()) {
			totalprogress = -1
		} else {
			totalprogress = 100

		}
	}
	file.Close()
	c.EmitEvent("progressFile", rendezvous, "me", totalprogress, fileInfo.Name())
	t := time.Now()
	date := t.Format("02/01/2006 15:04")
	var pa PathFilename
	ret := 100
	if totalprogress != 100 {
		ret = -1
	}
	pa.Path = path
	pa.Filename = fileInfo.Name()
	pa.Progress = ret

	mess := Message{Text: "", Date: date, Src: c.Host.ID().String(), Pa: pa, Status: ret}
	c.saveMessages(map[string]Message{rendezvous: mess})

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
	t := time.Now()
	date := t.Format("02/01/2006 15:04:30")
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

	filePath := filepath.Join(downloadDir, fmt.Sprintf("Uncompleted_%s%d.tmp", fileName, rand.Intn(100000)))

	c.fmtPrintln("Receiving file: ", fileName, " of size: ", fileSize, " bytes from rendezvous ", fromrendezvous, " from peer ", stream.Conn().RemotePeer(), "with hash:", string(hashoffile))

	var pa PathFilename
	pa.Path = filePath
	pa.Filename = fileName
	pa.Progress = -2 // queued

	var newFile *os.File
	if fromrendezvous == c.Host.ID().String() {
		c.EmitEvent("receiveFile", stream.Conn().RemotePeer().String(), stream.Conn().RemotePeer().String(), date, pa)
	} else {
		c.EmitEvent("receiveFile", fromrendezvous, stream.Conn().RemotePeer().String(), date, pa)
	}
	newFile, err = os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

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
		if progress%7 == 0 && progress != last && progress != 100 {

			if fromrendezvous == c.Host.ID().String() {

				c.EmitEvent("progressFile", stream.Conn().RemotePeer().String(), stream.Conn().RemotePeer().String(), progress, fileName)
			} else {
				c.EmitEvent("progressFile", fromrendezvous, stream.Conn().RemotePeer().String(), progress, fileName)
			}

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
		c.fmtPrintln("Received correctly")

		// check if a file with the same name exists

		ext := filepath.Ext(fileName)
		name := strings.TrimSuffix(fileName, ext)

		newFilePath := filepath.Join(downloadDir, fileName)

		if _, err := os.Stat(newFilePath); err == nil {
			i := 1
			for {
				newFileName := name + " (" + strconv.Itoa(i) + ")" + ext // cambia el nombre del archivo agregando el número entre paréntesis antes de la extensión
				c.fmtPrintln("checking if file exists: ", newFileName, " in ", downloadDir, " ...")
				newFilePath = filepath.Join(downloadDir, newFileName)
				_, err := os.Stat(newFilePath)

				if err != nil { // si el nuevo nombre de archivo no existe

					err = os.Rename(filePath, newFilePath)
					if err == nil {
						pa.Path = newFilePath
					}
					break
				}

				i++
			}
		} else {
			c.fmtPrintln("newFilePath: ", newFilePath)
			err = os.Rename(filePath, newFilePath)
			c.fmtPrintln("err: ", err)
			if err == nil {
				pa.Path = newFilePath
			}

		}

	}
	if fromrendezvous == c.Host.ID().String() {

		c.EmitEvent("progressFile", stream.Conn().RemotePeer().String(), stream.Conn().RemotePeer().String(), ret, fileName, pa.Path)
	} else {
		c.EmitEvent("progressFile", fromrendezvous, stream.Conn().RemotePeer().String(), ret, fileName, pa.Path)
	}
	t = time.Now()
	date = t.Format("02/01/2006 15:04:30")
	pa.Progress = ret
	mess := Message{Text: "", Date: date, Src: stream.Conn().RemotePeer().String(), Pa: pa, Status: ret}
	c.saveMessages(map[string]Message{fromrendezvous: mess})

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
