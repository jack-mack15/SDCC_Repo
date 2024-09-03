package node

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func ContactRegistry(localAddr *net.TCPAddr, addr string) string {

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

	defer conn.Close()

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

	fmt.Print("risposta server: ", reply)

	return reply
}
