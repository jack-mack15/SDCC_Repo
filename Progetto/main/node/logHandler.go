package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var file *os.File

func initLogFile() {

	err := os.MkdirAll("/log", 0755)
	if err != nil {
		log.Fatalf("Errore nella creazione della directory: %v", err)
	}
	// Apri (o crea) il file di log
	path := "/log/node" + strconv.Itoa(GetMyId()) + ".log"
	fmt.Printf("path is %s\n", path)
	file, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Errore nell'aprire il file di log:", err)
		return
	}

	log.SetOutput(file)
	log.Printf("PEER %d, log file\n", GetMyId())
}
