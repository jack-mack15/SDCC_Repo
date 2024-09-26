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

	//decremento il numero di retry
	if decrementNumberOfRetry(id) {
		return
	}

	//aggiungo l'id del nodo fault al digest
	addOfflineNode(id)

	//recupero id di tutti i nodi da contattare in multicast
	idMap := getNodesMulticast()

	gossipMessage := writeMulticastGossipMessage(getMyId(), getMyPort(), strconv.Itoa(id))

	for idNode, value := range idMap {
		fmt.Printf("[PEER %d] BM, gossip message send to: %d, fault node: %d\n", getMyId(), idNode, id)
		go sendMulticastMessage(gossipMessage, value)
	}
}

// funzione che viene eseguita quando ricevo un update da un nodo o
// quando ottengo il digest di un heartbeat
func (i BimodalGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("[PEER %d] BM, gossip message received from: %d, fault node: %s\n", getMyId(), idSender, message)
	idArray := extractIdArrayFromMessage(message)

	if len(idArray) == 0 {
		fmt.Printf("[PEER %d] BM, no digest from sender: %d\n", getMyId(), idSender)
		return
	} else {
		//nodi fault di cui non ero a conoscenza
		idFaultNodes := compareAndAddOfflineNodes(message)
		if len(idFaultNodes) == 0 {
			fmt.Printf("[PEER %d] BM, faults from sender: %d, already known\n", getMyId(), idSender)
			return
		}
		fmt.Printf("[PEER %d] BM, from sender: %d, discovered this faults: %v\n", getMyId(), idSender, idFaultNodes)
		//aggiorno lo stato dei nodi in nodeClass
		for i := 0; i < len(idFaultNodes); i++ {
			updateNodeStateToFault(idFaultNodes[i])
		}
	}
}

// funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (i BimodalGossiper) ReviveNode(id int) {
	fmt.Printf("[PEER %d] BM LAZZARUS node: %d\n", getMyId(), id)
	removeOfflineNode(id)
}

//BLIND RUMOR MONGERING

// struttura del blind rumor mongering
type BlindRumorGossiper struct{}

// funzione che gestisce un update ricevuto da un altro nodo
func (e BlindRumorGossiper) HandleGossipMessage(idSender int, message string) {

	fmt.Printf("[PEER %d] BCRM, received gossip message from: %d fault node: %s\n", getMyId(), idSender, message)
	//message deve contenere sempre un solo id fault
	updateIdArray := extractIdArrayFromMessage(message)
	faultId := updateIdArray[0]

	if len(updateIdArray) != 1 {
		return
	}

	if checkPresenceFaultNodesList(faultId) {
		//blocco di codice se già ero a conoscenza del fault
		//rimuovo idSender dalla lista di nodi da notificare se fosse presente
		fmt.Printf("[PEER %d] BCRM, gossip from: %d about fault node: %s already known\n", getMyId(), idSender, message)

		removeNodeToNotify(idSender, faultId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(faultId)

	} else {
		//blocco di codice se non ero a conoscenza del fault
		fmt.Printf("[PEER %d] BCRM, gossip from: %d about fault node: %s added to my knowledge\n", getMyId(), idSender, message)
		//aggiorno stato del nodo nella lista
		updateNodeStateToFault(faultId)

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

	//vado a decrementare il numero di retry
	if decrementNumberOfRetry(faultId) {
		return
	}

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
			fmt.Printf("[PEER %d] BCRM, sending gossip message to: %d about fault node: %d\n",
				getMyId(), selectedNodes[i], faultId)
			go sendBlindCounterGossipMessage(selectedNodes[i], faultId)
		}

		fmt.Printf("[PEER %d] BCRM, gossip iteration done\n", getMyId())

		time.Sleep(time.Duration(getGossipInterval()) * time.Millisecond)

		decrementNumOfUpdateForId(faultId)
	}
}

// funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (e BlindRumorGossiper) ReviveNode(faultId int) {
	fmt.Printf("[PEER %d] BCRM, LAZZARUS node: %d\n", getMyId(), faultId)
	removeUpdate(faultId)
}

var gossiper Gossiper

func InitGossiper() {
	if getGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
	}
}
