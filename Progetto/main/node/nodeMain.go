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
	err := readEnvVariable()
	if err == 0 {
		fmt.Println("errore nel recupero del file di conf")
		return
	}

	//"istanzio" un gossiper in base al file di config
	InitGossiper()

	//inizializzo le mie coordinate
	initMyCoordination()
	//contatto il registry
	tryContactRegistry()

	//avvio della goroutine di ricezione
	go receiverHandler()

	//sleep per dare tempo a netem di assegnare i ritardi o tirare su la rete
	time.Sleep(3 * time.Second)

	counter := 0
	//FASE ATTIVA
	for {
		//scelgo i nodi da contattare
		nodesToContact := getNodesToContact()

		contactNode(nodesToContact)

		time.Sleep(time.Duration(getHBDelay()) * time.Millisecond)

		counter++
		if counter == getPrintCounter() {
			printAllNodeList()
			printAllCoordinates()
			counter = 0
		}

		activeNodeLenght := getLenght()
		if activeNodeLenght == 0 {
			tryLazzarus()
		}
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
