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
	for appid, _ := range RegIds {
		dataRegist := packet.PktDataRegist{
			AppId:  appid,
			RegId: RegIds[appid],
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
	log.Printf("Received Regist_Resp")

	dataRegResp := packet.PktDataRegResp{}
	err := packet.Unpack(pkt, &dataRegResp)
	if err != nil {
		log.Printf("Error unpack pkt: %s", err.Error())
	}

	log.Printf("AppId: [%s] RegId: [%s]", dataRegResp.AppId, dataRegResp.RegId)
	RegIds[dataRegResp.AppId] = dataRegResp.RegId
	SaveRegIds()
}

func HandlePush(conn *net.TCPConn, pkt *packet.Pkt) {
	dataMsg := packet.PktDataMessage{}
	err := packet.Unpack(pkt, &dataMsg)
	if err != nil {
		log.Printf("Error unpack push msg: %s", err.Error())
	}
	log.Printf("Received push message: MsgId: %d, Type: %d,  Appid: %s, Msg: %s\n",
		dataMsg.MsgId, dataMsg.MsgType, dataMsg.AppId, dataMsg.Msg)

	dataAck := packet.PktDataACK{
		MsgId: dataMsg.MsgId,
		AppId: dataMsg.AppId,
		RegId: RegIds[dataMsg.AppId],
	}

	pktAck, err := packet.Pack(packet.PKT_ACK, 0, dataAck)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	OutPkt <- pktAck
}
