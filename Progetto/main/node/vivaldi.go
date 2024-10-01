package main

import (
	"math"
	"math/rand/v2"
	"sync"
)

//file che simula il comportamento dell'algoritmo vivaldi
//è una versione semplificata e con due coordinate

// TODO cambiare questi due valori
var ce float64
var cc float64

// struttura che rappresenta le coordinate di un nodo
type nodeCoordinate struct {
	x, y, z   float64
	error     float64
	nodeMutex sync.Mutex
}

var nodeMapCoordinates map[int]*nodeCoordinate
var myNodeMutex sync.Mutex

var myCoord nodeCoordinate

// funzione che inizializza myCoord
func initMyCoordination() {
	myCoord.x = 0.0
	myCoord.y = 0.0
	myCoord.z = 0.0
	myCoord.error = getDefError()

	nodeMapCoordinates = make(map[int]*nodeCoordinate)
}

func vivaldiAlgorithm(remoteNode *nodeCoordinate, rtt float64) {
	distance := euclideanDistance(remoteNode)
	error := rtt - distance

	delta := computeAdaptiveTimeStep(remoteNode, rtt, distance)
	direction := getDirection(remoteNode, distance)

	//aggiornamento coordinate
	myCoord.x += delta * error * direction[0]
	myCoord.y += delta * error * direction[1]
	myCoord.z += delta * error * direction[2]

}

// fuzione che calcola la distanza euclidea di un nodo
func euclideanDistance(remoteNode *nodeCoordinate) float64 {
	xDiff := myCoord.x - remoteNode.x
	yDiff := myCoord.y - remoteNode.y
	zDiff := myCoord.z - remoteNode.z
	distance := math.Sqrt(math.Pow(xDiff, 2) + math.Pow(yDiff, 2) + math.Pow(zDiff, 2))
	return distance
}

func computeAdaptiveTimeStep(remoteNode *nodeCoordinate, rtt float64, distance float64) float64 {

	relatError := math.Abs(rtt-distance) / rtt

	//TODO aggiungere i mutex sui nodi locale e remoto
	weight := myCoord.error / (myCoord.error + remoteNode.error)
	myCoord.error = (myCoord.error * (1 - ce*weight)) + (weight * relatError * ce)

	return cc * weight
}

func getDirection(remoteNode *nodeCoordinate, distance float64) []float64 {

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
		direction = append(direction, (myCoord.x-remoteNode.x)/distance)
		direction = append(direction, (myCoord.y-remoteNode.y)/distance)
		direction = append(direction, (myCoord.z-remoteNode.z)/distance)
		return direction
	}
}
