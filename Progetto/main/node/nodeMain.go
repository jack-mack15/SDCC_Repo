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
	err := ReadConfigFile()
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
	SetOwnUDPAddr(&net.UDPAddr{IP: ownIP, Port: GetMyPort()})
	SetOwnTCPAddr(&net.TCPAddr{IP: ownIP, Port: GetMyPort()})

	//contatto il registry
	sdResponseString := ContactRegistry(GetOwnTCPAddr(), GetSDInfoString())

	retry := 0
	howMany := 0
	for {
		howMany = extractNodeList(sdResponseString)
		if howMany > 0 || retry >= 10 {
			break
		} else {
			retry++
			time.Sleep(5 * time.Second)

			sdResponseString = ContactRegistry(GetOwnTCPAddr(), GetSDInfoString())
		}
	}

	//FASE ATTIVA
	initLogFile()
	//sleep per dare tempo a netem di assegnare i ritardi o tirare su la rete
	time.Sleep(5 * time.Second)
	go receiverHandler()

	lazzarusTry = 2

	for {
		//scelgo i nodi da contattare
		nodesToContact := GetNodeToContact()

		contactNode(nodesToContact)

		time.Sleep(time.Duration(getHBDelay()) * time.Second)

		PrintAllNodeList()

		activeNodeLenght := getLenght()
		if activeNodeLenght == 0 {
			tryLazzarus()
		}

		//TODO controllare tutta la robba da eliminare
		//in receiverHanlder()
		//in HanldeUDP
		//blind rumor gossip, in gossip, ma forse è da sistemare in un dato del file config
		//in handleUDP ci stanno i vechi delay

		//TODO sistemazione del codice

		//TODO gestire meglio le chiusure dei canali?

		//TODO sistema il fatto che se un nodo mi invia heartbeat, io devo aggiornargli lo stato, per il momento non lo fa

		//TODO sistemare il fatto che se mi rifaccio vivo potrebbero esserci altrin nodi che continuano a professarmi morto

		//TODO implementare un file log su cui scrive un nodo tutti quelli che sa vivi o morti

	}
}

// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	conn, err := net.ListenUDP("udp", GetOwnUDPAddr())
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

		go HandleUDPMessage(conn, remoteUDPAddr, buffer[:n])

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
	SetMyId(myId)

	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if currId == myId {
			continue
		}

		currStrAddr := strings.TrimSpace(currNodeParts[1])
		currTCPAddr, err := net.ResolveTCPAddr("tcp", currStrAddr)
		currUDPAddr, err := net.ResolveUDPAddr("udp", currStrAddr)
		if err != nil {
			log.Printf("extractNodeList()---> errore risoluzione indirizzo remoto %s: %v", currStrAddr, err)
		}

		AddActiveNode(currId, 0, currStrAddr, currUDPAddr, currTCPAddr)
		nodeCount++

	}

	return nodeCount
}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(selectedNodes []Node) {

	//contatto i nodi
	lenght := len(selectedNodes)

	var wg sync.WaitGroup

	for i := 0; i < lenght; i++ {
		wg.Add(1)
		go SendHeartbeat(selectedNodes[i], GetMyId(), &wg)
	}
	wg.Wait()

}
