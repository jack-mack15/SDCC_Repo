package node

/*
// funzione che va ad inviare heartbeat ad un nodo
func sendHeartbeat(nodePointer *Node, myId int, wg *sync.WaitGroup) {

	defer wg.Done()

	remoteAddr := (*nodePointer).StrAddr

	//TODO controllare che il nodo sia attivo e sistemare la comunicazione
	if (*nodePointer).State == -1 {
		return
	} else {
		conn, err := net.Dial("tcp", remoteAddr)
		if err != nil {
			fmt.Println("sendHeartBeat()--> errore durante la connessione:", err)
			return
		}
		defer conn.Close()

		remoteId := (*nodePointer).ID

		conn.SetReadDeadline(time.Now().Add(time.Millisecond * time.Duration(defRTT)))

		//info necessarie per il nodo contattato
		message := writeHeartBeatMessage(myId, My_addr.Port)

		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("sendHeartBeat()--> errore durante l'invio del messaggio:", err)
			return
		}

		//risposta dal nodo contattato
		reader := bufio.NewReader(conn)
		reply, err := reader.ReadString('\n')

		if err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				fmt.Printf("sendHeartBeat()--> time_out scaduto, nodo sospetto ID: %d\n", remoteId)

				//cambio dello stato del nodo
				nodesMutex.Lock()

				for _, node := range nodes {
					if node.ID == remoteId {
						node.State = 2
					}
				}

				nodesMutex.Unlock()

				signalSus(remoteId)

				return
			}
		}

		nodesMutex.Lock()

		for _, node := range nodes {
			if node.ID == remoteId {
				node.State = 1
			}
		}

		nodesMutex.Unlock()

		fmt.Printf("sendHeartBeat()--> risposta dal nodo: %s\n", reply)
	}
}


// funzione che smista le richieste di connessioni da parte di altri nodi
func receiverHandler() {

	listener, err := net.Listen("tcp", ":"+strconv.Itoa(My_addr.Port))
	if err != nil {
		fmt.Println("receiverHandler()--> errore durante l'ascolto:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("receiverHandler()--> errore nella connessione:", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}
*/
