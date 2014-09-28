package agent

import (
	"log"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(agent *Agent, pkt *packet.Pkt)


// Received response for the init packet
func HandleInit_Resp(agent *Agent, pkt *packet.Pkt) {
}

// Received response for the regist packet
func HandleRegist_Resp(agent *Agent, pkt *packet.Pkt) {
	log.Printf("Received Regist_Resp")

	dataRegResp := packet.PktDataRegResp{}
	err := packet.Unpack(pkt, &dataRegResp)
	if err != nil {
		log.Printf("Error unpack pkt: %s", err.Error())
	}

	log.Printf("AppId: [%s] RegId: [%s]", dataRegResp.AppId, dataRegResp.RegId)
	//RegIds[dataRegResp.AppId] = dataRegResp.RegId
	//SaveRegIds()
}

func HandlePush(agent *Agent, pkt *packet.Pkt) {
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
		//RegId: RegIds[dataMsg.AppId],
	}

	pktAck, err := packet.Pack(packet.PKT_ACK, 0, dataAck)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	agent.SendPkt(pktAck)
}
