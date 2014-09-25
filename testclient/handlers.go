package main

import (
	"log"
	"net"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(conn *net.TCPConn, pkt *packet.Pkt)

var PktHandlers = map[uint8]PktHandler{}

// Received response for the init packet
func HandleInit_Resp(conn *net.TCPConn, pkt *packet.Pkt) {
	// send Regist packet for each App
	for _, appid := range AppIds {
		dataRegist := packet.PktDataRegist{
			AppId: appid,
			AppKey: "temp_key",
		}
		pktRegist, err := packet.Pack(packet.PKT_Regist, 0, &dataRegist)
		if err != nil {
			log.Printf("Pack error: %s", err.Error())
			return
		}
		OutPkt <- pktRegist
	}
}

// Received response for the regist packet
func HandleRegist_Resp(conn *net.TCPConn, pkt *packet.Pkt) {
}

func HandlePush(conn *net.TCPConn, pkt *packet.Pkt) {
	dataMsg := packet.PktDataMessage{}
	err := packet.Unpack(pkt, &dataMsg)
	if err != nil {
		log.Printf("Error unpack push msg: %s", err.Error())
	}
	log.Printf("Received push message: Appid: %s, Msg: %s\n", dataMsg.AppId, dataMsg.Msg)

	pktAck, err := packet.Pack(packet.PKT_ACK, 0, nil)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	OutPkt <- pktAck
}
