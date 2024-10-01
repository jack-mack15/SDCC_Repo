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

	//contatto il registry
	sdResponseString := contact(getOwnTCPAddr(), getSDInfoString())

	retry := 0
	howMany := 0
	maxRetry := getSDRetry()
	for {
		howMany = extractNodeList(sdResponseString)
		if howMany > 0 || retry >= maxRetry {
			break
		} else {
			retry++
			time.Sleep(5 * time.Second)

			sdResponseString = contact(getOwnTCPAddr(), getSDInfoString())
		}
	}
}

// funzione che effettua il contatto con il service registry
func contact(localAddr *net.TCPAddr, addr string) string {

	//ottengo l'indirizzo del service registry
	remoteAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Printf("errore ottenimento indirizzo di %s: %v", addr, err)
	}

	conn, err := net.DialTCP("tcp", localAddr, remoteAddr)
	if err != nil {
		fmt.Println("errore connessione:", err.Error())
		os.Exit(1)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Errore nella chiusura della connessione: %v", err)
		}
	}()

	message := "0#hello registry\n"

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
		return nodeCount
	}

	count++

	parts := strings.SplitN(str, "#", count)
	myId, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
	setMyId(myId)

	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if currId == myId {
			continue
		}

		currStrAddr := strings.TrimSpace(currNodeParts[1])
		currUDPAddr, err := net.ResolveUDPAddr("udp", currStrAddr)
		if err != nil {
			log.Printf("extractNodeList()---> errore risoluzione indirizzo remoto %s: %v", currStrAddr, err)
		}

		addActiveNode(currId, 0, currUDPAddr)
		nodeCount++

	}

	return nodeCount
}
