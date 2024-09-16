package node

//file che gestisce la lista dei nodi della rete che hanno subito un fallimento e sono disattivi
//mantiene e aggiorna anche il digest che viene allegato in fase di "gossip repair"

import (
	"sort"
	"strconv"
	"sync"
)

var offlineNodes []int
var digestToSend string
var digestMutex sync.Mutex

// funzione che aggiunge un nodo alla lista e aggiorna il digest
// TODO inserimento nel posto corretto e rimuovere il sort
func AddOfflineNode(id int) {
	digestMutex.Lock()
	offlineNodes = append(offlineNodes, id)
	sortOfflineNodes()
	digestToSend = ""

	for i := 0; i < len(offlineNodes); i++ {
		digestToSend = digestToSend + "/" + strconv.Itoa(offlineNodes[i])
	}

	digestMutex.Unlock()
}

// funzione che ritorna il digest da allegare ad un messaggio
func GetDigest() string {
	if len(offlineNodes) == 0 {
		return ""
	}
	return digestToSend
}

// funzione che riceve un digest di un altro nodo e lo confronta con il proprio digest
// ritorna una lista di id di nodi fault di cui non ero a conoscenza
func CompareAndAddToDigest(remoteDigest string) []int {

	ownArray := extractIdArrayFromMessage(digestToSend)
	remoteArray := extractIdArrayFromMessage(remoteDigest)

	var didntKnow []int

	//condizione verificata se non conosco nessuno
	if len(ownArray) == 0 {
		UpdateDigest(remoteArray)
		return didntKnow
	}

	for i := 0; i < len(remoteArray); i++ {
		if !CheckPresenceActiveNodesList(remoteArray[i]) {
			didntKnow = append(didntKnow, remoteArray[i])
			AddOfflineNode(remoteArray[i])
		}
	}

	return didntKnow
}

// funzione che viene attivata da compareDigest se ci sono nodi falliti di cui non sono a conoscenza
func UpdateDigest(idArray []int) {
	for i := 0; i < len(idArray); i++ {
		if !CheckPresenceActiveNodesList(idArray[i]) {
			AddOfflineNode(idArray[i])
		}
		UpdateNodeState(idArray[i])
	}
}

// funzione che verifica se il nodo è stato già segnalato come fallito
// ritorna true se è presente, false altrimenti
func CheckPresenceDigestList(id int) bool {
	digestMutex.Lock()

	for i := 0; i < len(offlineNodes); i++ {
		if offlineNodes[i] == id {
			digestMutex.Unlock()
			return true
		}
	}

	digestMutex.Unlock()
	return false
}

// funzione di ausilio che ordina gli elementi della lista di ID falliti
func sortOfflineNodes() {
	sort.Ints(offlineNodes)
}
