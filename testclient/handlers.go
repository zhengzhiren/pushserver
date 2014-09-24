package main

var PktHandlers = map[uint8]tcpserver.PktHandler{}

func HandlePush(conn *net.TCPConn, pkt *packet.Pkt) {
	dataMsg := packet.PktDataMessage{}
	err := packet.Unpack(pkt, &dataMsg)
	if err != nil {
		log.Printf("Error unpack push msg: %s", err.Error())
	}
	log.Printf("Received push message: Appid: %s, Msg: %s\n", dataMsg.AppId, dataMsg.Msg)
}

func HandleACK(conn *net.TCPConn, pkt *packet.Pkt) {
}
