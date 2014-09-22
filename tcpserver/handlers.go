package tcpserver

import (
	"net"
	"log"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func (conn *net.TCPConn, pkt *packet.Pkt)

func HandleRegist(conn *net.TCPConn, pkt *packet.Pkt) {
	log.Printf("Handling packet type: Regist")
	var dataRegist = packet.PktDataRegist{}
	err := packet.Unpack(pkt, &dataRegist)
	if err != nil {
		log.Printf("Failed to Unpack: %s", err.Error())
		return
	}

	log.Printf("New Device is online, Id: %s", dataRegist.DevId)

	client := Client{
		Conn: conn,
		Id: dataRegist.DevId,
		MsgChan: make(chan *Message),
	}

	ClientMapLock.Lock()
	ClientMap[client.Id] = &client
	ClientMapLock.Unlock()
}

func HandleHeartbeat(conn *net.TCPConn, pkt *packet.Pkt) {
}

func HandleACK(conn *net.TCPConn, pkt *packet.Pkt) {
}
