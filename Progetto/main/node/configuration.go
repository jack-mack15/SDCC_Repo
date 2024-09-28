package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// parametro della exponential moving average
// se p = 0, resetto RTT con il nuovo valore
// se p = 0.5, faccio media aritmetica
var p float64

// ID del nodo attuale
var myId int

// numero di porta su cui sono in ascolto
var myPort int

// indirizzo sotto forma di UDPAddr e TCPAddr
var ownUDPAddress *net.UDPAddr
var ownTCPAddress *net.TCPAddr

// RTT default per nodi che non ho mai contattato
var defRTT int

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

// bool che stabilisce se usare la funzione max nell'aggiornamento del tempo di risposta
var usingMax bool

// massimo numero di vicini a cui il nodo corrente inoltra un update
var b int

// massimo numero di volte che un update verrà inoltrato
var f int

// numero di tentativi massimi per la funzione tryLazzarus()
var lazzarusTry int

// intervallo di tempo in secondi tra due lazzarusTry
var lazzarusTime int

func readConfigFile() int {

	//recupero il path del file delle configurazioni
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("readConfigFile()--> errore apertura file:", err)
	}

	exeDir := filepath.Dir(exePath)

	filePath := filepath.Join(exeDir, "node_config.txt")

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("readConfigFile()--> errore nell'apertura del file:", err)
		return 0
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	data := make(map[string]string)

	//leggo riga per riga, ometto i commenti e aggiungo elementi alla map
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Println("readConfigFile()--> formato della linea non valido:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("readConfigFile()--> 3rrore durante la lettura del file:", err)
		return 0
	}

	//lettura p
	p, err = strconv.ParseFloat(data["p"], 64)
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione:", err)
		return 0
	}
	//lettura my port
	myPort, err = strconv.Atoi(data["my_port"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione my_port:", err)
		return 0
	}
	//lettura defRTT
	defRTT, err = strconv.Atoi(data["def_RTT"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione defRTT:", err)
		return 0
	}
	//lettura rttMult
	rttMult, err = strconv.ParseFloat(data["rttMult"], 64)
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione rttMult:", err)
		return 0
	}
	//lettura hb_delay
	hbDelay, err = strconv.Atoi(data["hb_delay"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione hb_delay:", err)
		return 0
	}
	//lettura info service registry
	//sdIP = data["sd_ip"]
	sdPort = 8080
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

	sdPort, err = strconv.Atoi(data["sd_port"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione porta service:", err)
		return 0
	}

	sdRetry, err = strconv.Atoi(data["sd_retry"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione sd_retry:", err)
		return 0
	}
	//lettura max num
	maxNum, err = strconv.Atoi(data["num"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione max num:", err)
		return 0
	}
	//lettura gossiptype
	gossipType, err = strconv.Atoi(data["gossip_type"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione gossipType:", err)
		return 0
	}
	//lettura gossip_interval
	gossipInterval, err = strconv.Atoi(data["gossip_interval"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione gossip_interval:", err)
		return 0
	}
	//lettura b
	b, err = strconv.Atoi(data["max_neighbour"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione max neighbour:", err)
		return 0
	}
	//lettura f
	f, err = strconv.Atoi(data["max_iter"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione di max iter:", err)
		return 0
	}
	//lettura max_retry
	maxRetry, err = strconv.Atoi(data["max_retry"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore nella conversione max_retry:", err)
		return 0
	}
	//lettura using max
	maxFun := data["using_max"]
	if maxFun == "" {
		fmt.Println("readConfigFile()--> errore using_max not set")
	} else if maxFun == "1" {
		usingMax = true
	} else {
		usingMax = false
	}
	//lettura lazzarus_retry
	lazzarusTry, err = strconv.Atoi(data["lazzarus_try"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore converione lazzarus_try")
		return 0
	}
	//lettura lazzarus_time
	lazzarusTime, err = strconv.Atoi(data["lazzarus_time"])
	if err != nil {
		fmt.Println("readConfigFile()--> errore converione lazzarus_time")
		return 0
	}

	check := checkParameters()

	if !check {
		return 0
	} else {
		return 1
	}
}

func checkParameters() bool {

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

	//check b
	if b < 0 {
		fmt.Println("config file error: max neighbour must be a positive int")
		return false
	}

	//check b
	if f < 0 {
		fmt.Println("config file error: max iteration must be a positive int")
		return false
	}

	//check p
	if p < 0.0 || p > 1.0 {
		fmt.Println("config file error: parameter P must be between 0 and 1")
		return false
	}

	//check def_rtt
	if defRTT < 0 || defRTT > 1000 {
		fmt.Println("config file error: parameter Def_RTT must be between 0 and 1000")
		return false
	}

	//check rttMult
	if rttMult <= 0.0 {
		fmt.Println("config file error: parameter rttMult must be a positive float")
		return false
	}

	//check hb_delay
	if hbDelay <= 0 {
		fmt.Println("config file error: parameter hb_delay must be a positive integer")
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

	return true
}
func getMyPort() int {
	return myPort
}
func setOwnUDPAddr(UDPAddr *net.UDPAddr) {
	ownUDPAddress = UDPAddr
}
func setOwnTCPAddr(TCPAddr *net.TCPAddr) {
	ownTCPAddress = TCPAddr
}
func getGossipType() int {
	return gossipType
}
func getP() float64 {
	return p
}
func getOwnUDPAddr() *net.UDPAddr {
	return ownUDPAddress
}
func getOwnTCPAddr() *net.TCPAddr {
	return ownTCPAddress
}
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
	return b
}
func getMaxIter() int {
	return f
}
func getGossipInterval() int {
	return gossipInterval
}
func getMaxRetry() int {
	return maxRetry
}
func getUsingMax() bool {
	return usingMax
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
