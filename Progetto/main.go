package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"awesomeProject/node"
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	//SET UP del nodo
	err := node.ReadConfigFile()
	if err == 0 {
		fmt.Println("errore nel recupero del file di conf")
		return
	}

	//ottengo un numero di porta da so e ottengo il mio indirizzo
	listener, err2 := net.Listen("tcp", ":0")
	if err2 != nil {
		log.Fatalf("errore numero porta: %v", err)
	}
	myPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	node.SetOwnUDPAddr(&net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort})
	node.SetOwnTCPAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort})

	//contatto il registry
	sdResponseString := node.ContactRegistry(node.GetOwnTCPAddr(), node.GetSDInfoString())

	howMany := extractNodeList(sdResponseString)
	if howMany == 0 {
		//se sono il primo a contattarlo, ritento il contatto fino a che un altro nodo si collega
		//utilizzo un tot massimo di tentativi
		for count := 0; count < 5; count++ {
			if howMany == 0 && count < 5 {
				time.Sleep(5 * time.Second)
				sdResponseString = node.ContactRegistry(node.GetOwnTCPAddr(), node.GetSDInfoString())
				howMany = extractNodeList(sdResponseString)
			} else {
				break
			}
		}
	}

	//FASE ATTIVA

	go receiverHandler()

	time.Sleep(5 * time.Second)

	//TODO cambiare tutto, lasciar fare al nodeClass
	for {
		//scelgo i nodi da contattare
		nodesToContact := node.GetNodeToContact()

		go contactNode(nodesToContact)

		time.Sleep(5 * time.Second)
		node.PrintAllNodeList()

		//TODO in sendHeartBeat() nella deadline ci sta un "* 3" da modificare
		//TODO scelta tra "blind counter rumor mongering" e "bimodal multicast"
		//TODO ad ogni lettura e scrittura aggiungere un timeout
		//TODO sistemare le approssimazioni e il calcolo della distanza e tempo di risposta
		//TODO vedere se rimuovere o meno un nodo che non risponde
		//TODO aggiungere il digest al heartbeat?

		//MEGA TODO aggiungere in tutte le porzioni di codice, gestioni di fallimenti dei nodi contattati

		//THREAD PER LA RICEZIONE
		//invece se ricevo un nodo su un sospetto, aggiorno la mia lista senza rispondere

		/*LOOP
		  -come scelgo i nodi da contattare, mi serve una funzione per questo
		  -come modifico la lista in caso io venga notificato con un sospetto
		  -come diffondo il sospetto
		  -devo computare la distanza con i nodi calcolati

		*/
	}
}

// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	conn, err := net.ListenUDP("udp", node.GetOwnUDPAddr())
	if err != nil {
		fmt.Println("receiverHandler()--> errore creazione listener UDP:", err)
		return
	}
	defer conn.Close()

	for {
		buffer := make([]byte, 128)

		n, remoteUDPAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("receiverHandler()--> errore lettura pacchetto:", err)
			continue
		}

		go handleUDPMessage(conn, remoteUDPAddr, buffer[:n])
	}
}

// funzione che gestisce i messaggi ricevuti da altri nodi
func handleUDPMessage(conn *net.UDPConn, remoteUDPAddr *net.UDPAddr, buffer []byte) {

	//TODO verifica il tipo di messaggio
	message := string(buffer)
	fmt.Printf("handleUDPMessage()--> message: %s\n", message)

	count := strings.Count(message, "#") + 1
	if count == 0 {
		return
	}

	parts := strings.SplitN(message, "#", count)
	code := parts[0]

	if code == "000" {

		//GESTIONE SEMPLICE HEARTBEAT
		//rispondo al nodo che mi ha contattato con il messaggio di risposta attuale

		//TODO rimuovere questo blocco di codice che simula un ritardo
		randomNumber := rand.Intn(201) + 100
		time.Sleep(time.Duration(randomNumber) * time.Millisecond)
		//qui finisce il blocco in questione

		_, err := conn.WriteToUDP([]byte("hello\n"), remoteUDPAddr)
		if err != nil {
			fmt.Println("handleUDPMessage()--> errore invio risposta:", err)
			return
		}

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
			log.Printf("handleUDPMessage() 000 --> errore conversione id: %v", err.Error())
		}

		//currTCPAddr, err := net.ResolveTCPAddr("tcp", remoteAddrStr)
		//currUDPAddr, err := net.ResolveUDPAddr("udp", remoteAddrStr)
		portInt, err := strconv.Atoi(remotePort)
		if err != nil {
			log.Printf("handleUDPMessage() 000 --> errore conversione porta a int: %v", err.Error())
		}
		currTCPAddr := &net.TCPAddr{IP: net.ParseIP(remoteIP), Port: portInt}
		currUDPAddr := &net.UDPAddr{IP: net.ParseIP(remoteIP), Port: portInt}
		if err != nil {
			log.Printf("handleUDPMessage() 000 ---> errore risoluzione indirizzo remoto %s: %v", remoteAddrStr, err)
		}

		//aggiungo il nodo. se fosse già presente AddActiveNode() non lo aggiunge
		_ = node.AddActiveNode(id, remoteAddrStr, currUDPAddr, currTCPAddr)

		fmt.Printf("handleUDPMessage() 000 --> tutto ok\n\n")

	} else if code == "111" {
		//TODO gestione segnalazione

	}
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
	node.SetMyId(myId)

	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if currId == myId {
			continue
		}

		currStrAddr := strings.TrimSpace(currNodeParts[1])
		currTCPAddr, err := net.ResolveTCPAddr("tcp", currStrAddr)
		currUDPAddr, err := net.ResolveUDPAddr("udp", currStrAddr)
		if err != nil {
			log.Printf("extractNodeList()---> errore risoluzione indirizzo remoto %s: %v", currStrAddr, err)
		}

		check := node.AddActiveNode(currId, currStrAddr, currUDPAddr, currTCPAddr)

		if !check {
			nodeCount++
		}
	}

	return nodeCount
}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(selectedNodes []node.Node) {

	//contatto i nodi
	lenght := len(selectedNodes)

	var wg sync.WaitGroup

	for i := 0; i < lenght; i++ {
		wg.Add(1)
		go sendHeartbeat(selectedNodes[i], node.GetMyId(), &wg)
	}
	wg.Wait()
}

// funzione che va ad inviare heartbeat ad un nodo
func sendHeartbeat(singleNode node.Node, myId int, wg *sync.WaitGroup) {

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
			precResponseTime = node.GetDefRTT()
		}

		startTime := time.Now()

		message := writeHeartBeatMessage(myId, node.GetOwnUDPAddr().Port)

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
		reply, err := reader.ReadString('\n')
		//responseTime è di tipo Duration
		responseTime := time.Since(startTime)

		fmt.Printf("sendHeartBeat()--> responseTime in time: %d \n\n", int(responseTime.Milliseconds()))

		//entro in questo if se il timeout termina prima di ricezione
		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Printf("sendHeartBeat()--> time_out scaduto, nodo sospetto id: %d\n", singleNode.ID)

				//cambio dello stato del nodo
				node.UpdateFailureNode(singleNode.ID)

				//aggiungo il nodo alla lista di nodi falliti se ancora non presente
				if !node.CheckPresenceDigestList(singleNode.ID) {
					node.AddOfflineNode(singleNode.ID)
				}

				signalSus(singleNode.ID)

				return
			}
		}

		currDistance := calculateDistance(responseTime)
		node.UpdateNode(singleNode.ID, 1, int(responseTime.Milliseconds()), currDistance)

		fmt.Printf("sendHeartBeat()--> risposta dal nodo: %s\n", reply)
	}
}

// funzione che segnala un sospettato
func signalSus(id int) {
	//TODO segnalazione con gossip dei sospettati
	//TODO tenere conto di quale tipologia di gossip voglio usare
}

// funzione che scrive il messaggio di heartbeat
func writeHeartBeatMessage(id int, port int) string {
	message := "000#id:" + strconv.Itoa(id) + "#port:" + strconv.Itoa(port)
	return message
}

// funzione che calcola la distanza del nodo
func calculateDistance(responseTime time.Duration) int {

	//TODO implementare altri tipi di calcoli per la distanza?
	//ottengo la distanza in km
	distance := (responseTime.Milliseconds() * 200) / 2

	return int(distance)
}
