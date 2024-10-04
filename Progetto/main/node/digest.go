package main

//file che gestisce la lista dei nodi della rete che hanno subito un fallimento e sono disattivi
//mantiene e aggiorna anche il digest che viene allegato in fase di "gossip repair"

import (
	"strconv"
	"strings"
	"sync"
)

var offlineNodes []int
var offlineNodesMutex sync.Mutex

// funzione che aggiunge un nodo alla lista e aggiorna il digest
func addOfflineNode(id int) {
	offlineNodesMutex.Lock()

	i := 0
	for i < len(offlineNodes) && offlineNodes[i] < id {
		i++
	}

	if i < len(offlineNodes) && offlineNodes[i] == id {
		offlineNodesMutex.Unlock()
		return
	}

	offlineNodes = append(offlineNodes[:i], append([]int{id}, offlineNodes[i:]...)...)

	offlineNodesMutex.Unlock()
}

// funzione che ritorna il digest da allegare ad un messaggio
func getDigest() string {
	offlineNodesMutex.Lock()

	digest := ""

	for i := 0; i < len(offlineNodes); i++ {
		if i != 0 {
			digest += "/"
		}
		stringElem := strconv.Itoa(offlineNodes[i])
		digest = digest + stringElem
	}

	offlineNodesMutex.Unlock()

	return digest
}

// funzione che rimuove un elemento dalla lista dei nodi offline
func removeOfflineNode(id int) {
	offlineNodesMutex.Lock()
	for i := 0; i < len(offlineNodes); i++ {
		if offlineNodes[i] == id {
			offlineNodes = append(offlineNodes[:i], offlineNodes[i+1:]...)
			break
		}
	}
	offlineNodesMutex.Unlock()
}

// funzione che riceve un digest di un altro nodo e lo confronta con il proprio digest
// ritorna una lista di id di nodi fault di cui non ero a conoscenza
func compareAndAddOfflineNodes(remoteDigest string) []int {

	ownArray := extractIdArrayFromMessage(getDigest())
	remoteArray := extractIdArrayFromMessage(remoteDigest)

	var didntKnow []int

	//condizione verificata se non conosco nessuno
	if len(ownArray) == 0 {
		updateOfflineNodes(remoteArray)
		return remoteArray
	}

	for i := 0; i < len(remoteArray); i++ {
		if !checkPresenceFaultNodesList(remoteArray[i]) {
			didntKnow = append(didntKnow, remoteArray[i])
			addOfflineNode(remoteArray[i])
		}
	}

	return didntKnow
}

// funzione che viene attivata da compareAndAddOfflineNodes se ci sono nodi falliti di cui non sono a conoscenza
func updateOfflineNodes(idArray []int) {
	for i := 0; i < len(idArray); i++ {

		addOfflineNode(idArray[i])
		updateNodeStateToFault(idArray[i])
	}
}

// funzione di ausilio che mi trasforma il contenuto di un messaggio di gossip da stringa a array di interi
func extractIdArrayFromMessage(digest string) []int {
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
