package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// nodi da ignorare
var ignoreNode []int

// valore di default per l'errore delle coordinate
var defError float64

// precision weight
var precWeight float64

// scale factor
var scaleFact float64

// ID del nodo attuale
var myId int

// numero di porta su cui sono in ascolto
var myPort int

// RTT default per nodi che non ho mai contattato
var defRTT int

// specifica le info di quanti nodi allegare ad un vivaldi response
var vivaldiPlus int

// valore che indica quanti rtt aspettare che un nodo risponda
var rttMult float64

// delay in secondi tra due serie di heartbeat
var hbDelay int

// indirizzo ip e porta del service discovery
var sdIP string
var sdPort int

var sdRetry int

// valore che indica quanti nodi può contattare ad ogni iterazione
var maxNum int

// variabile che indica che tipologia di gossip usare
var gossipType int

// tempo in secondi tra due iterazioni di blind counter gossip per lo stesso update
var gossipInterval int

// numero massimo di retry prima di segnare un nodo fault
var maxRetry int

// massimo numero di vicini a cui il nodo corrente inoltra un update
var maxNeigh int

// massimo numero di volte che un update verrà inoltrato
var gossipF int

// numero di tentativi massimi per la funzione tryLazzarus()
var lazzarusTry int

// intervallo di tempo in secondi tra due lazzarusTry
var lazzarusTime int

// frequenza di stampa dei risultati
var printCounter int

func readEnvVariable() int {

	//recupero le variabili d'ambiente
	var err error
	//lettura DEFAULT_ERROR
	defError, err = strconv.ParseFloat(os.Getenv("DEFAULT_ERROR"), 64)
	if err != nil {
		fmt.Println("Error reading DEFAULT_ERROR:", err)
		return 0
	}
	//lettura SCALE_FACTOR
	scaleFact, err = strconv.ParseFloat(os.Getenv("SCALE_FACTOR"), 64)
	if err != nil {
		fmt.Println("Error reading SCALE_FACTOR:", err)
		return 0
	}
	//lettura PRECISION_WEIGHT
	precWeight, err = strconv.ParseFloat(os.Getenv("PRECISION_WEIGHT"), 64)
	if err != nil {
		fmt.Println("Error reading PRECISION_WEIGHT:", err)
		return 0
	}
	//lettura DEFAULT_RTT
	defRTT, err = strconv.Atoi(os.Getenv("DEFAULT_RTT"))
	if err != nil {
		fmt.Println("Error reading DEFAULT_RTT:", err)
		return 0
	}
	//lettura RTT_MULTIPLIER
	rttMult, err = strconv.ParseFloat(os.Getenv("RTT_MULTIPLIER"), 64)
	if err != nil {
		fmt.Println("Error reading RTT_MULTIPLIER:", err)
		return 0
	}
	//lettura MESSAGE_INTERVAL
	hbDelay, err = strconv.Atoi(os.Getenv("MESSAGE_INTERVAL"))
	if err != nil {
		fmt.Println("Error reading MESSAGE_INTERVAL:", err)
		return 0
	}
	//lettura VIVALDI_PLUS_INFO
	vivaldiPlus, err = strconv.Atoi(os.Getenv("VIVALDI_PLUS_INFO"))
	if err != nil {
		fmt.Println("Error reading VIVALDI_PLUS_INFO:", err)
		return 0
	}
	//lettura IGNORE_IDS
	tempStr := os.Getenv("IGNORE_IDS")
	if tempStr != "" {
		ignoreNode = extractIdArrayFromMessage(tempStr)
	}
	//lettura NODE_EACH_MESSAGE
	maxNum, err = strconv.Atoi(os.Getenv("NODE_EACH_MESSAGE"))
	if err != nil {
		fmt.Println("Error reading NODE_EACH_MESSAGE:", err)
		return 0
	}
	//lettura GOSSIP_TYPE
	gossipType, err = strconv.Atoi(os.Getenv("GOSSIP_TYPE"))
	if err != nil {
		fmt.Println("Error reading GOSSIP_TYPE:", err)
		return 0
	}
	//lettura GOSSIP_INTERVAL
	gossipInterval, err = strconv.Atoi(os.Getenv("GOSSIP_INTERVAL"))
	if err != nil {
		fmt.Println("Error reading GOSSIP_INTERVAL:", err)
		return 0
	}
	//lettura GOSSIP_MAX_NEIGHBOR
	maxNeigh, err = strconv.Atoi(os.Getenv("GOSSIP_MAX_NEIGHBOR"))
	if err != nil {
		fmt.Println("Error reading GOSSIP_MAX_NEIGHBOR:", err)
		return 0
	}
	//lettura GOSSIP_MAX_ITERATION
	gossipF, err = strconv.Atoi(os.Getenv("GOSSIP_MAX_ITERATION"))
	if err != nil {
		fmt.Println("Error reading GOSSIP_MAX_ITERATION:", err)
		return 0
	}
	//lettura FAULT_MAX_RETRY
	maxRetry, err = strconv.Atoi(os.Getenv("FAULT_MAX_RETRY"))
	if err != nil {
		fmt.Println("Error reading FAULT_MAX_RETRY:", err)
		return 0
	}
	//lettura LAZZARUS_TRY
	lazzarusTry, err = strconv.Atoi(os.Getenv("LAZZARUS_TRY"))
	if err != nil {
		fmt.Println("Error reading LAZZARUS_TRY:", err)
		return 0
	}
	//lettura LAZZARUS_INTERVAL
	lazzarusTime, err = strconv.Atoi(os.Getenv("LAZZARUS_INTERVAL"))
	if err != nil {
		fmt.Println("Error reading LAZZARUS_INTERVAL:", err)
		return 0
	}
	//lettura NODE_PORT
	myPort, err = strconv.Atoi(os.Getenv("NODE_PORT"))
	if err != nil {
		fmt.Println("Error reading NODE_PORT:", err)
		return 0
	}
	//lettura SERVICE_REGISTRY_RETRY
	sdRetry, err = strconv.Atoi(os.Getenv("SERVICE_REGISTRY_RETRY"))
	if err != nil {
		fmt.Println("Error reading SERVICE_REGISTRY_RETRY:", err)
		return 0
	}
	//lettura SERVICE_REGISTRY_PORT
	sdPort, err = strconv.Atoi(os.Getenv("SERVICE_REGISTRY_PORT"))
	if err != nil {
		fmt.Println("Error reading SERVICE_REGISTRY_PORT:", err)
		return 0
	}
	//lettura SERVER_ADDRESS
	tempIP := os.Getenv("SERVER_ADDRESS")
	temps := strings.Split(tempIP, ":")
	if len(temps) != 2 {
		fmt.Println("Formato della stringa non valido")
		return 0
	}
	// Assegna le sottostringhe a variabili
	sdIP = temps[0]
	if sdIP == "" {
		fmt.Println("SERVER_ADDRESS not set")
		return 0
	}
	//lettura ITERATION_PRINT
	printCounter, err = strconv.Atoi(os.Getenv("ITERATION_PRINT"))
	if err != nil {
		fmt.Println("Error reading ITERATION_PRINT:", err)
		return 0
	}

	check := checkParameters()

	if !check {
		return 0
	} else {
		return 1
	}
}

// funzione che verifica se tutti i parametri sono settati correttamente. ritorna true se non ci sono problemi.
func checkParameters() bool {
	//check defError
	if defError <= 0.0 || defError > 1.0 {
		fmt.Println("config file error: defError must be a float between 0.0 a 1.0")
		return false
	}
	//check scale factor
	if scaleFact <= 0.0 || scaleFact > 1.0 {
		fmt.Println("config file error: scale factor must be a float between 0.0 a 1.0")
		return false
	}
	//check prec weight
	if precWeight <= 0.0 || precWeight > 1.0 {
		fmt.Println("config file error: precision weight must be a float between 0.0 a 1.0")
		return false
	}
	//check gossiptype
	if gossipType != 1 && gossipType != 2 {
		fmt.Println("config file error: gossipType must be equal to 1 or 2")
		return false
	}
	//check gossip_interval
	if gossipInterval < 0 {
		fmt.Println("config file error: gossip_interval must be a positive number")
		return false
	}
	//check def_rtt
	if defRTT < 0 || defRTT > 1000 {
		fmt.Println("config file error: parameter Def_RTT must be between 0 and 1000")
		return false
	}
	//check rttMult
	if rttMult < 0.0 {
		fmt.Println("config file error: parameter rttMult must be a positive float")
		return false
	}
	//check hb_delay
	if hbDelay <= 0 {
		fmt.Println("config file error: parameter hb_delay must be a positive integer")
		return false
	}
	//check vivaldi plus
	if vivaldiPlus < 0 {
		fmt.Println("config file error: parameter vivaldiPlus must be a positive integer")
		return false
	}
	//check maxNeigh
	if maxNeigh < 0 {
		fmt.Println("config file error: max neighbour must be a positive int")
		return false
	}
	//check max iter
	if gossipF < 0 {
		fmt.Println("config file error: max iteration must be a positive int")
		return false
	}
	//check maxNum
	if maxNum < 0 {
		fmt.Println("config file error: MaxNum must be greater or equal to -1")
		return false
	}
	//check port number
	if sdPort != 8080 || myPort != 8081 {
		fmt.Println("config file error: please use port 8080 for registry and 8081 for node")
		return false
	}
	//check sd_retry
	if sdRetry < 0 {
		fmt.Println("config file error: sd_retry must be an positive integer")
		return false
	}
	//check max_retry
	if maxRetry <= 0 {
		fmt.Println("config file error: MaxRetry must be an integer bigger than 0")
		return false
	}
	//check lazzarus_try
	if lazzarusTry < 0 {
		fmt.Println("config file error: lazzarus_try must be a positive integer")
		return false
	}
	//check lazzarus_time
	if lazzarusTime <= 0 {
		fmt.Println("config file error: lazzarus_time must be a positive integer")
		return false
	}
	//check iter print
	if printCounter < 0 {
		fmt.Println("config file error: you specified a wrong print counter, using default 10")
		printCounter = 10
	}

	return true
}

func getPrintCounter() int { return printCounter }
func getVivaldiPlus() int  { return vivaldiPlus }
func getIgnoreNodes() []int {
	return ignoreNode
}
func getScaleFact() float64 {
	return scaleFact
}
func getPrecWeight() float64 {
	return precWeight
}
func getMyPort() int {
	return myPort
}
func getGossipType() int {
	return gossipType
}
func getDefError() float64 { return defError }
func getMaxNum() int {
	return maxNum
}
func getSDInfoString() string {
	portStr := strconv.Itoa(sdPort)
	return sdIP + ":" + portStr
}
func getSDRetry() int {
	return sdRetry
}
func setMyId(id int) {
	myId = id
}
func getMyId() int {
	return myId
}
func getDefRTT() int {
	return defRTT
}
func getRttMult() float64 {
	return rttMult
}
func getHBDelay() int {
	return hbDelay
}
func getMaxNeighbour() int {
	return maxNeigh
}
func getMaxIter() int {
	return gossipF
}
func getGossipInterval() int {
	return gossipInterval
}
func getMaxRetry() int {
	return maxRetry
}
func getLazzarusTry() int {
	return lazzarusTry
}
func setLazzarusTry(try int) {
	lazzarusTry = try
}
func getLazzarusTime() int {
	return lazzarusTime
}
