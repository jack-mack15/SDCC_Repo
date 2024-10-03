package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sync"
)

//file che simula il comportamento dell'algoritmo vivaldi

type coordinates struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Error float64 `json:"error"`
}

// struttura che rappresenta le coordinate di un nodo
type vivaldiNodeInfo struct {
	node      coordinates
	nodeMutex sync.Mutex
}

var nodeMapCoordinates map[int]*vivaldiNodeInfo
var mapMutex sync.Mutex
var myNodeMutex sync.Mutex

var myCoord coordinates

// funzione che inizializza myCoord
func initMyCoordination() {
	myCoord.X = 0.0
	myCoord.Y = 0.0
	myCoord.Z = 0.0
	myCoord.Error = getDefError()

	nodeMapCoordinates = make(map[int]*vivaldiNodeInfo)
}

func vivaldiAlgorithm(remoteId int, rtt float64) {

	remoteNode := getNodeCoordinates(remoteId)
	if remoteNode == nil {
		return
	}

	distance := euclideanDistance(remoteNode)
	error := rtt - distance

	delta := computeAdaptiveTimeStep(remoteNode, rtt, distance)
	direction := getDirection(remoteNode, distance)

	//aggiornamento coordinate
	myCoord.X += delta * error * direction[0]
	myCoord.Y += delta * error * direction[1]
	myCoord.Z += delta * error * direction[2]
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

func computeAdaptiveTimeStep(remoteNode *vivaldiNodeInfo, rtt float64, distance float64) float64 {

	relatError := 0.0

	if distance == 0 && rtt == 0 {
		relatError = 1.0
	} else if rtt == 0 {
		rtt = 10
	} else {
		relatError = math.Abs(rtt-distance) / rtt
	}

	//TODO aggiungere i mutex sui nodi locale e remoto
	weight := myCoord.Error / (myCoord.Error + remoteNode.node.Error)
	myCoord.Error = (myCoord.Error * (1 - getPrecWeight()*weight)) + (weight * relatError * getPrecWeight())

	return getScaleFact() * weight
}

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

func addCoordinateToMap(remoteId int, remoteCoor coordinates) {
	mapMutex.Lock()
	//se il nodo è già presente
	if elem, ok := nodeMapCoordinates[remoteId]; ok {
		elem.nodeMutex.Lock()
		elem.node = remoteCoor
		elem.nodeMutex.Unlock()
		mapMutex.Unlock()
		return
	} else {
		// se il nodo non è presente
		var newNode vivaldiNodeInfo
		newNode.node = remoteCoor
		nodeMapCoordinates[remoteId] = &newNode
		mapMutex.Unlock()
	}
}

func removeNode(remoteId int) {
	mapMutex.Lock()
	if _, ok := nodeMapCoordinates[remoteId]; ok {
		delete(nodeMapCoordinates, remoteId)
	}
	mapMutex.Unlock()
}

func getNodeCoordinates(id int) *vivaldiNodeInfo {
	mapMutex.Lock()

	if elem, ok := nodeMapCoordinates[id]; ok {
		mapMutex.Unlock()
		return elem
	}

	mapMutex.Unlock()
	return nil
}

func getMyCoordinate() coordinates {
	return myCoord
}

func printAllCoordinates() {
	mapMutex.Lock()

	fmt.Printf("my coord, coordinate: %.2f, %.2f, %.2f\n", myCoord.X, myCoord.Y, myCoord.Z)
	for key, value := range nodeMapCoordinates {
		fmt.Printf("nodo: %d, coordinate: %.2f, %.2f, %.2f\n", key, value.node.X, value.node.Y, value.node.Z)
	}

	fmt.Println()

	mapMutex.Unlock()
}
