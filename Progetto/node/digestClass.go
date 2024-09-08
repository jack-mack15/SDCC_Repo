package node

//file che gestisce la lista dei nodi della rete che hanno subito un fallimento e sono disattivi
//mantiene e aggiorna anche il digest che viene allegato in fase di "gossip repair"

import (
	"sort"
	"strconv"
	"strings"
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

// funzione che riceve un digest di un altro nodo e lo confronta con il digest del nuovo attuale
func CompareDigest(remoteDigest string) []int {
	ownArray := ExtractArrayFromDigest(digestToSend)
	remoteArray := ExtractArrayFromDigest(remoteDigest)

	//array che tiene conto di ciò che manca a me
	var ownProblems []int
	//array che tiene conto di ciò che manca al nodo che mi ha contattato
	var remoteProblems []int

	for {
		if len(ownArray) == 0 || len(remoteArray) == 0 {
			break
		}
		currOwnId := ownArray[0]
		currRemoteId := remoteArray[0]

		//currOwnId è presente in ownDigest ma non in remoteDigest
		//lo aggiungo ai problemi di remote e lo tolgo da ownArray
		if currOwnId < currRemoteId {
			remoteProblems = append(remoteProblems, currOwnId)
			ownArray = ownArray[1:]
		}
		//currRemoteId è presente in remoteDigest ma non in ownDigest
		//lo aggiungo ai problemi di own e lo tolgo da remoteArray
		if currOwnId > currRemoteId {
			ownProblems = append(ownProblems, currRemoteId)
			remoteArray = remoteArray[1:]
		}
		//currOwnId è presente in tutti e due i digest
		//quindi rimuovo quel valore da tutti e due
		if currOwnId == currRemoteId {
			ownArray = ownArray[1:]
			remoteArray = remoteArray[1:]
		}
	}
	//se uno dei due array ha ancora elementi, significa che questi non sono presenti nell'altro
	//caso in cui ownArray ha ancora elementi, li aggiungo a remoteProblems
	if len(ownArray) > 0 {
		for i := 0; i < len(ownArray); i++ {
			remoteProblems = append(remoteProblems, ownArray[i])
		}
	}
	//caso in cui remoteArray ha ancora elementi, li aggiungo a ownProblems
	if len(remoteArray) > 0 {
		for i := 0; i < len(remoteArray); i++ {
			ownProblems = append(ownProblems, remoteArray[i])
		}
	}

	//se ho scoperto id che mancavano al mio digest, li aggiungo
	if len(ownProblems) > 0 {
		UpdateDigest(ownProblems)
	}

	return remoteProblems
}

// funzione che viene attivata da compareDigest se ci sono nodi falliti di cui non sono a conoscenza
func UpdateDigest(idArray []int) {
	for i := 0; i < len(idArray); i++ {
		if !CheckPresenceNodeList(idArray[i]) {
			AddOfflineNode(idArray[i])
		}
		UpdateFailureNode(idArray[i])
	}
}

// funzione di ausilio che mi trasforma un digest da stringa a array di interi
func ExtractArrayFromDigest(digest string) []int {
	var array []int

	if digest == "" {
		return array
	}

	count := strings.Count(digest, "/") + 1
	arrayElems := strings.SplitN(digest, "/", count)

	for i := 0; i < count; i++ {
		currId, _ := strconv.Atoi(arrayElems[i])
		array = append(array, currId)
	}

	return array
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
