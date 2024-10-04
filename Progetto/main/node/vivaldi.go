package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
)

//file che simula il comportamento dell'algoritmo vivaldi

type coordinates struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Z       float64 `json:"z"`
	Error   float64 `json:"error"`
	LastRTT float64 `json:"last_rtt"`
}

// struttura che rappresenta le coordinate di un nodo
type vivaldiNodeInfo struct {
	node      coordinates
	nodeMutex sync.Mutex
	//state è pari ad 1 se il nodo è reachable; 2 se è unreachable
	state int
}

var unreachableList []*unreachNode
var unreachMutex sync.Mutex

// struttura di ausilio per i nodi unreacheable
type unreachNode struct {
	idNode int
	rttMap map[int]float64
	max    float64
	min    float64
	mutex  sync.Mutex
}

// info dei nodi
var nodeInfoMap map[int]*vivaldiNodeInfo

var mapMutex sync.Mutex
var myNodeMutex sync.Mutex

var myCoord coordinates

// funzione che inizializza myCoord
func initMyCoordination() {
	myCoord.X = 0.0
	myCoord.Y = 0.0
	myCoord.Z = 0.0
	myCoord.Error = getDefError()

	nodeInfoMap = make(map[int]*vivaldiNodeInfo)
}

// semplice implementazione dell'algoritmo di vivaldi
func vivaldiAlgorithm(remoteId int, rtt float64) {

	remoteNode := getNodeCoordinates(remoteId)
	if remoteNode == nil {
		return
	}

	distance := euclideanDistance(remoteNode)
	rttErr := rtt - distance

	delta := computeAdaptiveTimeStep(remoteNode, rtt, distance)
	direction := getDirection(remoteNode, distance)

	//aggiornamento coordinate
	myCoord.X += delta * rttErr * direction[0]
	myCoord.Y += delta * rttErr * direction[1]
	myCoord.Z += delta * rttErr * direction[2]
}

// fuzione che calcola la distanza euclidea di un nodo
func euclideanDistance(remoteNode *vivaldiNodeInfo) float64 {
	myNodeMutex.Lock()
	xDiff := myCoord.X - remoteNode.node.X
	yDiff := myCoord.Y - remoteNode.node.Y
	zDiff := myCoord.Z - remoteNode.node.Z
	myNodeMutex.Unlock()
	distance := math.Sqrt(math.Pow(xDiff, 2) + math.Pow(yDiff, 2) + math.Pow(zDiff, 2))
	return distance
}

// funzione che calcola il delta dell'algoritmo di vivaldi, inoltre internamente aggiorna il mio errore
func computeAdaptiveTimeStep(remoteNode *vivaldiNodeInfo, rtt float64, distance float64) float64 {

	relatError := 0.0

	if distance == 0 && rtt == 0 {
		relatError = 1.0
	} else if rtt == 0 {
		rtt = 10
	} else {
		relatError = math.Abs(rtt-distance) / rtt
	}

	myNodeMutex.Lock()
	remoteNode.nodeMutex.Lock()

	weight := myCoord.Error / (myCoord.Error + remoteNode.node.Error)
	myCoord.Error = (myCoord.Error * (1 - getPrecWeight()*weight)) + (weight * relatError * getPrecWeight())

	remoteNode.nodeMutex.Unlock()
	myNodeMutex.Unlock()

	return getScaleFact() * weight
}

// funzione che calcola la direzione dell'aggiornamento delle coordinate
func getDirection(remoteNode *vivaldiNodeInfo, distance float64) []float64 {

	var direction []float64

	//se la distanza di due nodi è 0, ovvero ancora inizializzati
	if distance == 0 {
		xRand := rand.Float64()
		yRand := rand.Float64()
		zRand := rand.Float64()
		norm := math.Sqrt(math.Pow(xRand, 2) + math.Pow(yRand, 2) + math.Pow(zRand, 2))
		direction = append(direction, xRand/norm)
		direction = append(direction, yRand/norm)
		direction = append(direction, zRand/norm)
		return direction
	} else {
		direction = append(direction, (myCoord.X-remoteNode.node.X)/distance)
		direction = append(direction, (myCoord.Y-remoteNode.node.Y)/distance)
		direction = append(direction, (myCoord.Z-remoteNode.node.Z)/distance)
		return direction
	}
}

// funzione di ausilio che aggiunge un elemento alla map se non è presente, altrimenti aggiorna le coordianet
func addCoordinateToMap(remoteId int, remoteCoor coordinates, rtt float64) {

	mapMutex.Lock()

	//se il nodo è già presente nella map
	if elem, ok := nodeInfoMap[remoteId]; ok {
		elem.nodeMutex.Lock()
		elem.node = remoteCoor
		elem.node.LastRTT = rtt
		elem.nodeMutex.Unlock()
		mapMutex.Unlock()
		return
	} else {

		isDead := checkPresenceFaultNodesList(remoteId)
		isAlive := checkPresenceActiveNodesList(remoteId)
		var newNode vivaldiNodeInfo
		newNode.node.X = remoteCoor.X
		newNode.node.Y = remoteCoor.Y
		newNode.node.Z = remoteCoor.Z
		newNode.node.LastRTT = rtt
		newNode.node.Error = remoteCoor.Error
		if checkIgnoreId(remoteId) {
			newNode.state = 2
			addUnreachToList(remoteId)
		} else if isDead || isAlive {
			newNode.state = 1
		} else {
			//se ho ricevuto info di un nodo non reachable che non conoscevo prima
			newNode.state = 2
			addUnreachToList(remoteId)
			addToIgnoreIds(remoteId)
		}
		nodeInfoMap[remoteId] = &newNode
		mapMutex.Unlock()
		return
	}
}

// funzione che aggiunge un nodo unreable se non è presente
func addUnreachToList(id int) {
	unreachMutex.Lock()

	for i := 0; i < len(unreachableList); i++ {
		if unreachableList[i].idNode == id {
			unreachMutex.Unlock()
			return
		}
	}

	unreachableList = append(unreachableList, &unreachNode{idNode: id, max: float64(getDefRTT()), min: 0.0})

	unreachMutex.Unlock()

}

// funzione che va a gestire un nodo unreachable
func unreachableHandler(idSender int, unreachId int, nodeInfo coordinates) {
	unreachMutex.Lock()

	var currUnreachNode *unreachNode
	//ottengo la struct del nodo unreachable corrente
	for i := 0; i < len(unreachableList); i++ {
		if unreachableList[i].idNode == unreachId {
			currUnreachNode = unreachableList[i]
			break
		}
	}
	//se non l'ho trovata la aggiungo e riavvio l'algoritmo
	if currUnreachNode == nil {
		unreachMutex.Unlock()
		addUnreachToList(unreachId)
		unreachableHandler(idSender, unreachId, nodeInfo)
		return
	}

	currUnreachNode.mutex.Lock()
	unreachMutex.Unlock()

	//questo è il rtt con il nodo conosciuto
	knowRTT := float64(getNodeRtt(idSender))

	//questo è il rtt che il nodo conosciuto ha misurato con il nodo unreachable
	indirectRTT := nodeInfo.LastRTT

	//somma dei due valori, massimo per la disuguaglianza lati triangolo
	currMax := knowRTT + indirectRTT - 1.0

	//aggiorno il massimo per il rtt del nodo unreachable
	if currMax < currUnreachNode.max {
		currUnreachNode.max = currMax
	}
	//adesso lavoro sul minimo per il rtt
	currMin := math.Abs(knowRTT-indirectRTT) + 1.0
	//aggiorno il minimo per il rtt del nodo unreachable
	if currMin > currUnreachNode.min {
		currUnreachNode.min = currMin
	}
	currUnreachNode.mutex.Unlock()

	//calcolo il valore medio e avvio vivaldi con tale rtt
	rttMean := (currUnreachNode.max + currUnreachNode.min) / 2

	vivaldiAlgorithm(unreachId, rttMean)

}

// funzione che aggiunge un elemento alla map
func addElemToMap(id int) {

	mapMutex.Lock()
	//se esiste non faccio nulla
	if value, ok := nodeInfoMap[id]; ok {
		value.state = 1
		mapMutex.Unlock()
		return
	} else {
		//entro qui se non esiste e lo inizializzo
		var vivNode vivaldiNodeInfo
		if checkIgnoreId(id) {
			vivNode.state = 2
		} else {
			vivNode.state = 1
		}
		var tempCoor coordinates
		tempCoor.X = 0.0
		tempCoor.Y = 0.0
		tempCoor.Z = 0.0
		tempCoor.Error = getDefError()
		vivNode.node = tempCoor

		nodeInfoMap[id] = &vivNode
	}
	mapMutex.Unlock()

}

// funzione che restituisce n nodi randomici le cui informazioni vengono aggiunte al messaggio di risposta vivaldi
func getRandomNodes(idSender int) map[int]coordinates {

	ids := selectRandomNodes(idSender)
	//length := len(ids)
	mapMutex.Lock()
	var selected map[int]coordinates
	selected = make(map[int]coordinates)

	for i := 0; i < len(ids); i++ {
		if elem, ok := nodeInfoMap[ids[i]]; ok {
			var temp coordinates
			temp.X = elem.node.X
			temp.Y = elem.node.Y
			temp.Z = elem.node.Z
			temp.LastRTT = elem.node.LastRTT
			temp.Error = elem.node.Error
			selected[ids[i]] = temp
		} else {
			continue
		}
	}
	mapMutex.Unlock()
	return selected
}

// funzione che restituisce n id random della mappa
func selectRandomNodes(idSender int) []int {
	num := getVivaldiPlus()

	var selectedIds []int

	mapMutex.Lock()

	lenght := len(nodeInfoMap)
	//se il numero scelto è maggiore dei nodi nella mappa
	if num > lenght {
		for k := range nodeInfoMap {
			if k == idSender {
				continue
			}
			selectedIds = append(selectedIds, k)
		}
		mapMutex.Unlock()
		return selectedIds
	}

	//estraggo tutti gli id
	ids := make([]int, 0, lenght)
	for k := range nodeInfoMap {
		if k == idSender {
			continue
		}
		ids = append(ids, k)
	}

	mapMutex.Unlock()

	//mescolo tutti gli id
	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

	selectedIds = ids[:num]

	return selectedIds
}

// funzione che restituisce le coordinate di un nodo
func getNodeCoordinates(id int) *vivaldiNodeInfo {
	mapMutex.Lock()

	//controllo presenza
	if elem, ok := nodeInfoMap[id]; ok {
		mapMutex.Unlock()
		return elem
	}

	mapMutex.Unlock()
	return nil
}

// funzione che restituisce le mie coordinate
func getMyCoordinate() coordinates {
	return myCoord
}

// funzione che stampa tutte le info per verificare il funzionamneto del sistema
func printAllCoordinates() {
	mapMutex.Lock()

	fmt.Printf("my coord, coordinate: %.2f, %.2f, %.2f\n", myCoord.X, myCoord.Y, myCoord.Z)
	if len(nodeInfoMap) > 0 {
		fmt.Printf("[PEER %d] NODES COORDINATES\n", getMyId())
		for key, value := range nodeInfoMap {
			if value.state == 1 {
				fmt.Printf("nodo: %d, state: %d coordinate: %.2f, %.2f, %.2f\n", key, value.state, value.node.X, value.node.Y, value.node.Z)
			} else {
				fmt.Printf("nodo unreachable: %d, state: %d coordinate: %.2f, %.2f, %.2f\n", key, value.state, value.node.X, value.node.Y, value.node.Z)
			}
		}
	}

	fmt.Println()

	fmt.Printf("[PEER %d] DISTANCES\n", getMyId())
	for key, value := range nodeInfoMap {
		rttEst := euclideanDistance(value)
		if value.state == 1 {
			fmt.Printf("nodo: %d, rtt misurato: %d  rtt calcolato: %.2f\n", key, getNodeRtt(key), rttEst)
		} else {
			fmt.Printf("nodo non reachable: %d, rtt artificiale: %.2f rtt calcolato (con coordinate): %.2f\n", key, value.node.LastRTT, rttEst)
		}
	}

	fmt.Println()

	mapMutex.Unlock()
}
