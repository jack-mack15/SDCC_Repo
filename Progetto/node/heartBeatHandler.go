package node

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// funzione che gestisce i messaggi ricevuti da altri nodi
func HandleUDPMessage(conn *net.UDPConn, remoteUDPAddr *net.UDPAddr, buffer []byte) {

	//TODO verifica il tipo di messaggio
	message := string(buffer)
	//fmt.Printf("handleUDPMessage()--> message: %s\n", message)

	count := strings.Count(message, "#") + 1
	if count == 0 {
		return
	}

	parts := strings.SplitN(message, "#", count)

	code := parts[0]

	if code == "000" {
		//GESTIONE SEMPLICE HEARTBEAT
		//TODO rimuovere questo blocco di codice che simula un ritardo
		max := 201
		if myId == 3 {
			max = 2001
		}
		randomNumber := rand.Intn(max) + 100
		time.Sleep(time.Duration(randomNumber) * time.Millisecond)
		if randomNumber > 300 {
			os.Exit(0)
		}
		//qui finisce il blocco in questione

		//invio risposta
		_, err := conn.WriteToUDP([]byte("hello\n"), remoteUDPAddr)
		if err != nil {
			fmt.Println("handleUDPMessage()--> errore invio risposta:", err)
			return
		}
		//gestisco le info sul nodo mittente
		handleNodeInfo(parts, remoteUDPAddr)

	} else if code == "222" {

		gossipMessage := parts[1]
		gossiper.HandleGossipMessage(gossipMessage)
		fmt.Printf("AYOOOOOOO messaggio sus gestito")

	}

	//fmt.Printf("handleUDPMessage() 000 --> tutto ok\n\n")
}

// funzione che recupera info dall'heartbeat ricevuto e verifica se il nodo mittente è presente
// nella lista di nodi conosciuti. In caso lo aggiunge
func handleNodeInfo(parts []string, remoteUDPAddr *net.UDPAddr) {
	//recupero id
	idSenderString := parts[1]
	idParts := strings.SplitN(idSenderString, ":", 2)
	remoteId := idParts[1]

	//recupero porta di ascolto, quella con cui il sender invia i messaggi è differente dalla porta di ascolto
	addressString := parts[2]
	addressParts := strings.SplitN(addressString, ":", 2)
	remoteIP := remoteUDPAddr.IP.String()
	remotePort := addressParts[1]

	//questo è l'address "corretto", quello corretto per contattare tale nodo
	remoteAddrStr := remoteIP + ":" + remotePort

	id, err := strconv.Atoi(remoteId)
	if err != nil {
		log.Printf("handleNodeInfo() 000 --> errore conversione id: %v", err.Error())
	}

	portInt, err := strconv.Atoi(remotePort)
	if err != nil {
		log.Printf("handleNodeInfo() 000 --> errore conversione porta a int: %v", err.Error())
	}
	currTCPAddr := &net.TCPAddr{IP: net.ParseIP(remoteIP), Port: portInt}
	currUDPAddr := &net.UDPAddr{IP: net.ParseIP(remoteIP), Port: portInt}
	if err != nil {
		log.Printf("handleNodeInfo() 000 ---> errore risoluzione indirizzo remoto %s: %v", remoteAddrStr, err)
	}

	//aggiungo il nodo. se fosse già presente AddActiveNode() non lo aggiunge
	_ = AddActiveNode(id, remoteAddrStr, currUDPAddr, currTCPAddr)
}

// funzione che invia gli update appena si scopre un nodo fault
func SendGossipSignal(message string, remoteUDPAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", nil, remoteUDPAddr)
	if err != nil {
		fmt.Println("sendGossipSignal()--> errore durante la connessione:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("sendGossipSignal()--> errore durante invio messaggio:", err)
		return
	}
}

// funzione che va ad inviare heartbeat ad un nodo
func SendHeartbeat(singleNode Node, myId int, wg *sync.WaitGroup) {

	defer wg.Done()

	if singleNode.State == -1 {
		return
	} else {
		conn, err := net.DialUDP("udp", nil, singleNode.UDPAddr)
		if err != nil {
			fmt.Println("sendHeartBeat()--> errore durante la connessione:", err)
			return
		}
		defer conn.Close()

		precResponseTime := singleNode.ResponseTime
		if precResponseTime == -1 {
			precResponseTime = GetDefRTT()
		}

		startTime := time.Now()

		message := writeHeartBeatMessage(myId, GetOwnUDPAddr().Port)

		timerErr := conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(precResponseTime) * 3))
		if timerErr != nil {
			return
		}

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("sendHeartBeat()--> errore durante l'invio del messaggio:", err)
			return
		}

		//risposta dal nodo contattato
		reader := bufio.NewReader(conn)
		_, err = reader.ReadString('\n')
		//responseTime è di tipo Duration
		responseTime := time.Since(startTime)

		//fmt.Printf("sendHeartBeat()--> responseTime in time: %d \n\n", int(responseTime.Milliseconds()))

		//entro in questo if se il timeout termina prima di ricezione
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Printf("sendHeartBeat()--> time_out scaduto, nodo sospetto id: %d\n", singleNode.ID)

				//cambio dello stato del nodo
				UpdateFailureNode(singleNode.ID)

				//aggiungo il nodo alla lista di nodi falliti se ancora non presente
				if !CheckPresenceDigestList(singleNode.ID) {
					AddOfflineNode(singleNode.ID)
				}

				gossiper.Gossip(singleNode.ID)

				return
			}
		}

		currDistance := calculateDistance(responseTime)
		UpdateNode(singleNode.ID, 1, int(responseTime.Milliseconds()), currDistance)

		//fmt.Printf("sendHeartBeat()--> risposta dal nodo: %s\n", reply)
	}
}

// funzione che scrive il messaggio di heartbeat
func writeHeartBeatMessage(id int, port int) string {

	message := "000#id:" + strconv.Itoa(id) + "#port:" + strconv.Itoa(port)
	digest := GetDigest()
	message = message + "#" + digest
	return message
}

// funzione che calcola la distanza del nodo
func calculateDistance(responseTime time.Duration) int {

	//TODO implementare altri tipi di calcoli per la distanza?
	//ottengo la distanza in km
	distance := (responseTime.Milliseconds() * 200) / 2

	return int(distance)
}
