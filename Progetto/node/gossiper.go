package node

import (
	"fmt"
	"strconv"
	"sync"
)

// interfaccia gossiper
type Gossiper interface {
	Gossip(id int)
	HandleGossipMessage(string string)
}

//BIMODAL MULTICAST

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

//BLIND RUMOR MONGERING

// struttura del blind rumor mongering
type BlindRumorGossiper struct{}

// struct di ausilio per il blind rumor mongering
var updatesList []blindInfoStruct
var updateListMutex sync.Mutex

// il parametro B visto nel corso
var maxNeighbourToContact int

// il parametro F visto nel corso
var repetitionTimes int

// funzione che vede se il digest del messaggio contine id faul che non conosco
// e va aggiornare la mappa dei fault conosciuti dagli altri nodi
func (e BlindRumorGossiper) HandleGossipMessage(message string) {
	//con compareDigest verifico se sono a conoscenza degli update ricevuti
	CompareDigest(message)
	//updateIds := ExtractArrayFromDigest(message)

}

// implementazione del blind rumor mongering
func (e BlindRumorGossiper) Gossip(id int) {
	checkAddUpdate(id)
}

// funzione di ausilio per il blind rumor gossip. verifica se un update è già conosciuto
func checkAddUpdate(id int) {
	updateListMutex.Lock()

	check := false
	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == id {
			check = true
			break
		}
	}

	if !check {
		currElem := blindInfoStruct{}
		currElem.id = id
		currElem.f = repetitionTimes
		currElem.b = maxNeighbourToContact
		currElem.toNotify = GetNodesId()

		updatesList = append(updatesList, currElem)
	}

	updateListMutex.Unlock()
}

// funzione che va a diffondere gli update per il blind rumor mongering
func blindRumorSpreading() {}

type blindInfoStruct struct {
	//id nodo fault da notificare
	id int
	//nodi che so che non conoscono l'update
	toNotify []int
	//numero di nodi massimo che posso contattare ad ogni iterazione
	b int
	//numero di volte che posso ancora diffondere l'update
	f int
}

var gossiper Gossiper

func InitGossiper(b int, f int) {
	if GetGossipType() == 2 {
		gossiper = &BlindRumorGossiper{}
	} else {
		gossiper = &BimodalGossiper{}
		maxNeighbourToContact = b
		repetitionTimes = f
	}
}
