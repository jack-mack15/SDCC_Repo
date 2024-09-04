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

// funzione che verifica se il nodo è stato già segnalato come fallito
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
