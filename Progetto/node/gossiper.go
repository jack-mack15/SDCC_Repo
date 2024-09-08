package node

import (
	"fmt"
	"strconv"
)

// interfaccia gossiper
type Gossiper interface {
	Gossip(id int)
	HandleGossipMessage(string string)
}

// struttura del bimodal multicast
type BimodalGossiper struct{}

// implementazione del bimodal multicast
// invio l'update a tutti i nodi che conosco
func (i BimodalGossiper) Gossip(id int) {

	idMap := GetNodesMulticast()

	message := "222#" + strconv.Itoa(id)

	for _, value := range idMap {
		go SendGossipSignal(message, value)
	}
}

// vedo se nel digest del messaggio è presente un id di un nodo fallito di cui non
// ero a conoscenza e in caso aggiorno il mio digest
func (i BimodalGossiper) HandleGossipMessage(message string) {

	CompareDigest(message)

	fmt.Printf("AYOOOOOOOOOOOOOOOOOOOOO: nodo sus is: %s\n", message)

}

// struttura del blind rumor mongering
type BlindRumorGossiper struct {
	//mappa che tiene traccia dei nodi fault che gli altri nodi non conoscono
	knownUpdates map[int][]int
	//TODO se uso questi due parametri devo aggiungerli al file di config
	//il parametro B visto nel corso
	maxNeighbourToContact int
	//il parametro F visto nel corso
	repetitionTimes int
}

// funzione che vede se il digest del messaggio contine id faul che non conosco
// e va aggiornare la mappa dei fault conosciuti dagli altri nodi
func (e BlindRumorGossiper) HandleGossipMessage(message string) {
	//updateList := ExtractArrayFromDigest(message)
	CompareDigest(message)

}

// implementazione del blind rumor mongering
func (e BlindRumorGossiper) Gossip(id int) {

}

var gossiper Gossiper

func InitGossiper() {
	if GetGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
	}
}
