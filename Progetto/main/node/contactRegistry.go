package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func contactRegistry(localAddr *net.TCPAddr, addr string) string {

	//ottengo l'indirizzo del service registry
	remoteAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Printf("errore ottenimento indirizzo di %s: %v", addr, err)
	}

	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	if err != nil {
		fmt.Println("errore connessione:", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Errore nella chiusura della connessione: %v", err)
		}
	}()

	message := "0#hello registry\n"

	// Invio del messaggio al server
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("errore invio messaggio:", err.Error())
		return ""
	}

	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("erore ricezione risp:", err.Error())
		return ""
	}

	fmt.Printf("risposta server: %s\n\n", reply)

	return reply
}
