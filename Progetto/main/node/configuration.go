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
var rtt_mult int

// delay in secondi tra due serie di heartbeat
var hb_delay int

// indirizzo ip e porta del service discovery
var sdIP string
var sdPort int

// valore che indica quanti nodi può contattare ad ogni iterazione
var maxNum int

// variabile che indica che tipologia di gossip usare
var gossipType int

// massimo numero di vicini a cui il nodo corrente inoltra un update
var b int

// massimo numero di volte che un update verrà inoltrato
var f int

func ReadConfigFile() int {

	//recupero il path del file delle configurazioni
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore apertura file:", err)
	}

	exeDir := filepath.Dir(exePath)

	filePath := filepath.Join(exeDir, "node_config.txt")

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nell'apertura del file:", err)
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
			fmt.Println("ReadConfigFile()--> formato della linea non valido:", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("ReadConfigFile()--> 3rrore durante la lettura del file:", err)
	}

	//lettura p
	p, err = strconv.ParseFloat(data["p"], 64)
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione:", err)
		return 0
	}
	//lettura my port
	myPort, err = strconv.Atoi(data["my_port"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione my_port:", err)
		return 0
	}
	//lettura defRTT
	defRTT, err = strconv.Atoi(data["def_RTT"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione defRTT:", err)
		return 0
	}
	//lettura rtt_mult
	rtt_mult, err = strconv.Atoi(data["rtt_mult"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione rtt_mult:", err)
	}
	//lettura hb_delay
	hb_delay, err = strconv.Atoi(data["hb_delay"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione hb_delay:", err)
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
		fmt.Println("ReadConfigFile()--> errore nella conversione porta service:", err)
		return 0
	}
	//lettura max num
	maxNum, err = strconv.Atoi(data["num"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione max num:", err)
		return 0
	}
	//lettura gossiptype
	gossipType, err = strconv.Atoi(data["gossip_type"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione gossipType:", err)
		return 0
	}
	//lettura b
	b, err = strconv.Atoi(data["max_neighbour"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione max neighbour:", err)
		return 0
	}
	//lettura f
	f, err = strconv.Atoi(data["max_iter"])
	if err != nil {
		fmt.Println("ReadConfigFile()--> errore nella conversione di max iter:", err)
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
	if p < 0 || p > 1 {
		fmt.Println("config file error: parameter P must be between 0 and 1")
		return false
	}

	//check def_rtt
	if defRTT < 0 || defRTT > 1000 {
		fmt.Println("config file error: parameter Def_RTT must be between 0 and 1000")
		return false
	}

	//check rtt_mult
	if rtt_mult <= 0 {
		fmt.Println("config file error: parameter rtt_mult must be a positive integer")
		return false
	}

	//check hb_delay
	if hb_delay <= 0 {
		fmt.Println("config file error: parameter hb_delay must be a positive integer")
	}
	//check maxNum
	if maxNum < 0 {
		fmt.Println("config file error: MaxNum must be greater or equal to -1")
		return false
	}

	//check port number
	if sdPort != 8080 || myPort != 8080 {
		fmt.Println("config file error: please use port 8080 for registry and 8081 for node")
		return false
	}

	return true
}
func GetMyPort() int {
	return myPort
}
func SetOwnUDPAddr(UDPAddr *net.UDPAddr) {
	ownUDPAddress = UDPAddr
}
func SetOwnTCPAddr(TCPAddr *net.TCPAddr) {
	ownTCPAddress = TCPAddr
}
func GetGossipType() int {
	return gossipType
}
func GetP() float64 {
	return p
}
func GetOwnUDPAddr() *net.UDPAddr {
	return ownUDPAddress
}
func GetOwnTCPAddr() *net.TCPAddr {
	return ownTCPAddress
}
func GetMaxNum() int {
	return maxNum
}
func GetSDInfoString() string {
	portStr := strconv.Itoa(sdPort)
	return sdIP + ":" + portStr
}
func SetMyId(id int) {
	myId = id
}
func GetMyId() int {
	return myId
}
func GetDefRTT() int {
	return defRTT
}
func getRttMult() int {
	return rtt_mult
}
func getHBDelay() int {
	return hb_delay
}
func GetMaxNeighbour() int {
	return b
}
func GetMaxIter() int {
	return f
}
