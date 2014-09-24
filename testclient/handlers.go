package main

import (
	"log"
	"net"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(conn *net.TCPConn, pkt *packet.Pkt)

var PktHandlers = map[uint8]PktHandler{}

// Received response for the init packet, send Regist packet
func HandleInit_Resp(conn *net.TCPConn, pkt *packet.Pkt) {
	dataRegist := packet.PktDataRegist{
		AppIds: AppIds,
	}

	pktRegist, err := packet.Pack(packet.PKT_Regist, 0, &dataRegist)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	b, err := pktRegist.Serialize()
	if err != nil {
		log.Printf("Serialize error: %s", err.Error())
	}
	conn.Write(b)
}

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
