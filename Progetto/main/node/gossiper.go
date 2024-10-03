package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// FaultGossiper interfaccia gossiper
type FaultGossiper interface {
	GossipFault(id int)
	HandleGossipFaultMessage(id int, message GossipMessage)
	ReviveNode(id int)
}

//BIMODAL MULTICAST

// BimodalGossiper struttura del bimodal multicast
type BimodalGossiper struct{}

// Gossip implementazione del bimodal multicast
// invio l'update a tutti i nodi che conosco
func (i BimodalGossiper) GossipFault(id int) {

	//decremento il numero di retry
	if decrementNumberOfRetry(id) {
		return
	}

	//aggiungo l'id del nodo fault al digest
	addOfflineNode(id)
	//rimuovo il nodo dalla lista delle coordinate
	removeNode(id)
	//recupero id di tutti i nodi da contattare in multicast
	idMap := getNodesMulticast()

	gossipMessage := writeGossipMessage(id)

	message, err := json.Marshal(gossipMessage)
	if err != nil {
		log.Fatalf("GossipFault()--> errore codifica JSON: %v", err)
		return
	}

	for idNode, value := range idMap {
		fmt.Printf("[PEER %d] BM, gossip message send to: %d, fault node: %d\n", getMyId(), idNode, id)
		go sendMulticastMessage(message, value)
	}
}

// HandleGossipMessage funzione che viene eseguita quando ricevo un update da un nodo o
// quando ottengo il digest di un heartbeat
func (i BimodalGossiper) HandleGossipFaultMessage(idSender int, message GossipMessage) {

	fmt.Printf("[PEER %d] BM, gossip message received from: %d, fault node: %s\n", getMyId(), idSender, message.Digest)
	idArray := extractIdArrayFromMessage(message.Digest)

	if len(idArray) == 0 {
		fmt.Printf("[PEER %d] BM, no digest from sender: %d\n", getMyId(), idSender)
		return
	} else {
		//nodi fault di cui non ero a conoscenza
		idFaultNodes := compareAndAddOfflineNodes(message.Digest)
		if len(idFaultNodes) == 0 {
			fmt.Printf("[PEER %d] BM, faults from sender: %d, already known\n", getMyId(), idSender)
			return
		}
		fmt.Printf("[PEER %d] BM, from sender: %d, discovered this faults: %v\n", getMyId(), idSender, idFaultNodes)
		//aggiorno lo stato dei nodi in nodeClass
		for i := 0; i < len(idFaultNodes); i++ {
			updateNodeStateToFault(idFaultNodes[i])
			//rimuovo il nodo da quelli delle coordinate
			removeNode(idFaultNodes[i])
		}
	}
}

// ReviveNode funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (i BimodalGossiper) ReviveNode(id int) {
	fmt.Printf("[PEER %d] BM LAZZARUS node: %d\n", getMyId(), id)
	removeOfflineNode(id)
}

//BLIND RUMOR MONGERING

// BlindRumorGossiper struttura del blind rumor mongering
type BlindRumorGossiper struct{}

// HandleGossipMessage funzione che gestisce un update ricevuto da un altro nodo
func (e BlindRumorGossiper) HandleGossipFaultMessage(idSender int, message GossipMessage) {

	fmt.Printf("[PEER %d] BCRM, received gossip message from: %d fault node: %d\n", getMyId(), idSender, message.IdFault)
	//message deve contenere sempre un solo id fault
	faultId := message.IdFault

	if checkPresenceFaultNodesList(faultId) {
		//blocco di codice se già ero a conoscenza del fault
		//rimuovo idSender dalla lista di nodi da notificare se fosse presente
		fmt.Printf("[PEER %d] BCRM, gossip from: %d about fault node: %d already known\n", getMyId(), idSender, message.IdFault)

		removeNodeToNotify(idSender, faultId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(faultId)

	} else {
		//blocco di codice se non ero a conoscenza del fault
		fmt.Printf("[PEER %d] BCRM, gossip from: %d about fault node: %d added to my knowledge\n", getMyId(), idSender, message.IdFault)
		//aggiorno stato del nodo nella lista
		updateNodeStateToFault(faultId)
		//rimuovo il nodo dalla lista di coordinat
		removeNode(faultId)
		//aggiungere struct per faultId
		addFaultNodeStruct(faultId)

		removeNodeToNotify(idSender, faultId)

		//decremento il contatore delle massime ripetizioni dell'update
		decrementNumOfUpdateForId(faultId)

		gossiper.GossipFault(faultId)
	}
}

// Gossip funzione che va a diffondere un update
func (e BlindRumorGossiper) GossipFault(faultId int) {

	//vado a decrementare il numero di retry
	if decrementNumberOfRetry(faultId) {
		return
	}
	//rimuovo il nodo dalla lista di coordinate
	removeNode(faultId)
	//aggiungo struct per faultId se non esistesse
	addFaultNodeStruct(faultId)

	for {
		if checkLenToNotifyList(faultId) <= 0 || getParameterF(faultId) <= 0 {
			break
		}

		selectedNodes := getNodesToNotify(faultId)
		gossipMessage := writeGossipMessage(faultId)
		message, err := json.Marshal(gossipMessage)
		if err != nil {
			log.Fatalf("GossipFault()--> errore codifica JSON: %v", err)
			return
		}

		for i := 0; i < len(selectedNodes); i++ {
			removeNodeToNotify(selectedNodes[i], faultId)
			//invio del gossipmessage
			fmt.Printf("[PEER %d] BCRM, sending gossip message to: %d about fault node: %d\n",
				getMyId(), selectedNodes[i], faultId)
			go sendBlindCounterGossipMessage(message, selectedNodes[i])
		}

		fmt.Printf("[PEER %d] BCRM, gossip iteration done\n", getMyId())

		time.Sleep(time.Duration(getGossipInterval()) * time.Millisecond)

		decrementNumOfUpdateForId(faultId)
	}
}

// ReviveNode funzione che gestisce il caso in cui un nodo fault si ripresenta nella rete
func (e BlindRumorGossiper) ReviveNode(faultId int) {
	fmt.Printf("[PEER %d] BCRM, LAZZARUS node: %d\n", getMyId(), faultId)
	removeUpdate(faultId)
}

var gossiper FaultGossiper

func InitGossiper() {
	if getGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
	}
}
