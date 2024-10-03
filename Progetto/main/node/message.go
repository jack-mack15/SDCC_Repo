package main

//codes:
//0 per ack
//1 per vivaldi message
//2 per vivaldi response
//3 per gossip message
//4 per vivaldi message più digest

// messaggio per vivaldi algorithm
type VivaldiMessage struct {
	Code        int         `json:"code"`
	IdSender    int         `json:"idsender"`
	PortSender  int         `json:"port"`
	Coordinates coordinates `json:"coord"`
	Digest      string      `json:"digest"`
}

// messaggio per segnalare la corretta attività
type SimpleAck struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// messaggio di gossip per blind counter rumor
type GossipMessage struct {
	Code       int    `json:"code"`
	IdSender   int    `json:"idsender"`
	PortSender int    `json:"port"`
	IdFault    int    `json:"idfault"`
	Digest     string `json:"digest"`
}

// funzione che compila il messaggio di gossip per il blind counter
func writeGossipMessage(faultId int) GossipMessage {
	var message GossipMessage
	message.Code = 3
	message.IdSender = getMyId()
	message.PortSender = getMyPort()
	message.IdFault = faultId
	return message
}

// funzione che compila il messaggio di semplice ack
func writeSimpleAck() SimpleAck {
	var message SimpleAck
	message.Code = 0
	message.Message = "hello"
	return message
}

func writeVivaldiMessage() VivaldiMessage {
	var message VivaldiMessage

	message.Code = 1
	message.IdSender = getMyId()
	message.PortSender = getMyPort()
	var coord coordinates
	coord.X = getMyCoordinate().X
	coord.Y = getMyCoordinate().Y
	coord.Z = getMyCoordinate().Z
	coord.Error = getMyCoordinate().Error
	message.Coordinates = coord

	if getGossipType() == 1 && getDigest() != "" {
		message.Code = 4
		message.Digest = getDigest()
		return message
	}

	message.Digest = ""
	return message
}
