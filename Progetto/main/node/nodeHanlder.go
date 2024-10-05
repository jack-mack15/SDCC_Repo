package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
)

//file che mantiene traccia della lista di nodi attivi della rete

// struttura singolo nodo
type node struct {
	//ID del nodo assegnato dal service registry
	ID int
	//indirizzo per identificare nodo, tipo puntatore a UDPAddr
	Addr string
	//State indica lo stato in cui si trova il nodo: 0 non conosciuto, 1 attivo, 2 disattivato
	State int
	//tempo risposta del nodo all'ultimo messaggio, -1 indica che non è conosciuto
	ResponseTime int
	//numero di retry restanti prima di segnare il nodo fault
	Retry int
}

var ignoreIds []int
var ignoreMutex sync.Mutex

var nodesList []node

var faultNodesList []node

var activeNodesMutex sync.Mutex

var faultNodesMutex sync.Mutex

// funzione che restituisce l'indirizzo UDP di un nodo della lista
func getSelectedTCPAddress(id int) string {
	activeNodesMutex.Lock()

	for _, node := range nodesList {
		if node.ID == id {
			activeNodesMutex.Unlock()
			return node.Addr
		}
	}
	activeNodesMutex.Unlock()

	return ""
}

// funzione che verifica se un nodo è presente. ritorna true se è presente, false altrimenti
func checkPresenceActiveNodesList(id int) bool {
	activeNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			activeNodesMutex.Unlock()
			return true
		}
	}

	activeNodesMutex.Unlock()
	return false
}

// funzione che verifica se un nodo è presente nella lista dei nodi fault. ritorna true se è presente
func checkPresenceFaultNodesList(id int) bool {
	faultNodesMutex.Lock()

	for i := 0; i < len(faultNodesList); i++ {
		if faultNodesList[i].ID == id {
			faultNodesMutex.Unlock()
			return true
		}
	}

	faultNodesMutex.Unlock()
	return false
}

// funzione che aggiunge un nodo alla lista. ritorna true se il nodo è stato aggiunto, false altrimenti
func addActiveNode(id int, state int, address string) {
	if !checkPresenceActiveNodesList(id) {

		currNode := node{}
		currNode.ID = id
		currNode.Addr = address
		currNode.State = state
		currNode.ResponseTime = -1
		currNode.Retry = getMaxRetry()

		activeNodesMutex.Lock()
		nodesList = append(nodesList, currNode)
		activeNodesMutex.Unlock()
	}
	return
}

// funzione che sceglie i nodi da contattare in base al valore impostato nel file di configurazione
func getNodesToContact() []*node {
	//scelta dei nodi da contattare
	actualLen := getLenght()
	howManyToContact := getMaxNum()

	if getMaxNum() == 0 {
		//calcolo rad quadr e arrotondo per eccesso
		sqr := math.Sqrt(float64(actualLen))
		howManyToContact = int(math.Ceil(sqr))
	}

	var selectedNode []*node

	//contatto in modo randomico
	elemToContact := make(map[int]bool)

	activeNodesMutex.Lock()
	//ottengo la lunghezza della lista dei nodi attivi
	lenght := getLenght()
	//genero dei numeri casuali tutti differenti, corrispondono alla scelta di nodi da contattare
	i := 0

	//caso in cui conosco un solo nodo
	if lenght == 1 {
		selectedNode = append(selectedNode, &nodesList[0])
		activeNodesMutex.Unlock()
		return selectedNode
	}

	//caso in cui la lunghezza dei nodi vivi combaci con il numero massimo da contattare
	if lenght == howManyToContact || lenght < howManyToContact {
		for i := 0; i < lenght; i++ {
			selectedNode = append(selectedNode, &nodesList[i])
		}
		activeNodesMutex.Unlock()
		return selectedNode
	}

	//caso generico
	for i < howManyToContact {
		if i >= lenght {
			break
		}
		random := rand.Intn(lenght)
		_, ok := elemToContact[random]
		if !ok && nodesList[random].State != 2 {
			elemToContact[random] = true
			selectedNode = append(selectedNode, &nodesList[random])
			i++
		} else {
			continue
		}
	}
	activeNodesMutex.Unlock()

	return selectedNode

}

// funzione che restituisce tutti i nodi a cui inviare un messaggio di multicast
func getNodesMulticast() map[int]string {

	idMap := make(map[int]string)

	activeNodesMutex.Lock()
	lenght := getLenght()
	for i := 0; i < lenght; i++ {
		idMap[nodesList[i].ID] = nodesList[i].Addr
	}
	activeNodesMutex.Unlock()

	return idMap
}

// funzione che restituisce il tempo di risposta osservato di un nodo
func getNodeRtt(id int) int {
	activeNodesMutex.Lock()
	lenght := getLenght()

	for i := 0; i < lenght; i++ {
		if nodesList[i].ID == id {
			var rtt = nodesList[i].ResponseTime
			activeNodesMutex.Unlock()
			return rtt
		}
	}

	activeNodesMutex.Unlock()
	return getDefRTT()
}

// funzione che restituisce la lista di tutti gli id dei nodi conosciuti
func getNodesId() []int {
	var array []int

	activeNodesMutex.Lock()

	lenght := getLenght()
	for i := 0; i < lenght; i++ {
		array = append(array, nodesList[i].ID)
	}

	activeNodesMutex.Unlock()

	return array
}

// funzione che aggiorna un nodo della lista, aggiorna stato, distanza e tempo di risposta
func updateNodeDistance(id int, state int, responseTime int) {
	activeNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			nodesList[i].State = state
			nodesList[i].Retry = getMaxRetry()
			nodesList[i].ResponseTime = responseTime

			break
		}
	}

	activeNodesMutex.Unlock()
}

// funzione che va a decrementare il numero di retry dopo un timeout
// se il numero di retry arriva a 0 si elimina tale nodo
// ritorna false se tale nodo ha 0 retry
// ritorna true se ha un numero di retry maggiori di 0
func decrementNumberOfRetry(id int) bool {
	activeNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			nodesList[i].Retry--
			if nodesList[i].Retry <= 0 {
				activeNodesMutex.Unlock()
				fmt.Printf("[PEER %d] time out expired for node: %d no retry left. Fault node!\n", getMyId(), id)
				updateNodeStateToFault(id)
				return false
			}
			fmt.Printf("[PEER %d] time out expired for node: %d retry left: %d\n", getMyId(), id, nodesList[i].Retry)
			break
		}

	}

	activeNodesMutex.Unlock()
	return true
}

// funzione che reimposta al massimo il numero di retry di un nodo e setta lo stato ad attivo
// chiamata dopo aver ricevuto un heartbeat da un nodo
func resetRetryNumber(id int) {
	activeNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			nodesList[i].Retry = getMaxRetry()
			nodesList[i].State = 1
			break
		}
	}

	activeNodesMutex.Unlock()
}

// funzione che segnala il nodo come fallito e lo rimuove dalla lista
func updateNodeStateToFault(id int) {
	activeNodesMutex.Lock()
	faultNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {

			nodesList[i].State = 2
			//rimuovo il nodo e lo aggiungo ai nodi falliti
			faultNodesList = append(faultNodesList, nodesList[i])
			nodesList = append(nodesList[:i], nodesList[i+1:]...)

			break
		}
	}

	activeNodesMutex.Unlock()
	faultNodesMutex.Unlock()
}

// funzione che elimina un nodo dalla lista dei nodi fault e lo aggiunge alla lista dei nodi fault
func reviveFaultNode(faultId int) {
	faultNodesMutex.Lock()

	for i := 0; i < len(faultNodesList); i++ {
		if faultNodesList[i].ID == faultId {
			activeNodesMutex.Lock()
			nodesList = append(nodesList, faultNodesList[i])
			activeNodesMutex.Unlock()

			faultNodesList = append(faultNodesList[:i], faultNodesList[i+1:]...)

			updateNodeDistance(faultId, 1, -1)
		}
	}

	faultNodesMutex.Unlock()
}

// funzione che ritorna il numero di nodi attivi
func getLenght() int {
	return len(nodesList)
}

// funzione che si attiva qual'ora la lista dei nodi attivi fosse vuota e la lista dei nodi fault piena
// si può verificare questa situazione quando si impostano in modo non appropriato le variabili del file
// di configurazione. Tenta di far rivivere un nodo
func tryLazzarus() {

	if getLazzarusTry() == 0 {
		os.Exit(-1)
	}

	time.Sleep(time.Duration(getLazzarusTime()) * time.Millisecond)

	activeNodesMutex.Lock()
	faultNodesMutex.Lock()

	if len(nodesList) == 0 && len(faultNodesList) > 0 {
		fmt.Printf("[PEER %d] TRYING LAZZARUS PROTOCOL\n", getMyId())
		faults := len(faultNodesList)

		activeNodesMutex.Unlock()
		faultNodesMutex.Unlock()

		for i := faults - 1; i >= 0; i-- {
			reviveFaultNode(faultNodesList[i].ID)
		}
		setLazzarusTry(getLazzarusTry() - 1)
		return
	}

	activeNodesMutex.Unlock()
	faultNodesMutex.Unlock()
}

// funzione per escludere delle comunicazioni tra nodi
func setIgnoreIds() {
	nodes := getIgnoreNodes()

	isMyRule := false

	for i := 0; i < len(nodes); i++ {
		if nodes[i] == getMyId() {
			isMyRule = true
			break
		}
	}

	if isMyRule {
		for i := 0; i < len(nodes); i++ {
			if nodes[i] == getMyId() {
				continue
			}
			ignoreIds = append(ignoreIds, nodes[i])
		}
	} else {
		return
	}
}

// funzione che aggiunge un nodo agli ignoreIds
func addToIgnoreIds(id int) {
	//verifico se sia presente
	ignoreMutex.Lock()
	if !checkIgnoreId(id) {
		ignoreIds = append(ignoreIds, id)
	}
	ignoreMutex.Unlock()
}

// verifica se un nodo è da ignorare o no. ritorna true se è da ignorare
func checkIgnoreId(id int) bool {
	for i := 0; i < len(ignoreIds); i++ {
		if ignoreIds[i] == id {
			return true
		}
	}
	return false
}

// funzione di ausilio che stampa tutti i nodi attivi e fault che il nodo conosce attualmente
// ritorna anche la distanza
func printAllNodeList() {
	activeNodesMutex.Lock()

	fmt.Printf("\n[PEER %d] active nodes\n", getMyId())
	for i := 0; i < len(nodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d  rtt: %d\n", nodesList[i].ID, nodesList[i].State, nodesList[i].ResponseTime)
	}

	if len(nodesList) == 0 {
		fmt.Printf("None\n")
	}

	fmt.Printf("[PEER %d] fault nodes\n", getMyId())

	for i := 0; i < len(faultNodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d\n", faultNodesList[i].ID, faultNodesList[i].State)
	}

	if len(faultNodesList) == 0 {
		fmt.Printf("None")
	}

	fmt.Printf("\n\n")

	activeNodesMutex.Unlock()
}
