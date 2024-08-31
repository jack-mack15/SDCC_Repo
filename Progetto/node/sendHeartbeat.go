package node

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func SendHeartbeat(localAddr *net.TCPAddr, remoteAddr *net.TCPAddr, myId int, remoteId int) {

	conn, err := net.DialTCP("udp", localAddr, remoteAddr)
	if err != nil {
		fmt.Println("errore durante la connessione:", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(5))

	message := "heartbeat from: " + strconv.Itoa(myId)
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("errore durante l'invio del messaggio:", err)
		return
	}

	//risposta dal nodo contattato
	buffer := make([]byte, 128)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("errore durante la risposta: %v", err)
		conn.Close()
		time.Sleep(2 * time.Second)
	}

	reply := string(buffer[:n])
	fmt.Printf("risposta dal nodo: %s\n", reply)
}
