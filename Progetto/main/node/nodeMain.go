package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {

	//SET UP del nodo
	err := readConfigFile()
	if err == 0 {
		fmt.Println("errore nel recupero del file di conf")
		return
	}

	//"istanzio" un gossiper in base al file di config
	InitGossiper()

	//inizializzo le mie coordinate
	initMyCoordination()

	//recupero il mio indirizzo ip
	conn, err2 := net.Dial("udp", "8.8.8.8:80")
	if err2 != nil {
		panic(err2)
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	err3 := conn.Close()
	if err3 != nil {
		return
	}

	ownIP := localAddr.IP
	setOwnUDPAddr(&net.UDPAddr{IP: ownIP, Port: getMyPort()})
	setOwnTCPAddr(&net.TCPAddr{IP: ownIP, Port: getMyPort()})

	tryContactRegistry()

	initLogFile()

	//sleep per dare tempo a netem di assegnare i ritardi o tirare su la rete
	time.Sleep(5 * time.Second)

	//avvio della goroutine di ricezione
	go receiverHandler()

	//FASE ATTIVA
	for {
		//scelgo i nodi da contattare
		nodesToContact := getNodesToContact()

		contactNode(nodesToContact)

		time.Sleep(time.Duration(getHBDelay()) * time.Millisecond)

		printAllNodeList()

		activeNodeLenght := getLenght()
		if activeNodeLenght == 0 {
			tryLazzarus()
		}

		//TODO se devo cambiare il tutto devo fare il seguente
		//TODO contact node fa anti entropy con 1 o più nodi
		//TODO send heartbeat deve inviare le proprie coordinate, error e un tot di rtt di altri nodi
		//TODO handleUDP deve avviare gli aggiornamenti delle coordinate

		//TODO controllare tutta la robba da eliminare

		//TODO sistemazione del codice

	}
}

// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	conn, err := net.ListenUDP("udp", getOwnUDPAddr())
	if err != nil {
		fmt.Println("receiverHandler()--> errore creazione listener UDP:", err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Errore nella chiusura della connessione: %v", err)
		}
	}()

	for {
		buffer := make([]byte, 128)

		n, remoteUDPAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("receiverHandler()--> errore lettura pacchetto:", err)
			continue
		}

		go handleUDPMessage(conn, remoteUDPAddr, buffer[:n])

	}
}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(selectedNodes []node) {

	//contatto i nodi
	lenght := len(selectedNodes)

	var wg sync.WaitGroup

	for i := 0; i < lenght; i++ {
		wg.Add(1)
		go sendHeartbeat(selectedNodes[i], getMyId(), &wg)
	}
	wg.Wait()

}
