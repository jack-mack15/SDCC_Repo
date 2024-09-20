package main

import (
	"fmt"
	"strconv"
	"time"
)

// interfaccia gossiper
type Gossiper interface {
	Gossip(id int)
	HandleGossipMessage(id int, string string)
	ReviveNode(id int)
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

	gossipMessage := writeMulticastGossipMessage(GetMyId(), GetMyPort(), strconv.Itoa(id))

	for _, value := range idMap {
		go SendMulticastMessage(gossipMessage, value)
	}
}

// funzione che viene eseguita quando ricevo un update da un nodo o
// quando ottengo il digest di un heartbeat
func (i BimodalGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("[PEER %d] BM, gossip message received from: %d, faul node: %s\n\n", GetMyId(), idSender, message)
	idArray := extractIdArrayFromMessage(message)

	if len(idArray) == 0 {
		return
	} else {
		//nodi fault di cui non ero a conoscenza
		idFaultNodes := CompareAndAddOfflineNodes(message)

		//aggiorno lo stato dei nodi in nodeClass
		for i := 0; i < len(idFaultNodes); i++ {
			UpdateNodeState(idFaultNodes[i])
		}
	}
}

// funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (i BimodalGossiper) ReviveNode(id int) {
	fmt.Printf("[PEER %d] BM node: %d has reactivated\n\n", GetMyId(), id)
	removeOfflineNode(id)
}

//BLIND RUMOR MONGERING

// struttura del blind rumor mongering
type BlindRumorGossiper struct{}

// funzione che gestisce un update ricevuto da un altro nodo
func (e BlindRumorGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("[PEER %d] BCRM, received gossip message from: %d fault node: %s\n\n", GetMyId(), idSender, message)
	//message deve contenere sempre un solo id fault
	updateIdArray := extractIdArrayFromMessage(message)
	faultId := updateIdArray[0]

	if len(updateIdArray) != 1 {
		return
	}

	if CheckPresenceFaultNodesList(faultId) {
		//blocco di codice se già ero a conoscenza del fault
		//rimuovo idSender dalla lista di nodi da notificare se fosse presente
		removeNodeToNotify(idSender, faultId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(faultId)

	} else {
		//blocco di codice se non ero a conoscenza del fault

		//aggiorno stato del nodo nella lista
		UpdateNodeState(faultId)

		//aggiungere struct per faultId
		addFaultNodeStruct(faultId)

		removeNodeToNotify(idSender, faultId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(faultId)

		gossiper.Gossip(faultId)
	}
}

// funzione che va a diffondere un update
func (e BlindRumorGossiper) Gossip(faultId int) {

	//aggiorno stato del nodo nella lista
	UpdateNodeState(faultId)

	//aggiungo struct per faultId se non esistesse
	addFaultNodeStruct(faultId)

	for {
		if checkLenToNotifyList(faultId) <= 0 || getParameterF(faultId) <= 0 {
			break
		}

		selectedNodes := getNodesToNotify(faultId)

		for i := 0; i < len(selectedNodes); i++ {
			removeNodeToNotify(selectedNodes[i], faultId)
			//invio del gossipmessage
			go sendBlindCounterGossipMessage(selectedNodes[i], faultId)
		}

		//TODO rimuovere la sleep?
		time.Sleep(2 * time.Second)

		decrementNumOfUpdateForId(faultId)
	}
}

// funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (e BlindRumorGossiper) ReviveNode(faultId int) {
	fmt.Printf("[PEER %d] BCRM, node: %d has reactivated\n\n", GetMyId(), faultId)
	removeUpdate(faultId)
}

//ANTI ENTROPY GOSSIP

// struttura di anti entropy
type AntiEntropyGossiper struct{}

// funzione che gestisce gli update ricevuti da un nodo
func (e AntiEntropyGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("Anti Entropy Handler, ricevuto nodo sus: %s\n", message)
	idArray := extractIdArrayFromMessage(message)

	if len(idArray) == 0 {
		return
	} else {
		//nodi fault di cui non ero a conoscenza
		idFaultNodes := CompareAndAddOfflineNodes(message)

		//aggiorno lo stato dei nodi in nodeClass
		for i := 0; i < len(idFaultNodes); i++ {
			UpdateNodeState(idFaultNodes[i])
		}
	}

	fmt.Println("Anti Entropy Handler, ho terminato")
}

// funzione che non fa nulla
func (e AntiEntropyGossiper) Gossip(faultId int) {

	//TODO implementare un ciclo for infinito con selezione random e
}

// funzione che rimuove un nodo fault che si è ripresentato nella rete
func (e AntiEntropyGossiper) ReviveNode(id int) {
	removeOfflineNode(id)
}

var gossiper Gossiper

func InitGossiper() {
	if GetGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
	}
}
