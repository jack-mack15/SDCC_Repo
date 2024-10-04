package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

// funzione che gestisce i messaggi ricevuti da altri nodi
func handleUDPHandler(conn net.Conn) {

	//ottengo indirizzo del sender
	remoteAddr := conn.RemoteAddr()
	remoteTCPAddr, ok := remoteAddr.(*net.TCPAddr)
	if !ok {
		fmt.Println("handleUDPHandler()--> errore durante la conversione dell'indirizzo locale a TCPAddr")
	}

	decoder := json.NewDecoder(conn)

	//imposto un timeout per la Decode()
	decodeTimeout := time.Duration(getDefRTT()) * time.Millisecond
	decodeErr := conn.SetReadDeadline(time.Now().Add(decodeTimeout))
	if decodeErr != nil {
		fmt.Printf("Errore durante l'impostazione del timeout: %v\n", decodeErr)
		return
	}

	//devo capire quale tipo di messaggio ho ricevuto
	var dummy map[string]interface{}
	decodeErr = decoder.Decode(&dummy)
	if netErr, ok := decodeErr.(net.Error); ok && netErr.Timeout() {
		fmt.Println("handleUDPHandler()--> Decode() timeout scaduto")
		return
	} else if decodeErr == io.EOF {
		fmt.Println("handleUDPHandler()--> connessione chiusa dal client")
		return
	} else if decodeErr != nil {
		log.Printf("handleUDPHandler()--> errore durante decodifica messaggio: %v", decodeErr)
		return
	}

	// Controlla il tipo di messaggio ricevuto
	messageTypeFloat, ok := dummy["code"].(float64)
	if !ok {
		log.Println("handleUDPHandler()--> messaggio ricevuto senza campo 'type'.")
		return
	}
	messageType := int(messageTypeFloat)

	switch messageType {
	case 1:
		var message VivaldiMessage
		err := mapToStruct(dummy, &message)
		if err != nil {
			log.Println("messageHandler()--> errore decodifica vivaldi message")
			return
		}
		vivaldiMessageHandler(conn, remoteTCPAddr, message, false)

	case 3:

		var message GossipMessage
		err := mapToStruct(dummy, &message)
		if err != nil {
			log.Println("messageHandler()--> errore decodifica gossip message")
			return
		}
		gossiper.HandleGossipFaultMessage(message.IdSender, message)
	case 4:
		var message VivaldiMessage
		err := mapToStruct(dummy, &message)
		if err != nil {
			log.Println("messageHandler()--> errore decodifica vivaldi message")
			return
		}
		vivaldiMessageHandler(conn, remoteTCPAddr, message, true)
	default:
		log.Println("messageHandler()--> tipo messaggio non valido")
		return
	}

}

// funzione che mi aiuta nella conversione di struct
func mapToStruct(data map[string]interface{}, v interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

// funzione che recupera info dall'heartbeat ricevuto e verifica se il nodo mittente è presente
// nella lista di nodi conosciuti. In caso lo aggiunge. Se fosse già presente va a resettare il numero di retry
func handleNodeInfo(idSender int, remotePort int, remoteTCPAddr *net.TCPAddr) {
	if !checkPresenceActiveNodesList(idSender) {
		//se il nodo era fault e si è riattivato, lo elimino dalla lista dei nodi fault
		if checkPresenceFaultNodesList(idSender) {
			gossiper.ReviveNode(idSender)
			reviveFaultNode(idSender)
		}

		remoteAddr := remoteTCPAddr.IP.String() + ":" + strconv.Itoa(remotePort)

		//aggiungo il nodo. se fosse già presente addActiveNode() non lo aggiunge
		addActiveNode(idSender, 1, remoteAddr)
		//aggiungo il nodo alla map delle coordinate se non è presente
		addElemToMap(idSender)
	} else {
		//aggiorno numero di retry e stato
		resetRetryNumber(idSender)
	}
}

// funzione che va a gestire correttamente un messaggio vivaldi, invia i messaggi di risposta
func vivaldiMessageHandler(conn net.Conn, remoteTCPAddr *net.TCPAddr, message VivaldiMessage, isDigest bool) {

	//se è presente un digest allora lo vado a gestire
	if isDigest {
		tempMess := writeGossipDigestMessage(message.IdSender, message.Digest)
		go gossiper.HandleGossipFaultMessage(message.IdSender, tempMess)
	}

	//verifico se conosco il nodo, in caso aggiungo tutte le info necessarie
	handleNodeInfo(message.IdSender, message.PortSender, remoteTCPAddr)

	//inviare dati per il sender
	response := writeVivaldiResponse(message.IdSender)

	data, err := json.Marshal(response)

	if err != nil {
		log.Fatalf("vivaldiMessageHandler()--> errore codifica message vivaldi JSON: %v", err)
	}

	// Invia i dati JSON sulla connessione
	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("Errore durante l'invio dei dati: %v\n", err)
		return
	}
	time.Sleep(1 * time.Second)
}

func sendVivaldiMessage(singleNode *node, wg *sync.WaitGroup) {
	defer wg.Done()

	if singleNode.State == -1 {
		return
	}

	//recupero il rtt che mi aspetto per verificare se il nodo è vivo
	precResponseTime := singleNode.ResponseTime
	//in caso sia la prima volta che contatto il nodo uso un rtt di default
	if precResponseTime <= 0 {
		precResponseTime = getDefRTT()
	}

	//preparazione del messaggio da inviare
	message := writeVivaldiMessage()
	data, err := json.Marshal(message)
	if err != nil {
		log.Fatalf("sendVivaldiMessage()--> errore codifica JSON: %v", err)
		return
	}

	fmt.Printf("[PEER %d] sending vivaldi message to: %d\n", getMyId(), singleNode.ID)

	//genero un timeout per la dialtimeout
	timeout := time.Duration(float64(precResponseTime)*getRttMult()) * time.Millisecond

	//inizio a misurare il rtt
	startTime := time.Now()

	// Provo ad instaurare una connessione con timeout
	conn, timeOutErr := net.DialTimeout("tcp", singleNode.Addr, timeout)
	if timeOutErr != nil {
		// Verifica se l'errore è un errore di rete
		if netErr, ok := timeOutErr.(net.Error); ok && netErr.Timeout() {
			fmt.Printf("sendVivaldiMessage()--> timeout scaduto per %d\n", singleNode.ID)
			go gossiper.GossipFault(singleNode.ID)
		} else {
			fmt.Printf("sendVivaldiMessage()--> errore durante la connessione: %v\n", timeOutErr)
		}
		return
	}
	defer conn.Close()

	//prendo il rempo di risposta
	actualRTT := time.Since(startTime).Milliseconds()
	singleNode.ResponseTime = int(actualRTT)

	// Invio messaggio Vivaldi
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalf("sendVivaldiMessage()--> errore invio dati: %v", err)
		return
	}

	//ricezione risposta
	buffer := make([]byte, 16*1024)
	//inizio timer per la Read(), sfrutto il tempo calcolato sopra
	err = conn.SetReadDeadline(time.Now().Add(timeout * 2))
	if err != nil {
		fmt.Printf("sendVivaldiMessage()--> errore durante l'impostazione del timeout: %v\n", err)
		return
	}

	n, readErr := conn.Read(buffer)
	if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
		fmt.Println("sendVivaldiMessage()--> Read() timeout scaduto")
		go gossiper.GossipFault(singleNode.ID)
		return
	} else if readErr != nil {
		log.Printf("sendVivaldiMessage()--> errore ricezione messaggio: %v", readErr)
		return
	}

	//ricostruisco la struttura dati
	var infoMessage VivaldiMessage
	errUnmarshal := json.Unmarshal(buffer[:n], &infoMessage)
	if errUnmarshal != nil {
		log.Println("sendVivaldiMessage()--> errore decodifica vivaldi response: ", errUnmarshal)
		return
	}

	//aggiungo il nodo se non presente
	addCoordinateToMap(infoMessage.IdSender, infoMessage.Coordinates, float64(actualRTT))

	//aggiungo le coordinate di altri nodi alla mia map
	for id, value := range infoMessage.MapCoor {
		fmt.Printf("[PEER %d] info aggiuntive su %d\n", getMyId(), id)
		addCoordinateToMap(id, value, value.LastRTT)
		if checkIgnoreId(id) {
			unreachableHandler(infoMessage.IdSender, id, value)
		}
	}

	//eseguo vivaldi
	vivaldiAlgorithm(infoMessage.IdSender, float64(actualRTT))

	//aggiorno stato del nodo
	updateNodeDistance(infoMessage.IdSender, 1, int(actualRTT))
}

// funzione che invia gli update per il bimodal multicast gossip
func sendMulticastMessage(id int, message []byte, remoteAddr string) {

	//recupero il tempo di risposta attes
	responseTime := getNodeRtt(id)
	//genero un timeout per la dialtimeout
	timeout := time.Duration(float64(responseTime)*getRttMult()) * time.Millisecond

	//provo ad instaurare una connessione con timeout
	conn, timeOutErr := net.DialTimeout("tcp", remoteAddr, timeout)

	if timeOutErr != nil {
		//verico se l'errore è un errore di rete, quindi dovuto al timeout
		if netErr, ok := timeOutErr.(net.Error); ok && netErr.Timeout() {
			fmt.Printf("sendMulticastMessage()--> timeout scaduto per %s\n", remoteAddr)
			go gossiper.GossipFault(id)
		} else {
			fmt.Printf("sendMulticastMessage()--> errore durante la connessione: %v\n", timeOutErr)
		}
		return
	}
	defer conn.Close()

	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("sendMulticastMessage()--> errore durante invio messaggio:", err)
		return
	}
}

// funzione che invia i messaggi per il blind counter rumor mongering
func sendBlindCounterGossipMessage(message []byte, toNotifyId int) {
	//recupero indirizzo
	remoteAddr := getSelectedTCPAddress(toNotifyId)
	//recupero il tempo di risposta attes
	responseTime := getNodeRtt(toNotifyId)
	//genero un timeout per la dialtimeout
	timeout := time.Duration(float64(responseTime)*getRttMult()) * time.Millisecond

	//provo ad instaurare una connessione con timeout
	conn, timeOutErr := net.DialTimeout("tcp", remoteAddr, timeout)

	if timeOutErr != nil {
		//verico se l'errore è un errore di rete, quindi dovuto al timeout
		if netErr, ok := timeOutErr.(net.Error); ok && netErr.Timeout() {
			fmt.Printf("sendMulticastMessage()--> timeout scaduto per %s\n", remoteAddr)
			go gossiper.GossipFault(toNotifyId)
		} else {
			fmt.Printf("sendMulticastMessage()--> errore durante la connessione: %v\n", timeOutErr)
		}
		return
	}
	defer conn.Close()

	_, err := conn.Write(message)
	if err != nil {
		fmt.Println("sendBlindCounterGossipMessage()--> errore durante invio messaggio:", err)
		return
	}
}
