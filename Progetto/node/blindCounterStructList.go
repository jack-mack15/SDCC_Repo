package node

import (
	"math/rand"
	"sync"
)

// struct di ausilio per il blind rumor mongering
var updatesList []*blindInfoStruct
var updateListMutex sync.Mutex

type blindInfoStruct struct {
	//mutex per la manipolazione della struct
	structMutex sync.Mutex
	//id nodo fault da notificare
	id int
	//nodi che so che non conoscono l'update
	toNotify []int
	//numero di nodi massimo che posso contattare ad ogni iterazione
	b int
	//numero di volte che posso ancora diffondere l'update
	f int
}

// funzione di ausilio per il blind rumor gossip. verifica se un update è già conosciuto
// ritorna true se ero a conoscenza del fault, false altrimenti
func checkUpdate(id int) bool {
	updateListMutex.Lock()

	check := false
	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == id {
			check = true
			break
		}
	}

	updateListMutex.Unlock()

	if !check {
		addFaultNodeStruct(id)
		return false
	}

	return true
}

// funzione cha aggiunge un nodo alla lista di struct se non fosse presente
func addFaultNodeStruct(faultId int) {

	updateListMutex.Lock()
	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			updateListMutex.Unlock()
			return
		}
	}

	currElem := blindInfoStruct{}
	currElem.id = faultId
	currElem.f = GetMaxIter()
	currElem.b = GetMaxNeighbour()
	currElem.toNotify = GetNodesId()

	updatesList = append(updatesList, &currElem)
	updateListMutex.Unlock()

}

// funzione che verifica se senderId è tra i nodi ancora da notificare per il fault faultId
// se è presente elimina tale id dalla lista da notificare
func removeNodeToNotify(senderId int, faultId int) {
	updateListMutex.Lock()

	var currStruct *blindInfoStruct

	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			currStruct = updatesList[i]
			break
		}
	}

	updateListMutex.Unlock()

	if currStruct == nil {
		return
	}
	currStruct.structMutex.Lock()

	for i := 0; i < len(currStruct.toNotify); i++ {
		//blocco di codice se senderId è presente
		if currStruct.toNotify[i] == senderId {
			//rimuovo senderId dalla lista poichè conosco il fault
			currStruct.toNotify = append(currStruct.toNotify[:i], currStruct.toNotify[i+1:]...)
			if len(currStruct.toNotify) == 0 {
				removeStruct(faultId)
			}
			break
		}
	}

	currStruct.structMutex.Unlock()

	return
}

// funzione che verifica se tutti i nodi da notificare per il fault sono stati notificati.
// in tale caso elimina tale struct dalla lista
func removeStruct(faultId int) {
	updateListMutex.Lock()

	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			//rimuovo la struct
			updatesList = append(updatesList[:i], updatesList[i+1:]...)
		}
	}

	updateListMutex.Unlock()
}

// funzione che decrementa il contatore f della struct di faultId
// f indica quante volte sono stato contattato da un update su faultId
// se f = 0 allora per il blind counter, non diffonderò più update su faultId
// ritorna 0 se f <= 0 e quindi è stata eliminata la struct
func decrementNumOfUpdateForId(faultId int) {
	updateListMutex.Lock()

	var currStruct *blindInfoStruct

	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			currStruct = updatesList[i]
			break
		}
	}

	updateListMutex.Unlock()

	if currStruct == nil {
		return
	}
	currStruct.structMutex.Lock()

	currStruct.f--
	if currStruct.f <= 0 {
		currStruct.structMutex.Unlock()
		removeStruct(faultId)
		return
	}

	currStruct.structMutex.Unlock()

	return
}

// funzione che restituisce gli id dei nodi da notificare
func getNodesToNotify(faultId int) []int {

	var nodesToNotifyList []int

	faultIdStruct := getStruct(faultId)
	faultIdStruct.structMutex.Lock()

	structToNotifyList := faultIdStruct.toNotify
	if len(structToNotifyList) == 0 {
		return nodesToNotifyList
	}

	nodesToNotifyList = randomBlindCounterSelection(structToNotifyList)

	faultIdStruct.structMutex.Unlock()

	return nodesToNotifyList
}

// funzione che restituisce b nodi da contattare in modo randomico
func randomBlindCounterSelection(idArray []int) []int {
	//mutex lockato dal chiamante
	elemToContact := make(map[int]bool)
	var selectedNodes []int

	lenght := len(idArray)

	i := 0
	b := GetMaxNeighbour()

	if b >= lenght {
		//ho meno nodi da contattare del numero massimo di nodi da contattare per iterazione
		return idArray
	}

	//qui scelgo in modo randomico
	for i < b {
		random := rand.Intn(lenght)
		_, ok := elemToContact[random]
		if !ok {
			elemToContact[random] = true
			selectedNodes = append(selectedNodes, idArray[random])
			i++
		} else {
			continue
		}
	}

	return selectedNodes
}

// funzione che recupera la struct del faultId
func getStruct(faultId int) *blindInfoStruct {
	updateListMutex.Lock()

	var currStruct *blindInfoStruct

	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			currStruct = updatesList[i]
			break
		}
	}
	updateListMutex.Unlock()

	if currStruct == nil {
		return nil
	}
	return currStruct
}

// funzione che restituisce il parametro f di una struct
func getParameterF(faultId int) int {
	faultStruct := getStruct(faultId)
	if faultStruct == nil {
		return 0
	}
	return getStruct(faultId).f
}

// funzione che restituisce la lunghezza dell'array di interi ToNotify di faultId
func checkLenToNotifyList(faultId int) int {
	updateListMutex.Lock()

	var currStruct *blindInfoStruct

	for i := 0; i < len(updatesList); i++ {
		if updatesList[i].id == faultId {
			currStruct = updatesList[i]
			break
		}
	}

	updateListMutex.Unlock()

	if currStruct == nil {
		return 0
	}

	currStruct.structMutex.Lock()

	lenght := len(currStruct.toNotify)

	currStruct.structMutex.Unlock()

	return lenght
}
