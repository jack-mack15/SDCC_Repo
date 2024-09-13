package node

import (
	"fmt"
	"strconv"
	"time"
)

// interfaccia gossiper
type Gossiper interface {
	Gossip(id int)
	HandleGossipMessage(id int, string string)
}

//BIMODAL MULTICAST

// struttura del bimodal multicast
type BimodalGossiper struct{}

// implementazione del bimodal multicast
// invio l'update a tutti i nodi che conosco
func (i BimodalGossiper) Gossip(id int) {

	//cambio dello stato del nodo nella lista di nodi
	UpdateNodeState(id)
	//aggiungo l'id del nodo fault al digest
	AddOfflineNode(id)

	//recupero id di tutti i nodi da contattare in multicast
	idMap := GetNodesMulticast()
	fmt.Printf("Bimodal Multicast Gossip, invio update a tutti: %v\n", idMap)

	gossipMessage := writeMulticastGossipMessage(GetMyId(), GetMyPort(), strconv.Itoa(id))

	for _, value := range idMap {
		go SendMulticastMessage(gossipMessage, value)
	}
}

// vedo se nel digest del messaggio è presente un id di un nodo fallito di cui non
// ero a conoscenza e in caso aggiorno il mio digest
func (i BimodalGossiper) HandleGossipMessage(_ int, message string) {

	fmt.Printf("Bimodal Multicast Handler, ricevuto nodo sus: %s\n", message)
	idArray := ExtractArrayFromDigest(message)

	if len(idArray) == 0 {
		return
	} else {
		//nodi fault di cui non ero a conoscenza
		idFaultNodes := CompareAndAddToDigest(message)

		//aggiorno lo stato dei nodi in nodeClass
		for i := 0; i < len(idFaultNodes); i++ {
			UpdateNodeState(idFaultNodes[i])
		}
	}

	fmt.Println("Bimodal Multicast Handler, ho terminato")
}

//BLIND RUMOR MONGERING

// struttura del blind rumor mongering
type BlindRumorGossiper struct{}

// funzione che vede se il digest del messaggio contine id faul che non conosco
// e va aggiornare la mappa dei fault conosciuti dagli altri nodi
func (e BlindRumorGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("Blind Counter Handler, ricevuto nodo sus: %s\n", message)
	//message deve contenere sempre un solo id fault
	updateIdArray := ExtractArrayFromDigest(message)
	updateId := updateIdArray[0]

	if len(updateIdArray) != 1 {
		return
	}

	if CheckPresenceDigestList(updateId) {
		//blocco di codice se già ero a conoscenza del fault
		//rimuovo idSender dalla lista di nodi da notificare se fosse presente
		removeNodeToNotify(idSender, updateId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(updateId)

	} else {
		//blocco di codice se non ero a conoscenza del fault

		//aggiungere struct per updateId
		addFaultNodeStruct(updateId)
		removeNodeToNotify(idSender, updateId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(updateId)

		go gossiper.Gossip(updateId)
	}

	fmt.Println("Blind Counter Handler, ho terminato")

}

// implementazione del blind rumor mongering
func (e BlindRumorGossiper) Gossip(faultId int) {

	//aggiungo nodo al digest
	AddOfflineNode(faultId)

	//aggiorno stato del nodo nella lista
	UpdateNodeState(faultId)

	//aggiungo struct per faultId se non esistesse
	addFaultNodeStruct(faultId)

	for {
		if checkLenToNotifyList(faultId) <= 0 && getParameterF(faultId) <= 0 {
			break
		}

		selectedNodes := getNodesToNotify(faultId)

		for i := 0; i < len(selectedNodes); i++ {
			removeNodeToNotify(selectedNodes[i], faultId)
			//invio del gossipmessage
			fmt.Printf("Blind Counter Gossip, invio update to: %d\n", selectedNodes[i])
			go sendBlindCounterGossipMessage(selectedNodes[i], faultId)
		}

		//TODO rimuovere la sleep?
		time.Sleep(5 * time.Second)

		decrementNumOfUpdateForId(faultId)
	}
}

var gossiper Gossiper

func InitGossiper() {
	if GetGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
	}
}
