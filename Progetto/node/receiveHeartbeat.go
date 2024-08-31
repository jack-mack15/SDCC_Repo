package node

import (
	"fmt"
	"net"
)

func listenForHeartbeats(port string) {
	addr, err := net.ResolveUDPAddr("udp", ":"+port)
	if err != nil {
		fmt.Println("errore durante la risoluzione dell'indirizzo UDP:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("errore durante l'ascolto:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("errore durante la ricezione del messaggio:", err)
			continue
		}

		fmt.Printf("heartbeat ricevuto da %s: %s\n", addr, string(buffer[:n]))
	}
}
