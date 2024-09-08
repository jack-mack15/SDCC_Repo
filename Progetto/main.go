package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"awesomeProject/node"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var gossiper node.Gossiper

func main() {

	//SET UP del nodo
	err := node.ReadConfigFile()
	if err == 0 {
		fmt.Println("errore nel recupero del file di conf")
		return
	}

	//"istanzio" un gossiper in base al file di config
	node.InitGossiper()

	//ottengo un numero di porta da so e ottengo il mio indirizzo
	listener, err2 := net.Listen("tcp", ":0")
	if err2 != nil {
		log.Fatalf("errore numero porta: %v", err)
	}
	myPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	node.SetOwnUDPAddr(&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort})
	node.SetOwnTCPAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort})

	//contatto il registry
	sdResponseString := node.ContactRegistry(node.GetOwnTCPAddr(), node.GetSDInfoString())

	howMany := extractNodeList(sdResponseString)
	if howMany == 0 {
		//se sono il primo a contattarlo, ritento il contatto fino a che un altro nodo si collega
		//utilizzo un tot massimo di tentativi
		for count := 0; count < 5; count++ {
			if howMany == 0 && count < 5 {
				time.Sleep(5 * time.Second)
				sdResponseString = node.ContactRegistry(node.GetOwnTCPAddr(), node.GetSDInfoString())
				howMany = extractNodeList(sdResponseString)
			} else {
				break
			}
		}
	}

	//FASE ATTIVA

	go receiverHandler()

	time.Sleep(5 * time.Second)

	//TODO cambiare tutto, lasciar fare al nodeClass
	for {
		//scelgo i nodi da contattare
		nodesToContact := node.GetNodeToContact()

		go contactNode(nodesToContact)

		time.Sleep(5 * time.Second)
		node.PrintAllNodeList()

		//TODO in sendHeartBeat() nella deadline ci sta un "* 3" da modificare
		//TODO scelta tra "blind counter rumor mongering" e "bimodal multicast"
		//TODO ad ogni lettura e scrittura aggiungere un timeout
		//TODO sistemare le approssimazioni e il calcolo della distanza e tempo di risposta
		//TODO aggiungere il digest al heartbeat?
		//TODO notificare anche il service registry dopo un fault, volendo comportamento settabile da impostazioni

		//MEGA TODO aggiungere in tutte le porzioni di codice, gestioni di fallimenti dei nodi contattati

	}
}

// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	conn, err := net.ListenUDP("udp", node.GetOwnUDPAddr())
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

		go node.HandleUDPMessage(conn, remoteUDPAddr, buffer[:n])
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
	node.SetMyId(myId)

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

		check := node.AddActiveNode(currId, currStrAddr, currUDPAddr, currTCPAddr)

		if !check {
			nodeCount++
		}
	}

	return nodeCount
}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(selectedNodes []node.Node) {

	//contatto i nodi
	lenght := len(selectedNodes)

	var wg sync.WaitGroup

	for i := 0; i < lenght; i++ {
		wg.Add(1)
		go node.SendHeartbeat(selectedNodes[i], node.GetMyId(), &wg)
	}
	wg.Wait()
}
