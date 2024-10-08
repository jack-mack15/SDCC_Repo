# Progetto SDCC: Gossip based distance estimation and failure detection

### Set up Sistema Locale
Il set up del progetto può essere fatto in automatico lanciando lo script `setUpLocal.sh` dopo avergli fornito gli opportuni permessi per l’esecuzione, eseguendo il comando `sudo chmod +x setUpSystem.sh`.
Lo script di setUp eseguirà i seguenti step:
-	Installazione del tool Netem e caricamento dei pacchetti nel kernel (selezionare il comando di installazione nello script);
-	Build delle image per i container “node” e il container “registry”;
-	Gestione dei permessi di esecuzione degli script per lanciare il progetto.
Per il primo setUp: avviare l’intero script. Per i successivi setUp: commentare i comandi da non eseguire.

### Esecuzione locale
Per l’esecuzione del progetto: posizionarsi nella directory SDCC_repo/Progetto/main del progetto e installare Docker-Compose nella propria macchina.

Il progetto può essere eseguito in locale lanciando uno degli script: `simpleNetem.sh`, `variableNetem.sh`, `packetLossTest.sh` o `crashNode.sh`. 

Eseguendo uno di questi script, verranno creati i container, verranno eseguiti i comandi del tool Netem per il caso di test. Dopo tot secondi i container verranno fermati (`sudo docker-compose stop`) e viene restituito il controllo del terminale. Per osservare i risultati dell’esecuzione, eseguire il comando `sudo docker logs <name>`, sostituendo <name> con il nome dei container (node1, node2, node3, node4 e node5).

Prima di eseguire un successivo lancio, accertarsi di aver eliminato i container con il comando `sudo docker-compose down`.

Si possono modificare i parametri del software tramite il file `.env`, dove sono riportate tutte le variabili ambientali impostate nei container. Per informazioni su tali parametri vedere la sezione VARIABILI AMBIENTE.

È possibile anche modificare gli script per alterare i ritardi, i valori di packet loss o variazioni nei ritardi.



### Set Up su Istanza EC2
Avviata una istanza EC2, collegarsi via SSH e installare il progetto nell'istanza EC2. Un modo per installare il progetto è installare git con il comando `sudo yum install git -y` e successivamente installare il progetto con il comando `git clone https://github.com/jack-mack15/SDCC_repo.git`. 
Successivamente spostarsi nella directory main del progetto eseguendo il comando `cd SDCC_repo/Progetto/main`.
A questo punto, avviare lo script `setUpEc2.sh`. In modo automatico lo script installerà: Netem con il caricamento pacchetti nel kernel e Docker. Inoltre lo script andrà ad eseguire la build delle image dei nodi e del service registry ed infine cambierà i permessi di esecuzione dei vari test. Prima di lanciare lo script di set up, cambiare i permessi di esecuzione con il comando `sudo chmod +x setUpEc2.sh`.

### Esecuzione su Istanza EC2
Se si è effettuato il set up con lo script `setUpEc2.sh`, per l'esecuzione del sistema occorre semplicemente eseguire uno degli script: : `simpleNetem.sh`, `variableNetem.sh`, `packetLossTest.sh` o `crashNode.sh`, con il comando `sudo ./name.sh `. 
Per output, configurazioni sistema e netem, leggere Esecuzione Locale.

### Script di Esecuzione
-	`simpleNetem.sh`: dopo aver lanciato i container, vengono assegnati dei ritardi fissi ad ogni container;
-	`crashNode.sh`: dopo aver lanciato i container, viene rimosso un container simulando un crash;
-	`variableNetem.sh`: dopo aver lanciato i container, vengono assegnati dei ritardi fissi con piccole variazioni ad ogni container;
-	`packetLossTest.sh`: dopo aver lanciato i container, vengono assegnati dei ritardi fissi ad ogni container ed infine viene impostato un valore di packet loss ad uno dei container.

### Variabili Ambiente
Di seguito sono elencate tutte le variabili ambiente che vengono impostate in ogni container di tipo “node”, necessarie per configurare il comportamento del programma.
-	`DEFAULT_ERROR`: Indica il default error utilizzato dall’algoritmo di Vivaldi. L’errore poi ad ogni iterazione verrà modificato ma inizialmente viene impostato ad 1.0. Questo parametro è un float64 positivo.
-	`SCALE_FACTOR`: Indica lo scale factor, parametro utilizzato nell’algoritmo di Vivaldi nel calcolo del parametro delta. Corrisponde alla variabile Cc del paper originale. Float64 positivo.  
-	`PRECISION_WEIGHT`: Indica il parametro precision weight utilizzato nell’algoritmo di Vivaldi nel calcolo dell’errore. Corrisponde alla variabile Ce del paper originale. Float64 positivo.
-	`DEFAULT_RTT`: Valore di default per il round trip time (in millisecondi). Un nodo che contatta un altro nodo di cui non possiede un precedente valore di rtt misurato, usa il valore di default per i timeout delle connessioni. Intero positivo.
-	`RTT_MULTIPLIER`: Parametro che impatta il timeout delle connessioni e delle letture dei messaggi di risposta. Il timer sarà risultato della moltiplicazione tra questo parametro e il rtt precedentemente misurato tra la coppia di nodi. Float64 positivo.
-	`MESSAGE_INTERVAL`: Parametro che indica l’intervallo di tempo (in millisecondi) tra l’invio di messaggi Vivaldi ad altri nodi. Intero positivo.
-	`VIVALDI_PLUS_INFO`: Un messaggio Vivaldi contiene sia le informazioni del nodo contattato ma anche informazioni di altri nodi di cui dispone il nodo contattato. Questo parametro indica quanti nodi aggiuntivi il messaggio Vivaldi deve riportare. Intero positivo.
-	`IGNORE_IDS`: Stringa che indica quali nodi non posso scambiare messaggi direttamente. È una stringa che riporta gli id dei nodi separati da “/”.
-	`NODE_EACH_MESSAGE`: Parametro che indica quanti nodi contattare ad ogni iterazione di scambio messaggi. Intero positivo.
-	`GOSSIP_TYPE`: Parametro che indica quale tipo di algoritmo di gossip usare per la gestione e comunicazione dei fault del Sistema. Il valore 1 corrisponde a Bimodal Multicast; il valore 2 corrisponde a Blind Counter Rumor Mongering.
-	`GOSSIP_INTERVAL`: Parametro che indica l’intervallo di tempo (in millisecondi) tra l’invio di messaggi gossip per comunicare fault ai nodi vicini. Intero positivo.
-	`GOSSIP_MAX_NEIGHBOR`: Parametro del Blind Counter Rumor Mongering. Indica il numero massimo di nodi a cui inviare un messaggio di gossip ad ogni iterazione. Parametro B. Intero positivo. 
-	`GOSSIP_MAX_ITERATION`: Parametro del Blind Counter Rumor Mongering. Indica il numero massimo che un fault può essere ricevuto da un altro nodo o inoltrato ad altri nodi. Raggiunto questo numero il nodo attuale perde interessa a diffondere il fault. Parametro F. Intero positivo.
-	`FAULT_MAX_RETRY`: Parametro che indica quante volte un nodo contattato può non rispondere cosecutivamente prima di essere marcato come fault. Intero positivo.
-	`LAZZARUS_TRY`: Parametro che indica quante volte un nodo apparentemente fault può tentare di eseguire il protocollo Lazzarus. Dopo LAZZARUS_TRY volte con esito negativo, il nodo termina l’esecuzione. Intero positivo.
-	`LAZZARUS_INTERVAL`: Intervallo di tempo (in millisecondi) tra due tentativi di esecuzione protocollo Lazzarus. Intero positivo.
-	`SERVICE_REGISTRY_PORT`: Numero di porta su cui è in ascolto il service registry. 
-	`NODE_PORT`: Numero di porta su cui sono in ascolto tutti i nodi.
-	`SERVICE_REGISTRY_RETRY`: Numero di tentativi per contattare correttamente il service registry. Se dopo SERVICE_REGISTRY_RETRY tentative, un nodo non è riuscito a contattare il service registry, il nodo termina l’esecuzione. Intero positivo.
-	`ITERATION_PRINT`: Parametro che indica la frequenza di stampa dei risultati. Indica il numero di iterazioni di scambi messaggi. Intero positivo.


