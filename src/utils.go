package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

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
		fmt.Println(e)
	}
	writer.Flush()
	if e != nil {
		fmt.Println(e)
	}

}
