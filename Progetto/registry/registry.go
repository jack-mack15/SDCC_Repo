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

//TODO gestire notifiche di fallimento

type nodeInfo struct {
	id   int
	ip   string
	port int
}

// da settare con file config
var numGoroutines = 5

var mutex sync.Mutex

var nodeList []nodeInfo
var messageList string

func main() {

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("errore durante l'ascolto:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	var wg sync.WaitGroup
	i := 0
	for {

		for i = 0; i < numGoroutines; i++ {

			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("errore nella connessione:", err.Error())
				continue
			}
			wg.Add(1)
			go handleConnection(conn, &wg)
		}
		wg.Wait()

		fmt.Println(messageList)

		i = 0
	}

}

func handleConnection(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()

	//la lettura si può anche omettere
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("errore lettura da client:", err.Error())
		return
	}

	count := strings.Count(message, "#")
	if count == 1 {

		count++
		parts := strings.SplitN(message, "#", count)

		fmt.Printf("il registry ha ricevuto: %s", parts[1])

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
		count++
		//TODO lettura notifica di fallimento
		//uso il messaggio ricevuto per ottenere id fallito

		id := 2
		mutex.Lock()

		lenght := len(nodeList)
		for i := 0; i < lenght; i++ {
			if nodeList[i].id == id {
				nodeList = append(nodeList[:i], nodeList[i+1:]...)
				break
			}
		}

		mutex.Unlock()
	}
}

func addNode(addr string, port int) string {

	check, oldId := checkNodePresence(addr, port)

	if check {
		currMessage := strconv.Itoa(oldId) + messageList + "\n"
		return currMessage
	}
	currNode := nodeInfo{}
	currNode.ip = addr
	currNode.port = port

	mutex.Lock()

	listLen := len(nodeList)
	id := 0
	if listLen == 0 {
		id = 1
	} else {
		id = nodeList[listLen-1].id + 1
	}

	currNode.id = id
	nodeList = append(nodeList, currNode)
	currMessage := strconv.Itoa(id) + messageList + "\n"
	messageList = messageList + "#" + strconv.Itoa(id) + "/" + addr + ":" + strconv.Itoa(port)
	fmt.Println(messageList)
	mutex.Unlock()

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
