package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// funzione che gestisce i messaggi ricevuti da altri nodi
func handleUDPHandler(conn net.Conn) {

	//ottengo indirizzo del sender
	remoteAddr := conn.RemoteAddr()
	remoteTCPAddr, ok := remoteAddr.(*net.TCPAddr)
	if !ok {
		fmt.Println("Errore durante la conversione dell'indirizzo locale a TCPAddr")
	}

	decoder := json.NewDecoder(conn)

	//devo capire quale tipo di messaggio ho ricevuto
	var dummy map[string]interface{}
	err := decoder.Decode(&dummy)
	if err == io.EOF {
		fmt.Println("Connessione chiusa dal client.")
		return
	}
	if err != nil {
		log.Printf("Errore durante la decodifica del messaggio: %v", err)
		return
	}

	// Controlla il tipo di messaggio ricevuto
	messageTypeFloat, ok := dummy["code"].(float64)
	if !ok {
		log.Println("Messaggio ricevuto senza campo 'type'.")
		return
	}
	messageType := int(messageTypeFloat)

	switch messageType {
	case 1:
		fmt.Printf("ENTRO NEL CASE 1\n")
		var message VivaldiMessage
		err := mapToStruct(dummy, &message)
		if err != nil {
			log.Println("messageHandler()--> errore decodifica vivaldi message")
			return
		}
		vivaldiMessageHandler(conn, remoteTCPAddr, message, false)

	case 3:

		fmt.Printf("ENTRO NEL CASE 3\n")
		var message GossipMessage
		err := mapToStruct(dummy, &message)
		if err != nil {
			log.Println("messageHandler()--> errore decodifica gossip message")
			return
		}
		gossiper.HandleGossipFaultMessage(message.IdSender, message)
	case 4:
		fmt.Printf("ENTRO NEL CASE 4\n")
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
	} else {
		//aggiorno numero di retry e stato
		resetRetryNumber(idSender)
	}
}

// funzione che va a gestire correttamente un messaggio vivaldi, invia i messaggi di risposta
func vivaldiMessageHandler(conn net.Conn, remoteTCPAddr *net.TCPAddr, message VivaldiMessage, isDigest bool) {

	//se è presente un digest allora lo vado a gestire
	if isDigest {
		var tempMessage GossipMessage
		tempMessage.Digest = message.Digest
		go gossiper.HandleGossipFaultMessage(message.IdSender, tempMessage)
	}

	//verifico se conosco il nodo, in caso aggiungo tutte le info necessarie
	handleNodeInfo(message.IdSender, message.PortSender, remoteTCPAddr)

	//inviare dati per il sender
	response := writeVivaldiMessage()

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
	//aggiungo un \n
	//_, err = conn.Write([]byte("\n"))
	//if err != nil {
	//	fmt.Printf("Errore durante l'invio di newline: %v\n", err)
	//	return
	//}
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

	//provo ad instaurare una connessione
	conn, err := net.Dial("tcp", singleNode.Addr)
	if err != nil {
		fmt.Printf("sendVivaldiMessage()--> errore durante la connessione: %v\n", err)
		go gossiper.GossipFault(singleNode.ID)
		return
	}
	defer conn.Close()

	//timer per interrompe l'attesa di risposta in caso di problemi
	//timerErr := conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(float64(precResponseTime)*getRttMult())))
	//if timerErr != nil {
	//	log.Printf("sendVivaldiMessage()--> errore durante impostazione timeout ricezione: %v", err)
	//	return
	//}

	// Inizio timer per misurare il RTT
	startTime := time.Now()

	// Invio messaggio Vivaldi
	_, err = conn.Write(data)
	if err != nil {
		log.Fatalf("sendVivaldiMessage()--> errore invio dati: %v", err)
		return
	}

	//ricezione risposta
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("sendVivaldiMessage()--> errore ricezione messaggio: %v", err)
		return
	}

	//resetto il timer se ho ricevuto il messaggio
	//err = conn.SetReadDeadline(time.Time{})
	//if err != nil {
	//	log.Printf("sendVivaldiMessage()--> errore durante reset timeout: %v", err)
	//	return
	//}

	//prendo il rempo di risposta
	actualRTT := time.Since(startTime).Milliseconds()
	singleNode.ResponseTime = int(actualRTT)

	//ricostruisco la struttura dati
	var infoMessage VivaldiMessage
	err = json.Unmarshal(buffer[:n], &infoMessage)
	if err != nil {
		log.Println("sendVivaldiMessage()--> errore decodifica vivaldi response: ", err)
		return
	}

	//aggiungo il nodo se non presente
	addCoordinateToMap(infoMessage.IdSender, infoMessage.Coordinates)

	//eseguo vivaldi
	fmt.Printf("coordinate ricevute: %.2f, %.2f, %.2f\n", infoMessage.Coordinates.X, infoMessage.Coordinates.Y, infoMessage.Coordinates.Z)
	vivaldiAlgorithm(infoMessage.IdSender, float64(actualRTT))

	//aggiorno stato del nodo
	updateNodeDistance(infoMessage.IdSender, 1, int(actualRTT), 1)

	fmt.Printf("sendVivaldiMessage()--> tutto ok\n")
}

// funzione che invia gli update per il bimodal multicast gossip
func sendMulticastMessage(message []byte, remoteAddr string) {

	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		fmt.Println("errore connessione:", err.Error())
		return
	}
	defer conn.Close()

	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("sendMulticastMessage()--> errore durante invio messaggio:", err)
		return
	}
}

// funzione che invia i messaggi per il blind counter rumor mongering
func sendBlindCounterGossipMessage(message []byte, toNotifyId int) {
	remoteAddr := getSelectedTCPAddress(toNotifyId)
	conn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		fmt.Println("errore connessione:", err.Error())
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Errore nella chiusura della connessione: %v", err)
		}
	}()

	_, err = conn.Write(message)
	if err != nil {
		fmt.Println("sendBlindCounterGossipMessage()--> errore durante invio messaggio:", err)
		return
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
