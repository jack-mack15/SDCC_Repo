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

	howMany := extractNodeList(sdResponseString)
	if howMany == 0 {
		//se sono il primo a contattarlo, ritento il contatto fino a che un altro nodo si collega
		//utilizzo un tot massimo di tentativi
		for count := 0; count < 5; count++ {
			if howMany == 0 && count < 5 {
				time.Sleep(5 * time.Second)
				sdResponseString = ContactRegistry(GetOwnTCPAddr(), GetSDInfoString())
				howMany = extractNodeList(sdResponseString)
			} else {
				break
			}
		}
	}

	//FASE ATTIVA

	go receiverHandler()

	time.Sleep(5 * time.Second)

	for {
		//scelgo i nodi da contattare
		nodesToContact := GetNodeToContact()

		contactNode(nodesToContact)

		//TODO eliminare questa sleep
		if GetMyId() == 3 {
			time.Sleep(8 * time.Second)
		}
		time.Sleep(4 * time.Second)

		PrintAllNodeList()

		//TODO controllare tutta la robba da eliminare
		//in receiverHanlder()
		//in HanldeUDP
		//blind rumor gossip, in gossip, ma forse è da sistemare in un dato del file config
		//in handleUDP ci stanno i vechi delay

		//TODO sistemazione del codice

		//TODO gestire meglio le chiusure dei canali?

		//TODO riguardare il modo con cui si contatta il registry

		//TODO in sendHeartBeat() nella deadline ci sta un "* 3" da modificare e mettere un valore confiigurabile da file

		//TODO aggiungere anche anti entropy: ovvero seleziono randomicamente un solo nodo e gli dico quello che so
		//se proprio serve eh

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

		//TODO elimina questa parte
		if GetMyId() == 3 {
			time.Sleep(8 * time.Second)
		}
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

		check := AddActiveNode(currId, 0, currStrAddr, currUDPAddr, currTCPAddr)

		if !check {
			nodeCount++
		}
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
