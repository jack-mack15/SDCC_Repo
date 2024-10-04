package main

//codes:
//1 per vivaldi message
//2 per vivaldi response
//3 per gossip message
//4 per vivaldi message più digest

// messaggio per vivaldi algorithm
type VivaldiMessage struct {
	Code        int                 `json:"code"`
	IdSender    int                 `json:"idsender"`
	PortSender  int                 `json:"port"`
	Coordinates coordinates         `json:"coord"`
	MapCoor     map[int]coordinates `json:"map"`
	Digest      string              `json:"digest"`
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

// funzion eche serve per scrivere un messaggio di Gossip solo per il digest
func writeGossipDigestMessage(senderId int, digest string) GossipMessage {
	var message GossipMessage
	message.Code = 4
	message.IdSender = senderId
	message.Digest = digest
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
	message.MapCoor = nil
	message.Digest = ""
	return message
}

func writeVivaldiResponse(idSender int) VivaldiMessage {
	message := writeVivaldiMessage()
	message.MapCoor = getRandomNodes(idSender)
	return message
}
