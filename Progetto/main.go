package main

//nella shell dove si lancia il codice go eseguire prima:  export GODEBUG=netdns=go
import (
	"awesomeProject/node"
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// parametro della exponential moving average
// se p = 0, resetto RTT con il nuovo valore
// se p = 0.5, faccio media aritmetica
var p float64

// id del nodo attuale
var my_id int

// indirizzo sotto forma di UDPAddr e TCPAddr
var ownUDPAddress *net.UDPAddr
var ownTCPAddress *net.TCPAddr

// RTT default per nodi che non ho mai contattato
var def_RTT int

// indirizzo ip e porta del service discovery
var sd_ip string
var sd_port int

// valore che indica quanti nodi può contattare ad ogni iterazione
var max_num int

// struttura singolo nodo
type Node struct {
	//id del nodo assegnato dal service registry
	id int
	//indirizzo per identificare nodo, tipo puntatore a UDPAddr
	UDPAddr *net.UDPAddr
	//indirizzo per identificare nodo, tipo puntatore a TCPAddr
	TCPAddr *net.TCPAddr
	//indirizzo per identificare nodo, tipo string
	strAddr string
	//state indica lo stato in cui si trova il nodo: 0 non conosciuto, 1 attivo, 2 sospettato, -1 disattivo
	state int
	//contatore per il tempo, da rimuovere
	counter int
	//distanza del nodo
	distance int
}

// funzione che calcola il RTT
// RTT = p*prec + (1-p)*curr
func calculateRTT(node Node) (float64, error) {
	//ottengo indirizzo ip
	address := fmt.Sprintf("%s", node.UDPAddr)

	start := time.Now()
	//creo connessione tcp
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	elapsed := time.Since(start)

	//ottengo il valore di RTT in millisec
	rtt := float64(elapsed.Microseconds()) / 1000.0
	return rtt, nil
}

// lista di tutti i nodi della rete
var nodes []Node

// mutex della lista dei nodi
var nodesMutex sync.Mutex

func main() {

	//SET UP del nodo
	conf, err := node.ReadConfigFile()
	if err == 0 {
		fmt.Println("errore nel recupero del file di conf")
		return
	}
	p = conf.P
	def_RTT = conf.Def_RTT
	sd_ip = conf.Sd_ip
	sd_port = conf.Sd_port
	max_num = conf.MaxNum

	//ottengo un numero di porta da so e ottengo il mio indirizzo
	listener, err2 := net.Listen("tcp", ":0")
	if err2 != nil {
		log.Fatalf("errore numero porta: %v", err)
	}
	myPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	ownUDPAddress = &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort}
	ownTCPAddress = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: myPort}

	//contatto il registry
	list := node.ContactRegistry(ownTCPAddress, sd_ip+":"+strconv.Itoa(sd_port))
	fmt.Println(len(list))
	extractNodeList(list)

	//se sono il primo a contattarlo, ritento il contatto fino a che un altro nodo si collega
	//utilizzo un tot massimo di tentativi
	for count := 0; count < 5; count++ {
		if len(nodes) == 0 && count < 5 {
			time.Sleep(5 * time.Second)
			list = node.ContactRegistry(ownTCPAddress, sd_ip+":"+strconv.Itoa(sd_port))
			extractNodeList(list)
		} else {
			break
		}
	}

	//TODO eliminare questo if
	if len(nodes) == 0 {
		time.Sleep(500 * time.Second)
	}

	//FASE ATTIVA

	go receiverHandler()

	time.Sleep(5 * time.Second)
	for {
		//scelta dei nodi da contattare
		actualLen := len(nodes)
		nodeToContact := max_num
		isRand := true

		if max_num == 0 {
			//calcolo rad quadr e arrotondo per eccesso
			sqr := math.Sqrt(float64(actualLen))
			nodeToContact = int(math.Ceil(sqr))
		}
		if max_num == -1 {
			//contatto tutti i nodi che conosco
			nodeToContact = actualLen
			isRand = false
		}

		go contactNode(nodeToContact, isRand)
		time.Sleep(5 * time.Second)
		fmt.Println("succede qualcosa?")

		//TODO scelta tra "blind counter rumor mongering" e "bimodal multicast"
		//TODO contattare i nodi per calcolo distanza e vedere se sono vivi
		//TODO contattare i nodi tramite UDP e contattare il server tramite TCP
		//TODO ad ogni lettura e scrittura aggiungere un timeout

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

	conn, err := net.ListenUDP("udp", ownUDPAddress)
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

		_, err := conn.WriteToUDP([]byte("hello\n"), remoteUDPAddr)
		if err != nil {
			fmt.Println("handleUDPMessage()--> errore invio risposta:", err)
			return
		}

		//recupero indirizzo del nodo
		remoteAddr := remoteUDPAddr.IP.String()

		//recupero id e porta dal contenuto del messaggio
		idSenderString := parts[1]
		portSenderString := parts[2]
		idParts := strings.SplitN(idSenderString, ":", 2)
		portParts := strings.SplitN(portSenderString, ":", 2)
		remoteId := idParts[1]
		remotePort := portParts[1]

		id, err := strconv.Atoi(remoteId)
		if err != nil {
			log.Printf("handleUDPMessage() 000 --> errore conversione id: %v", err.Error())
		}

		fmt.Printf("handleUDPMessage() 000 --> id: %d address: %s:%s\n", id, remoteAddr, remotePort)

		//controllo se il nodo è già presente nella lista
		//in caso non lo fosse lo aggiungo alla lista
		nodesMutex.Lock()
		length := len(nodes)
		check := false

		for i := 0; i < length; i++ {
			if nodes[i].id == id {
				check = true
				break
			}
		}
		nodesMutex.Unlock()

		if !check {
			addNode(id, remotePort, remoteAddr)
		}

		fmt.Printf("handleUDPMessage() 000 --> tutto ok\n\n")

	} else if code == "111" {
		//TODO gestione segnalazione

	}
}

// funzione che aggiunge un nuovo nodo alla lista in modo concorrente
func addNode(id int, port string, address string) {

	currNode := Node{}

	currNode.id = id
	currNode.state = 1
	currNode.strAddr = address + ":" + port

	remoteUDPAddr, err := net.ResolveUDPAddr("udp", currNode.strAddr)
	remoteTCPAddr, err := net.ResolveTCPAddr("tcp", currNode.strAddr)
	if err != nil {
		currNode.UDPAddr = nil
		log.Printf("addNode()---> errore ottenimento indirizzo di %s: %v", currNode.strAddr, err)
	} else {
		currNode.UDPAddr = remoteUDPAddr
		currNode.TCPAddr = remoteTCPAddr
	}

	nodesMutex.Lock()

	nodes = append(nodes, currNode)

	nodesMutex.Unlock()

}

// funzione che riceve il messaggio di risposta da il service registry, ottiene id del nodo attuale e
// completa la lista dei nodi che conosce il nodo attuale
func extractNodeList(str string) {
	count := strings.Count(str, "#")

	//se sono il primo della rete count == 0
	if count == 0 {
		return
	}

	count++

	parts := strings.SplitN(str, "#", count)
	my_id, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
	for i := 1; i < count; i++ {

		currNodeInfo := strings.TrimSpace(parts[i])
		currNodeParts := strings.Split(currNodeInfo, "/")

		currNode := Node{}
		tempId, _ := strconv.Atoi(strings.TrimSpace(currNodeParts[0]))

		//se il corrente id corrisponde al mio id, non aggiungo me stesso alla lista
		if tempId == my_id {
			continue
		}
		currNode.id = tempId

		tempAddr := strings.TrimSpace(currNodeParts[1])
		tempTCPAddr, err := net.ResolveTCPAddr("tcp", tempAddr)
		tempUDPAddr, err := net.ResolveUDPAddr("udp", tempAddr)
		if err != nil {
			log.Printf("extractNodeList()---> errore risoluzione indirizzo remoto %s: %v", tempAddr, err)
		}
		currNode.UDPAddr = tempUDPAddr
		currNode.TCPAddr = tempTCPAddr
		currNode.strAddr = tempAddr
		currNode.state = 0

		nodes = append(nodes, currNode)
	}
}

// funzione che va a contattare i nodi della lista per vedere se sono attivi
// sceglie i nodi e poi invoca sendHeartBeat()
func contactNode(maxNumToContact int, isRand bool) {

	var selectedNode []*Node

	if isRand {
		//contatto in modo randomico

		elemToContact := make(map[int]bool)

		nodesMutex.Lock()
		lenght := len(nodes)
		//genero dei numeri casuali tutti differenti, corrispondono alla scelta di nodi da contattare
		i := 0
		for i < maxNumToContact {
			random := rand.Intn(lenght)
			_, ok := elemToContact[random]
			if !ok {
				elemToContact[random] = true
				selectedNode = append(selectedNode, &nodes[random])
				i++
			} else {
				continue
			}
		}
		nodesMutex.Unlock()

	} else {
		//contatto tutti quelli che conosco
		nodesMutex.Lock()
		lenght := len(nodes)
		for i := 0; i < lenght; i++ {
			selectedNode = append(selectedNode, &nodes[i])
		}
		nodesMutex.Unlock()
	}

	//contatto i nodi
	len := len(selectedNode)

	var wg sync.WaitGroup

	for i := 0; i < len; i++ {
		wg.Add(1)
		go sendHeartbeat(selectedNode[i], my_id, &wg)
	}
	wg.Wait()
}

// funzione che va ad inviare heartbeat ad un nodo
func sendHeartbeat(nodePointer *Node, myId int, wg *sync.WaitGroup) {

	defer wg.Done()

	//TODO controllare che il nodo sia attivo e sistemare la comunicazione
	if (*nodePointer).state == -1 {
		return
	} else {

		conn, err := net.DialUDP("udp", nil, nodePointer.UDPAddr)
		if err != nil {
			fmt.Println("sendHeartBeat()--> errore durante la connessione:", err)
			return
		}
		defer conn.Close()

		remoteId := (*nodePointer).id

		//TODO aggiungere il recupero del tempo di attesa
		//info necessarie per il nodo contattato

		startTime := time.Now()

		message := writeHeartBeatMessage(myId, ownUDPAddress.Port)

		timerErr := conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(def_RTT)))
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
		responseTime := time.Since(startTime)

		fmt.Printf("sendHeartBeat()--> responseTime: %v \n", responseTime)

		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Printf("sendHeartBeat()--> time_out scaduto, nodo sospetto id: %d\n", remoteId)

				//cambio dello stato del nodo
				nodesMutex.Lock()

				for _, node := range nodes {
					if node.id == remoteId {
						node.state = 2
					}
				}

				nodesMutex.Unlock()

				signalSus(remoteId)

				return
			}
		}

		nodesMutex.Lock()

		for _, node := range nodes {
			if node.id == remoteId {
				node.state = 1
			}
		}

		nodesMutex.Unlock()

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
	message := "000#id:" + strconv.Itoa(id) + "#port:" + strconv.Itoa(port) + "\n"
	return message
}
