package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
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

	//contatto il registry
	sdResponseString := contactRegistry(getOwnTCPAddr(), getSDInfoString())

	retry := 0
	howMany := 0
	for {
		howMany = extractNodeList(sdResponseString)
		if howMany > 0 || retry >= 10 {
			break
		} else {
			retry++
			time.Sleep(5 * time.Second)

			sdResponseString = contactRegistry(getOwnTCPAddr(), getSDInfoString())
		}
	}

	initLogFile()
	//sleep per dare tempo a netem di assegnare i ritardi o tirare su la rete
	time.Sleep(5 * time.Second)

	//avvio della goroutine di ricezione
	go receiverHandler()

	lazzarusTry = 2

	//FASE ATTIVA
	for {
		//scelgo i nodi da contattare
		nodesToContact := getNodesToContact()

		contactNode(nodesToContact)

		time.Sleep(time.Duration(getHBDelay()) * time.Second)

		printAllNodeList()

		activeNodeLenght := getLenght()
		if activeNodeLenght == 0 {
			tryLazzarus()
		}

		//TODO settare i lazzarus retry da file di conf

		//TODO controllare tutta la robba da eliminare
		//in receiverHanlder()
		//in HanldeUDP
		//blind rumor gossip, in gossip, ma forse è da sistemare in un dato del file config
		//in handleUDP ci stanno i vechi delay

		//TODO sistemazione del codice

		//TODO gestire meglio le chiusure dei canali?

	}
}

// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	conn, err := net.ListenUDP("udp", getOwnUDPAddr())
	if err != nil {
		fmt.Println("receiverHandler()--> errore creazione listener UDP:", err)
		return
	}
	defer conn.Close()

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

// funzione che riceve il messaggio di risposta da il service registry, ottiene id del nodo attuale e
// completa la lista dei nodi che conosce il nodo attuale
func extractNodeList(str string) int {
	count := strings.Count(str, "#")
	nodeCount := 0

	//se sono il primo della rete count == 0
	if count == 0 {
		return nodeCount
	}

	count++

	parts := strings.SplitN(str, "#", count)
	myId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	setMyId(myId)

	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if currId == myId {
			continue
		}

		currStrAddr := strings.TrimSpace(currNodeParts[1])
		currUDPAddr, err := net.ResolveUDPAddr("udp", currStrAddr)
		if err != nil {
			log.Printf("extractNodeList()---> errore risoluzione indirizzo remoto %s: %v", currStrAddr, err)
		}

		addActiveNode(currId, 0, currUDPAddr)
		nodeCount++

	}

	return nodeCount
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
