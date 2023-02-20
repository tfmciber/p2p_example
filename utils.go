package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/libp2p/go-libp2p/core/peer"
)

//func to get the default download directory os independent
func GetDefaultDownloadDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE") // windows
	}
	downloadir := home + "/Downloads"
	return downloadir
}

//func to create a directory if it does not exist os independent
func createDirIfNotExist(dir string) {

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir, 0777)
	}

}

//funct to create a csv file
func createFile(file string) {
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
}

//func to append string into csv
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
	writer.Flush()

	if e != nil {
		fmt.Println(e)
	}

}

//func to add to a map of slices of peers if not already there else set peer to online
func Add(Map map[string][]Peer, Peeraddr peer.AddrInfo, rendezvous string) map[string][]Peer {

	found := false
	var addrs []peer.AddrInfo
	for i, v := range Map[rendezvous] {
		for _, v2 := range v.peer {
			if v2.ID == Peeraddr.ID {
				addrs = append(addrs, v2)

				found = true

			}
		}
		Map[rendezvous][i] = Peer{addrs, true}

	}

	if found == false {
		addrs = append(addrs, Peeraddr)
		Map[rendezvous] = append(Map[rendezvous], Peer{addrs, true})

	}

	return Map

}

//func to check if slice contains a value
func Contains(s []peer.ID, e peer.ID) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
