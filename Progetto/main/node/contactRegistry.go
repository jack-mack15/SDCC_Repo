package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// funzione che prova a contattare il service registry, in caso di fallimento riprova per un numero di
// volte pari a maxRetry
func tryContactRegistry() {

	setMyId(-1)
	//contatto il registry
	sdResponseString := contact(getSDInfoString())

	retry := 0
	howMany := 0
	maxRetry := getSDRetry()
	for {
		howMany = extractNodeList(sdResponseString)
		if howMany > 0 || retry >= maxRetry {
			break
		} else {
			retry++
			time.Sleep(2 * time.Second)

			sdResponseString = contact(getSDInfoString())
		}
	}
}

// funzione che effettua il contatto con il service registry
func contact(addr string) string {

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("errore connessione:", err.Error())
		os.Exit(1)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("contact()--> errore chiusura connessione")
		}
	}(conn)

	message := ""
	if getMyId() == -1 {
		message = "0#hello registry\n"
	} else {
		message = strconv.Itoa(getMyId()) + "#hello registry\n"
	}

	// Invio del messaggio al server
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("errore invio messaggio:", err.Error())
		return ""
	}

	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("erore ricezione risp:", err.Error())
		return ""
	}

	fmt.Printf("risposta server: %s\n\n", reply)

	return reply
}

// funzione che riceve il messaggio di risposta da il service registry, ottiene id del nodo attuale e
// completa la lista dei nodi che conosce il nodo attuale
func extractNodeList(str string) int {
	count := strings.Count(str, "#")
	nodeCount := 0

	//se sono il primo della rete count == 0
	if count == 0 {
		cleanStr := strings.TrimSpace(str)
		temp, _ := strconv.Atoi(cleanStr)
		setMyId(temp)
		return nodeCount
	}

	count++

	parts := strings.SplitN(str, "#", count)

	myId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	setMyId(myId)

	setIgnoreIds()

	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if currId == myId {
			continue
		}
		if checkIgnoreId(currId) {
			continue
		}

		currStrAddr := strings.TrimSpace(currNodeParts[1])
		currStrParts := strings.Split(currStrAddr, ":")
		strIp := currStrParts[0]
		strPort := strconv.Itoa(getMyPort())
		addActiveNode(currId, 0, strIp+":"+strPort)
		addElemToMap(currId)
		nodeCount++

	}

	return nodeCount
}
