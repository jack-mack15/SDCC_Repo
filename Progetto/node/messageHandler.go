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

	message := string(buffer)
	//fmt.Printf("handleUDPMessage()--> message: %s\n", message)

	count := strings.Count(message, "#") + 1
	if count == 0 {
		return
	}

	parts := strings.SplitN(message, "#", count)

	code := parts[0]

	if code == "000" || code == "111" {
		//GESTIONE SEMPLICE HEARTBEAT

		//TODO rimuovere questo blocco di codice che simula un ritardo
		max := 200
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

		if code == "111" {

			//Heatbeat con digest del multicast bimodal
			id := getIdFromMessage(parts[1])

			gossipMessage := parts[3]
			go gossiper.HandleGossipMessage(id, gossipMessage)
			fmt.Println("messaggio gossip ricevuto")
		}

	} else if code == "222" {
		//codice 222 è associato al blind counter rumor mongering in caso il messaggio
		//riporti anche info nodi fault

		id := getIdFromMessage(parts[1])

		//gestisco le info sul nodo mittente se non lo conosco
		handleNodeInfo(parts, remoteUDPAddr)

		go gossiper.HandleGossipMessage(id, parts[3])

	}

	//fmt.Printf("handleUDPMessage() 000 --> tutto ok\n\n")
}

// funzione che recupera info dall'heartbeat ricevuto e verifica se il nodo mittente è presente
// nella lista di nodi conosciuti. In caso lo aggiunge
func handleNodeInfo(parts []string, remoteUDPAddr *net.UDPAddr) {
	//recupero id
	id := getIdFromMessage(parts[1])

	if !CheckPresenceActiveNodesList(id) {

		//se il nodo era fault e si è riattivato, lo elimino dalla lista dei nodi fault
		if CheckPresenceFaultNodesList(id) {
			gossiper.ReviveNode(id)
			reviveFaultNode(id)
		}

		//recupero porta di ascolto, quella con cui il sender invia i messaggi è differente dalla porta di ascolto
		addressString := parts[2]
		addressParts := strings.SplitN(addressString, ":", 2)
		remoteIP := remoteUDPAddr.IP.String()
		remotePort := addressParts[1]

		//questo è l'address "corretto", quello corretto per contattare tale nodo
		remoteAddrStr := remoteIP + ":" + remotePort

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
		_ = AddActiveNode(id, 1, remoteAddrStr, currUDPAddr, currTCPAddr)
	}
}

// funzione che invia gli update per il bimodal multicast gossip
func SendMulticastMessage(message string, remoteUDPAddr *net.UDPAddr) {
	conn, err := net.DialUDP("udp", nil, remoteUDPAddr)
	if err != nil {
		fmt.Println("SendMulticastMessage()--> errore durante la connessione:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("SendMulticastMessage()--> errore durante invio messaggio:", err)
		return
	}
}

// funzione che invia i messaggi per il blind counter rumor mongering
func sendBlindCounterGossipMessage(toNotifyId int, faultId int) {
	remoteAddr := getSelectedUDPAddress(toNotifyId)
	message := writeBlindCounterGossipMessage(faultId)

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		fmt.Println("sendBlindCounterGossipMessage()--> errore durante la connessione:", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("sendBlindCounterGossipMessage()--> errore durante invio messaggio:", err)
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

				//invoco il gossiper poichè ho scoperto un nodo fault
				go gossiper.Gossip(singleNode.ID)

				return
			}
		}

		currDistance := calculateDistance(responseTime)
		UpdateNodeDistance(singleNode.ID, 1, int(responseTime.Milliseconds()), currDistance)
	}
}

// funzione di ausilio che mi trasforma il contenuto di un messaggio di gossip da stringa a array di interi
func extractIdArrayFromMessage(digest string) []int {
	var array []int

	if digest == "" {
		return array
	}

	count := strings.Count(digest, "/") + 1
	arrayElems := strings.SplitN(digest, "/", count)

	for i := 0; i < count; i++ {
		currId, _ := strconv.Atoi(arrayElems[i])
		array = append(array, currId)
	}

	return array
}

// funzione che estrae l'id del sender dal messaggio
func getIdFromMessage(messagePart string) int {

	idParts := strings.SplitN(messagePart, ":", 2)
	idString := idParts[1]

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Printf("getIdFromMessage() 000 --> errore conversione id: %v", err.Error())
	}
	return id
}

// funzione che scrive il messaggio di heartbeat
func writeHeartBeatMessage(id int, port int) string {

	message := "000#id:" + strconv.Itoa(id) + "#port:" + strconv.Itoa(port)
	digest := GetDigest()
	message = message + "#" + digest
	return message
}

// funzione che scrive il messaggio per il bimodal multicast
func writeMulticastGossipMessage(id int, port int, digest string) string {
	message := "111#id:" + strconv.Itoa(id) + "#port:" + strconv.Itoa(port)
	message = message + "#" + digest
	return message
}

// funzione che scrive il messaggio per il blind counter rumor
func writeBlindCounterGossipMessage(faultId int) string {
	message := "222#id:" + strconv.Itoa(GetMyId()) + "#port:" + strconv.Itoa(GetMyPort())
	message = message + "#" + strconv.Itoa(faultId)
	return message
}

// funzione che calcola la distanza del nodo
func calculateDistance(responseTime time.Duration) int {

	//TODO implementare altri tipi di calcoli per la distanza?
	//ottengo la distanza in km
	distance := (responseTime.Milliseconds() * 200) / 2

	return int(distance)
}
