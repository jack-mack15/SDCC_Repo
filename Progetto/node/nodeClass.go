package node

import (
	"fmt"
	"math"
	"math/rand"
	"net"
	"sync"
)

//file che mantiene traccia della lista di nodi attivi della rete

// struttura singolo nodo
type Node struct {
	//ID del nodo assegnato dal service registry
	ID int
	//indirizzo per identificare nodo, tipo puntatore a UDPAddr
	UDPAddr *net.UDPAddr
	//indirizzo per identificare nodo, tipo puntatore a TCPAddr
	TCPAddr *net.TCPAddr
	//indirizzo per identificare nodo, tipo string
	StrAddr string
	//State indica lo stato in cui si trova il nodo: 0 non conosciuto, 1 attivo, 2 sospettato, -1 disattivo
	State int
	//distanza del nodo, -1 indica che non è conosciuto
	Distance int
	//tempo risposta del nodo all'ultimo messaggio, -1 indica che non è conosciuto
	ResponseTime int
}

var nodesList []Node

var faultNodesList []Node

var activeNodesMutex sync.Mutex

var faultNodesMutex sync.Mutex

// funzione che restituisce l'indirizzo UDP di un nodo della lista
func getSelectedUDPAddress(id int) *net.UDPAddr {
	activeNodesMutex.Lock()

	for _, node := range nodesList {
		if node.ID == id {
			activeNodesMutex.Unlock()
			return node.UDPAddr
		}
	}
	activeNodesMutex.Unlock()

	return nil
}

// funzione che verifica se un nodo è presente. ritorna true se è presente, false altrimenti
func CheckPresenceActiveNodesList(id int) bool {
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
func CheckPresenceFaultNodesList(id int) bool {
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
func AddActiveNode(id int, state int, strAddr string, UDPAddr *net.UDPAddr, TCPAddr *net.TCPAddr) bool {
	if !CheckPresenceActiveNodesList(id) {

		currNode := Node{}
		currNode.ID = id
		currNode.StrAddr = strAddr
		currNode.UDPAddr = UDPAddr
		currNode.TCPAddr = TCPAddr
		currNode.State = state
		currNode.Distance = -1
		currNode.ResponseTime = -1

		activeNodesMutex.Lock()
		nodesList = append(nodesList, currNode)
		activeNodesMutex.Unlock()
		return true
	}
	return false
}

// funzione che sceglie i nodi da contattare in base al valore maxNum della configurazione
func GetNodeToContact() []Node {
	//scelta dei nodi da contattare
	actualLen := getLenght()
	howManyToContact := GetMaxNum()
	isRand := true

	if GetMaxNum() == 0 {
		//calcolo rad quadr e arrotondo per eccesso
		sqr := math.Sqrt(float64(actualLen))
		howManyToContact = int(math.Ceil(sqr))
	}
	if GetMaxNum() == -1 {
		//contatto tutti i nodi che conosco
		howManyToContact = actualLen
		isRand = false
	}

	var selectedNode []Node

	if isRand {
		//contatto in modo randomico
		elemToContact := make(map[int]bool)

		activeNodesMutex.Lock()
		lenght := getLenght()
		//genero dei numeri casuali tutti differenti, corrispondono alla scelta di nodi da contattare
		i := 0
		for i < howManyToContact {
			random := rand.Intn(lenght)
			_, ok := elemToContact[random]
			if !ok && nodesList[random].State != 2 {
				elemToContact[random] = true
				selectedNode = append(selectedNode, nodesList[random])
				i++
			} else {
				continue
			}
		}
		activeNodesMutex.Unlock()

	} else {
		//contatto tutti quelli che conosco
		activeNodesMutex.Lock()
		lenght := getLenght()
		for i := 0; i < lenght; i++ {
			selectedNode = append(selectedNode, nodesList[i])
		}
		activeNodesMutex.Unlock()
	}

	return selectedNode

}

// funzione che restituisce tutti i nodi per eseguire il multicast
func GetNodesMulticast() map[int]*net.UDPAddr {

	idMap := make(map[int]*net.UDPAddr)

	activeNodesMutex.Lock()
	lenght := getLenght()
	for i := 0; i < lenght; i++ {
		idMap[nodesList[i].ID] = nodesList[i].UDPAddr
	}
	activeNodesMutex.Unlock()

	return idMap
}

// funzione che restituisce la lista di tutti gli id dei nodi conosciuti
func GetNodesId() []int {
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
func UpdateNode(id int, state int, responseTime int, distance int) {
	activeNodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			nodesList[i].State = state
			p := GetP()
			nodesList[i].Distance = int(p*float64(nodesList[i].Distance) + (1-p)*float64(distance))
			nodesList[i].ResponseTime = responseTime
			break
		}
	}

	activeNodesMutex.Unlock()
}

// funzione che segnala il nodo come fallito e lo rimuove dalla lista
func UpdateNodeState(id int) {
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
			faultNodesList = append(faultNodesList[:i], faultNodesList[i+1:]...)
		}
	}

	faultNodesMutex.Unlock()
}

// funzione che ritorna il numero di nodi attivi
func getLenght() int {
	return len(nodesList)
}

func PrintAllNodeList() {
	activeNodesMutex.Lock()

	fmt.Println("nodi attivi")
	for i := 0; i < len(nodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d  distanza: %d \n", nodesList[i].ID, nodesList[i].State, nodesList[i].Distance)
	}
	fmt.Println("nodi falliti")
	for i := 0; i < len(faultNodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d  distanza: %d \n", faultNodesList[i].ID, faultNodesList[i].State, faultNodesList[i].Distance)
	}
	fmt.Printf("\n\n")
	activeNodesMutex.Unlock()
}
