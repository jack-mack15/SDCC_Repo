package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type nodeInfo struct {
	id   int
	ip   string
	port int
}

var listMutex sync.Mutex
var messageListMutex sync.Mutex

var nodeList []nodeInfo
var messageList string

func main() {

	fmt.Println("listening")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("errore durante l'ascolto:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	for {

		fmt.Println("waiting node")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("errore nella connessione:", err.Error())
			continue
		}
		go handleConnection(conn)

		fmt.Println(messageList)

	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	//la lettura si può anche omettere
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("errore lettura da client:", err.Error())
		return
	}

	count := strings.Count(message, "#") + 1

	parts := strings.SplitN(message, "#", count)
	senderID := parts[0]
	fmt.Printf("il registry ha ricevuto: %s", message)

	if senderID == "0" {
		//recupero indirizzo del nodo e numero porta
		clientAddr := conn.RemoteAddr().String()
		parts = strings.SplitN(clientAddr, ":", 2)
		if len(parts) != 2 {
			fmt.Println("formato della linea non valido:", clientAddr)
		}
		address := strings.TrimSpace(parts[0])

		port, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			fmt.Println("errore nella conversione:", err)
			return
		}

		//aggiungo il nuovo nodo alla lista di nodi e aggiorno il messaggio di risposta
		currMessage := addNode(address, port)

		//rispondo al nodo che mi ha contattato con il messaggio di risposta attuale
		_, err = conn.Write([]byte(currMessage))
		if err != nil {
			fmt.Println("errore invio risp:", err.Error())
			return
		}
		fmt.Printf("tutto ok\n\n")
	} else {
		//rispondo al nodo che mi ha contattato con il messaggio di risposta attuale
		messageListMutex.Lock()
		currMessage := senderID + messageList + "\n"
		messageListMutex.Unlock()
		_, err = conn.Write([]byte(currMessage))
		if err != nil {
			fmt.Println("errore invio risp:", err.Error())
			return
		}
		fmt.Printf("tutto ok\n\n")
	}

}

func addNode(addr string, port int) string {

	listMutex.Lock()

	check, oldId := checkNodePresence(addr, port)

	if check {
		currMessage := strconv.Itoa(oldId) + messageList + "\n"
		listMutex.Unlock()
		return currMessage
	}
	currNode := nodeInfo{}
	currNode.ip = addr
	currNode.port = port

	listLen := len(nodeList)
	id := 0
	if listLen == 0 {
		id = 1
	} else {
		id = nodeList[listLen-1].id + 1
	}

	currNode.id = id
	nodeList = append(nodeList, currNode)
	listMutex.Unlock()

	messageListMutex.Lock()
	currMessage := strconv.Itoa(id) + messageList + "\n"
	messageList = messageList + "#" + strconv.Itoa(id) + "/" + addr + ":" + strconv.Itoa(port)
	messageListMutex.Unlock()

	return currMessage

}

// controllo se il nodo è già presente nei nodi della lista
func checkNodePresence(addr string, port int) (bool, int) {
	if len(nodeList) == 0 {
		return false, 0
	}

	nodeListLen := len(nodeList)
	for i := 0; i < nodeListLen; i++ {
		if nodeList[i].port == port && nodeList[i].ip == addr {
			return true, nodeList[i].id
		}
	}
	return false, 0
}
