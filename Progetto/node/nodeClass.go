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

var failedNodesList []Node

var nodesMutex sync.Mutex

// funzione che verifica se un nodo è presente. ritorna true se è presente, false altrimenti
func CheckPresenceNodeList(id int) bool {
	digestMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			digestMutex.Unlock()
			return true
		}
	}

	digestMutex.Unlock()
	return false
}

// funzione che aggiunge un nodo alla lista. ritorna true se il nodo è stato aggiunto, false altrimenti
func AddActiveNode(id int, strAddr string, UDPAddr *net.UDPAddr, TCPAddr *net.TCPAddr) bool {
	if !CheckPresenceNodeList(id) {

		currNode := Node{}
		currNode.ID = id
		currNode.StrAddr = strAddr
		currNode.UDPAddr = UDPAddr
		currNode.TCPAddr = TCPAddr
		currNode.State = 0
		currNode.Distance = -1
		currNode.ResponseTime = -1

		nodesMutex.Lock()
		nodesList = append(nodesList, currNode)
		nodesMutex.Unlock()
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

		nodesMutex.Lock()
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
		nodesMutex.Unlock()

	} else {
		//contatto tutti quelli che conosco
		nodesMutex.Lock()
		lenght := getLenght()
		for i := 0; i < lenght; i++ {
			selectedNode = append(selectedNode, nodesList[i])
		}
		nodesMutex.Unlock()
	}

	return selectedNode

}

// funzione che restituisce tutti i nodi per eseguire il multicast
func GetNodesMulticast() map[int]*net.UDPAddr {

	idMap := make(map[int]*net.UDPAddr)

	nodesMutex.Lock()
	lenght := getLenght()
	for i := 0; i < lenght; i++ {
		idMap[nodesList[i].ID] = nodesList[i].UDPAddr
	}
	nodesMutex.Unlock()

	return idMap
}

// funzione che aggiorna un nodo della lista, aggiorna stato, distanza e tempo di risposta
func UpdateNode(id int, state int, responseTime int, distance int) {
	nodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {
			nodesList[i].State = state
			p := GetP()
			nodesList[i].Distance = int(p*float64(nodesList[i].Distance) + (1-p)*float64(distance))
			nodesList[i].ResponseTime = responseTime
			break
		}
	}

	nodesMutex.Unlock()
}

// funzione che segnala il nodo come fallito
func UpdateFailureNode(id int) {
	nodesMutex.Lock()

	for i := 0; i < len(nodesList); i++ {
		if nodesList[i].ID == id {

			nodesList[i].State = 2
			failedNodesList = append(failedNodesList, nodesList[i])
			nodesList = append(nodesList[:i], nodesList[i+1:]...)

			break
		}
	}

	nodesMutex.Unlock()
}

// funzione che ritorna il numero di nodi attivi
func getLenght() int {
	return len(nodesList)
}

func PrintAllNodeList() {
	nodesMutex.Lock()

	fmt.Println("nodi attivi")
	for i := 0; i < len(nodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d  distanza: %d \n", nodesList[i].ID, nodesList[i].State, nodesList[i].Distance)
	}
	fmt.Println("nodi falliti")
	for i := 0; i < len(failedNodesList); i++ {
		fmt.Printf("nodo id: %d  stato: %d  distanza: %d \n", failedNodesList[i].ID, failedNodesList[i].State, failedNodesList[i].Distance)
	}

	nodesMutex.Unlock()
}
