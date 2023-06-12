package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type Message struct {
	Text   string       `json:"text"`
	Date   string       `json:"date"`
	Src    string       `json:"src"`
	Pa     PathFilename `json:"pa"`
	Status int          `json:"status"`
}
type Config struct {
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	PreferQUIC bool   `json:"preferQUIC"`
	Refresh    int    `json:"refresh"`
}

func (c *P2Papp) SendDM(aux string) {

	peerid, err := peer.Decode(aux)
	if err == nil {
		//check if peerid in map of peers
		if _, ok := c.direcmessages[peerid.String()]; !ok {
			c.direcmessages[peerid.String()] = DmData{Status: true}
			c.EmitEvent("directMessage", c.direcmessages)
		}

	}

}

func (c *P2Papp) SendTextHandler(text string, rendezvous string) int {
	c.fmtPrintln("SendTextHandler " + text + " " + rendezvous)
	////get time and date dd/mm/yyyy hh:mm
	date := time.Now().Format("02/01/2006 15:04:30")
	ok := true
	message := (rendezvous + "$" + text + "$" + date)

	x, y := c.Get(rendezvous, true)

	mess := Message{Text: text, Date: date, Src: c.Host.ID().String()}
	if y == true {
		//we are sending a direct message
		c.AddDm(x[0])
	} else if x == nil || len(x) == 0 {
		mess.Status = -1
		c.saveMessages(map[string]Message{rendezvous: mess})
		return -1
	}

	c.writeDataRendFunc(c.textproto, rendezvous, func(stream network.Stream) {

		fmt.Println("sending data to: ", stream.Conn().RemotePeer().String())

		n, err := stream.Write([]byte(message))
		c.fmtPrintln(fmt.Sprintf("Sent [*] %s [%s] %s = %s,%d \n", date, rendezvous, c.Host.ID(), text, n), "err:", err)

		if err != nil {
			if err != io.EOF {
				ok = false
				c.fmtPrintln("SendTextHandler error", err)
				//c.disconnectHost(stream, err, string(stream.Protocol()))
			}
		}

	})

	intstatus := 1
	if !ok {
		intstatus = -1
	}
	mess.Status = intstatus
	c.saveMessages(map[string]Message{rendezvous: mess})
	return 1

}
func (c *P2Papp) LeaveChat(rendezvous string) {
	c.fmtPrintln("[*] LeaveChat " + rendezvous)
	//check if rendezvus exists
	if rendezvous == "" {
		c.fmtPrintln("rendezvous is empty")
		c.chatadded <- rendezvous
		c.useradded <- true
		return
	}
	peerid, err := peer.Decode(rendezvous)
	fmt.Println("peer.Decode(rendezvous)", peerid, err, rendezvous)
	if err == nil {

		c.fmtPrintln("Leave DM")
		c.leaveDm(peerid)
		c.EmitEvent("directMessage", c.direcmessages)
	} else if c.checkRend(rendezvous) == false {

		c.fmtPrintln("rendezvous does not exist or is not a peerid")
		c.chatadded <- rendezvous
		c.useradded <- true
		return
	} else {
		c.fmtPrintln("Leave chat")
		c.CancelRendezvous(rendezvous)
		c.fmtPrintln("rendezvous deleted "+rendezvous, "c.data:", c.data)
		c.leaveChat(rendezvous)
	}

	c.writeDataRendFunc(c.cmdproto, rendezvous, func(stream network.Stream) {
		tries := 0
	leave:
		n, err := stream.Write([]byte("leave$" + rendezvous))
		if err != nil {
			tries++
			peerid := stream.Conn().RemotePeer()
			c.fmtPrintln("leave sent to "+rendezvous+" "+c.Host.ID().String()+" "+fmt.Sprintf("%d", n)+" "+peerid.String(), err)
			stream.Close()

			stream = c.streamStart(peerid, c.cmdproto)
			if tries < 3 {

				goto leave
			}
		}
		fmt.Println(n, err)
		c.fmtPrintln("leave sent to " + rendezvous + " " + c.Host.ID().String() + " " + fmt.Sprintf("%d", n))

	})

	c.chatadded <- rendezvous
	c.useradded <- true

	c.fmtPrintln("LeaveChat " + rendezvous + " done")

}

func (c *P2Papp) leaveDm(peerid peer.ID) {

	if _, ok := c.direcmessages[peerid.String()]; ok {
		c.direcmessages[peerid.String()] = DmData{Status: false}

		runtime.EventsEmit(c.ctx, "directMessage", c.direcmessages)
	}

}
func (c *P2Papp) leaveChat(rendezvous string) {
	c.fmtPrintln("LeaveChat " + rendezvous)

	c.data[rendezvous] = HostData{Peers: c.data[rendezvous].Peers, Timer: c.data[rendezvous].Timer, Status: false}
	c.chatadded <- rendezvous

}
func (c *P2Papp) DeleteChat(rendezvous string) {
	c.fmtPrintln("DeleteChat " + rendezvous)

	peerid, err := peer.Decode(rendezvous)
	fmt.Println("peer.Decode(rendezvous)", peerid, err, rendezvous)
	if err == nil {
		delete(c.direcmessages, peerid.String())
		c.EmitEvent("directMessage", c.direcmessages)
	} else {
		c.CancelRendezvous(rendezvous)
		delete(c.data, rendezvous)
		c.chatadded <- rendezvous
	}

}

func (c *P2Papp) receiveTexthandler(stream network.Stream) {

	for {
		buff := make([]byte, 2000)
		_, err := stream.Read(buff)

		//if err is not EOF
		if err != nil {
			if err != nil {
				if err != io.EOF {
					c.fmtPrintln("SendTextHandler error", err)
					//c.disconnectHost(stream, err, string(stream.Protocol()))

				}
			}

		}

		data := strings.SplitN(string(buff[:]), "$", 3)
		var rendezvous string
		var text string
		var date string
		if len(data) > 1 {
			rendezvous = data[0]
			text = data[1]

		}
		if len(data) > 2 {
			date = data[2]
		} else {
			t := time.Now()
			date = t.Format("02/01/2006 15:04:30")
		}
		if rendezvous == c.Host.ID().String() {
			// if we receive our ID as rendezvous, it means we are receiving a direct message
			c.AddDm(stream.Conn().RemotePeer())
			rendezvous = stream.Conn().RemotePeer().String()
		}

		c.fmtPrintln(fmt.Sprintf("received message [*] %s [%s] %s = %s \n", date, rendezvous, stream.Conn().RemotePeer(), text))
		c.EmitEvent("receiveMessage", rendezvous, text, stream.Conn().RemotePeer().String(), date)

		mess := Message{Text: text, Date: date, Src: stream.Conn().RemotePeer().String()}

		c.saveMessages(map[string]Message{rendezvous: mess})

		stream.Close()
		return

	}

}

func (c *P2Papp) LoadData() {

	file := fmt.Sprintf("%s.data", c.Host.ID().String())
	if _, err := os.Stat(file); os.IsNotExist(err) {
		c.fmtPrintln("data file not found")
		return
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("ioutil.ReadFile(file)", err)
		return
	}

	var dat map[string]interface{}
	aux, err := c.decrypt(data, c.key)
	if err != nil {
		fmt.Println("c.decrypt(data, c.key)", err)
		return
	}

	if err := json.Unmarshal(aux, &dat); err != nil {

		fmt.Println("json.Unmarshal", err)
		return
	}
	if dat["data"] != nil {

		aux := dat["data"].(map[string]interface{})
		for k, v := range aux {
			if v == nil {
				continue
			}
			var peers []peer.ID
			if v.(map[string]interface{})["Peers"] != nil {

				for _, p := range v.(map[string]interface{})["Peers"].([]interface{}) {

					peerid, err := peer.Decode(p.(string))

					if err == nil {
						peers = append(peers, peerid)
					}

				}

			}

			c.data[k] = HostData{Peers: peers, Timer: uint(v.(map[string]interface{})["Timer"].(float64)), Status: v.(map[string]interface{})["Status"].(bool)}
		}

	}
	if dat["direcmessages"] != nil {

		aux := dat["direcmessages"].(map[string]interface{})
		for k, v := range aux {
			if v == nil {
				continue
			}
			c.direcmessages[k] = DmData{Status: v.(map[string]interface{})["Status"].(bool)}
		}

	}

	runtime.EventsEmit(c.ctx, "updateChats", c.GetData())
	runtime.EventsEmit(c.ctx, "updateUsers", c.ListUsers())

	if c.direcmessages == nil {
		c.direcmessages = make(map[string]DmData)
	}
	c.fmtPrintln("LoadData", c.direcmessages)
	runtime.EventsEmit(c.ctx, "directMessage", c.direcmessages)

	for rendezvous, item := range c.data {
		if rendezvous != "" {
			if item.Status {

				c.AddRendezvous(rendezvous)
				c.SetTimer(rendezvous, c.refresh)
			}
		}
	}

	if dat["message"] != nil {

		for chat, v := range dat["message"].(map[string]interface{}) {

			aux := v.([]interface{})
			for _, m := range aux {

				var textstr, datestr, srcstr string
				var path PathFilename
				var intstatus int

				text := m.(map[string]interface{})["text"]
				if text != nil {
					textstr = text.(string)
				}

				date := m.(map[string]interface{})["date"]
				if date != nil {
					datestr = date.(string)
				}

				src := m.(map[string]interface{})["src"]
				if src != nil {
					srcstr = src.(string)
				}

				pa := m.(map[string]interface{})["pa"]

				if pa != nil {
					//converto to PathFilename

					path.Filename = pa.(map[string]interface{})["filename"].(string)
					path.Path = pa.(map[string]interface{})["path"].(string)
					path.Progress = int(pa.(map[string]interface{})["progress"].(float64))

				}

				if srcstr == c.Host.ID().String() {
					srcstr = "me"
				}
				status := m.(map[string]interface{})["status"]
				if status != nil {
					intstatus = int(status.(float64))
				}

				if path.Path != "" {
					c.EmitEvent("loadMessages", chat, textstr, srcstr, datestr, []PathFilename{path}, intstatus)
				} else {
					c.EmitEvent("loadMessages", chat, textstr, srcstr, datestr, nil, intstatus)
				}
			}
		}
	}
	for peer, item := range c.direcmessages {

		if item.Status {

			c.AddRendezvous(peer)

		}
	}
}
func (c *P2Papp) saveMessages(message map[string]Message) {

	//join c.messages with message
	for k, v := range message {
		if c.messages[k] == nil {
			c.messages[k] = []Message{}
		}
		c.messages[k] = append(c.messages[k], v)
	}

}

func (c *P2Papp) saveData(typ string, data interface{}) { //type message, chats or Dms, thrashchats

	if data != nil {

		file := fmt.Sprintf("%s.data", c.Host.ID().String())

		err := c.updateJSONField(file, typ, data)
		if err != nil {
			fmt.Println("error updating json field")
			fmt.Println(err)
		}
	} else {
		fmt.Println("data is nil", typ)
	}

}

func (c *P2Papp) updateJSONField(filename string, field string, value interface{}) error {
	// Open the file for reading, create if it doesn't exist
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("error opening file", err)
		return err
	}

	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}
	data := make(map[string]interface{})

	if len(dat) == 0 {
		fmt.Println("empty file")
		data[field] = value

	} else {

		plaintext, err := c.decrypt(dat, c.key)

		if err != nil {
			fmt.Println("Error decrypting file:", err)

		} else {

			err = json.Unmarshal(plaintext, &data)
			if err != nil {
				fmt.Println("error decoding file", err)
				data[field] = value
			}
		}

		//append the new value to the existing slice

		if data[field] == nil {
			data[field] = value
		} else {

			switch value.(type) {

			case map[string][]Message:
				fmt.Println("map[string][]Message", value)
				aux := map[string][]Message{}

				for chat, v := range data[field].(map[string]interface{}) {

					for _, m := range v.([]interface{}) {

						var textstr, datestr, srcstr string
						var path PathFilename
						var intstatus int

						text := m.(map[string]interface{})["text"]
						if text != nil {
							textstr = text.(string)
						}

						date := m.(map[string]interface{})["date"]
						if date != nil {
							datestr = date.(string)
						}

						src := m.(map[string]interface{})["src"]
						if src != nil {
							srcstr = src.(string)
						}

						pa := m.(map[string]interface{})["pa"]
						if pa != nil {
							//converto to PathFilename

							path.Filename = pa.(map[string]interface{})["filename"].(string)
							path.Path = pa.(map[string]interface{})["path"].(string)
							path.Progress = int(pa.(map[string]interface{})["progress"].(float64))

						}
						status := m.(map[string]interface{})["status"]
						if status != nil {
							intstatus = int(status.(float64))
							if intstatus != 1 {
								intstatus = -1
							}
						}

						aux[chat] = append(aux[chat], Message{textstr, datestr, srcstr, path, intstatus})
					}

				}
				for chat, v := range value.(map[string][]Message) {
					aux[chat] = append(aux[chat], v...)
				}

				data[field] = aux

			default:
				fmt.Println("default", value)

				data[field] = value
			}
		}
	}
	dataBytes, err := json.Marshal(data)

	cipheredData := c.encrypt(dataBytes, c.key)

	// Write the bytes to the file and remove other data
	file.Seek(0, 0)
	file.Truncate(0)

	_, err = file.WriteAt(cipheredData, 0)
	if err != nil {
		fmt.Println("error writing file", err)
		return err
	}

	return nil
}
