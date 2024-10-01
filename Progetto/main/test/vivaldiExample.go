package main

import (
	"fmt"
	"math"
	"math/rand/v2"
)

var ce = 0.25
var cc = 0.1

type nodeCoordinate struct {
	x, y, z float64
	error   float64
}

// funzione che inizializza myCoord
func initCoordination(coordinate *nodeCoordinate) {
	coordinate.x = 0.0
	coordinate.y = 0.0
	coordinate.z = 0.0
	coordinate.error = 0.1

}

func vivaldiAlgorithm(localNode *nodeCoordinate, remoteNode *nodeCoordinate, rtt float64) {
	distance := euclideanDistance(localNode, remoteNode)
	error := rtt - distance

	delta := computeAdaptiveTimeStep(localNode, remoteNode, rtt, distance)
	direction := getDirection(localNode, remoteNode, distance)

	//aggiornamento coordinate
	localNode.x += delta * error * direction[0]
	localNode.y += delta * error * direction[1]
	localNode.z += delta * error * direction[2]
	//fmt.Println("aooooo", localNode.x, localNode.y, localNode.z)

}

// fuzione che calcola la distanza euclidea di un nodo
func euclideanDistance(localNode *nodeCoordinate, remoteNode *nodeCoordinate) float64 {
	xDiff := localNode.x - remoteNode.x
	yDiff := localNode.y - remoteNode.y
	zDiff := localNode.z - remoteNode.z
	distance := math.Sqrt(math.Pow(xDiff, 2) + math.Pow(yDiff, 2) + math.Pow(zDiff, 2))
	return distance
}

func computeAdaptiveTimeStep(localNode *nodeCoordinate, remoteNode *nodeCoordinate, rtt float64, distance float64) float64 {

	relatError := math.Abs(rtt-distance) / rtt

	//TODO aggiungere i mutex sui nodi locale e remoto
	weight := localNode.error / (localNode.error + remoteNode.error)
	localNode.error = (localNode.error * (1 - ce*weight)) + (weight * relatError * ce)

	return cc * weight
}

func getDirection(localNode *nodeCoordinate, remoteNode *nodeCoordinate, distance float64) []float64 {

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
		direction = append(direction, (localNode.x-remoteNode.x)/distance)
		direction = append(direction, (localNode.y-remoteNode.y)/distance)
		direction = append(direction, (localNode.z-remoteNode.z)/distance)
		return direction
	}
}

func main() {
	node1 := nodeCoordinate{}
	node2 := nodeCoordinate{}
	node3 := nodeCoordinate{}
	node4 := nodeCoordinate{}
	node5 := nodeCoordinate{}
	initCoordination(&node1)
	initCoordination(&node2)
	initCoordination(&node3)
	initCoordination(&node4)
	initCoordination(&node5)

	for i := 0; i < 200; i++ {
		vivaldiAlgorithm(&node1, &node2, 20)
		vivaldiAlgorithm(&node2, &node1, 20)
		vivaldiAlgorithm(&node1, &node3, 38)
		vivaldiAlgorithm(&node3, &node1, 40)
		vivaldiAlgorithm(&node2, &node3, 22)
		vivaldiAlgorithm(&node3, &node2, 25)

		vivaldiAlgorithm(&node1, &node4, 17)
		vivaldiAlgorithm(&node4, &node1, 19)
		vivaldiAlgorithm(&node1, &node5, 8)
		vivaldiAlgorithm(&node5, &node1, 10)
		vivaldiAlgorithm(&node4, &node5, 31)
		vivaldiAlgorithm(&node5, &node4, 30)
		vivaldiAlgorithm(&node2, &node4, 20)
		vivaldiAlgorithm(&node4, &node2, 20)
		vivaldiAlgorithm(&node2, &node5, 29)
		vivaldiAlgorithm(&node5, &node2, 24)

		vivaldiAlgorithm(&node3, &node4, 24)
		vivaldiAlgorithm(&node4, &node3, 28)
		vivaldiAlgorithm(&node3, &node5, 31)
		vivaldiAlgorithm(&node5, &node3, 35)

	}

	fmt.Printf("node1 ha coordinate: %.2f   %.2f  %.2f \n", node1.x, node1.y, node1.z)
	fmt.Printf("node2 ha coordinate: %.2f   %.2f  %.2f \n", node2.x, node2.y, node2.z)
	fmt.Printf("node3 ha coordinate: %.2f   %.2f  %.2f \n", node3.x, node3.y, node3.z)
	fmt.Printf("node4 ha coordinate: %.2f   %.2f  %.2f \n", node4.x, node4.y, node4.z)
	fmt.Printf("node5 ha coordinate: %.2f   %.2f  %.2f \n", node5.x, node5.y, node5.z)

	fmt.Printf("1/2 dist: giusto is 20 stima %.2f\n", euclideanDistance(&node1, &node2))
	fmt.Printf("2/3 dist: giusto is 25 stima %.2f\n", euclideanDistance(&node2, &node3))
	fmt.Printf("1/3 dist: giusto is 40 stima %.2f\n", euclideanDistance(&node1, &node3))
	fmt.Printf("1/4 dist: giusto is 17 stima %.2f\n", euclideanDistance(&node1, &node4))
	fmt.Printf("2/5 dist: giusto is 20 stima %.2f\n", euclideanDistance(&node2, &node5))
	fmt.Printf("3/5 dist: giusto is 35 stima %.2f\n", euclideanDistance(&node3, &node5))
	fmt.Printf("4/5 dist: giusto is 30 stima %.2f\n", euclideanDistance(&node4, &node5))
	fmt.Printf("1/5 dist: giusto is 10 stima %.2f\n", euclideanDistance(&node1, &node5))
	fmt.Printf("1/4 dist: giusto is 17 stima %.2f\n", euclideanDistance(&node1, &node4))
}
