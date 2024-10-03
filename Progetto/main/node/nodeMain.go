package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"fmt"
	"log"
	"net"
	"strconv"
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
	time.Sleep(2 * time.Second)

	ownIP := localAddr.IP
	setOwnTCPAddr(&net.TCPAddr{IP: ownIP, Port: getMyPort() + getMyId()})

	tryContactRegistry()

	//initLogFile()

	//avvio della goroutine di ricezione
	go receiverHandler()

	//sleep per dare tempo a netem di assegnare i ritardi o tirare su la rete
	time.Sleep(8 * time.Second)

	counter := 0
	//FASE ATTIVA
	for {
		//scelgo i nodi da contattare
		nodesToContact := getNodesToContact()

		contactNode(nodesToContact)

		time.Sleep(time.Duration(getHBDelay()) * time.Millisecond)

		counter++
		if counter == 10 {
			printAllNodeList()
			printAllCoordinates()
			counter = 0
		}

		activeNodeLenght := getLenght()
		if activeNodeLenght == 0 {
			tryLazzarus()
		}

		//TODO ci sta qualche problema con il marshal e unmarshal
		//in sendVivaldiMessage() alla fine devo fare la robba per avviare vivaldi
		//handleUDP è a posto
		//devo aggiungere le funzioni per gestire i nodi coordinate degli altri nodi
		//in modo tale che poi vivaldi algorithm viene avviato tranquillamente
		//collegare tutte le liste di nodi, quelli delle coordinate, quelli di nodeClass e quelli di digest/blind struct
		//modificare l'invio dei messaggi di gossip

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

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(getMyPort()))
	if err != nil {
		fmt.Println("reciverHandler()--> errore durante l'ascolto:", err.Error())
		return
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Printf("errore nella chiusura della connessione: %v", err)
		}
	}()

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("receiverHandler()--> errore lettura pacchetto:", err)
			continue
		}
		go handleUDPHandler(conn)

	}

}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(selectedNodes []*node) {

	//contatto i nodi
	lenght := len(selectedNodes)

	var wg sync.WaitGroup

	for i := 0; i < lenght; i++ {
		wg.Add(1)

		go sendVivaldiMessage(selectedNodes[i], &wg)
	}
	wg.Wait()

}
