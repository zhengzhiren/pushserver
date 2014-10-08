package agent

import (
	//"log"

	log "github.com/cihub/seelog"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(agent *Agent, pkt *packet.Pkt)

// Received response for the init packet
func HandleInit_Resp(agent *Agent, pkt *packet.Pkt) {
	log.Info("Received Init_Resp")
}

// Received response for the regist packet
func HandleRegist_Resp(agent *Agent, pkt *packet.Pkt) {
	dataRegResp := packet.PktDataRegResp{}
	err := packet.Unpack(pkt, &dataRegResp)
	if err != nil {
		log.Errorf("Error unpack pkt: %s", err.Error())
	}

	log.Infof("Received Regist_Resp AppId: [%s] RegId: [%s]", dataRegResp.AppId, dataRegResp.RegId)
	agent.RegIds[dataRegResp.AppId] = dataRegResp.RegId
	agent.SaveRegIds()
	if agent.OnRegResponse != nil {
		agent.OnRegResponse(dataRegResp.AppId, dataRegResp.RegId, dataRegResp.Result)
	}
}

// Received response for the unregist packet
func HandleUnregist_Resp(agent *Agent, pkt *packet.Pkt) {
	dataUnregResp := packet.PktDataUnregResp{}
	err := packet.Unpack(pkt, &dataUnregResp)
	if err != nil {
		log.Errorf("Error unpack pkt: %s", err.Error())
	}

	log.Infof("Received Unregist_Resp AppId: [%s] RegId: [%s]", dataUnregResp.AppId, dataUnregResp.RegId)
	delete(agent.RegIds, dataUnregResp.AppId)
	agent.SaveRegIds()
}

func HandlePush(agent *Agent, pkt *packet.Pkt) {
	dataMsg := packet.PktDataMessage{}
	err := packet.Unpack(pkt, &dataMsg)
	if err != nil {
		log.Errorf("Error unpack push msg: %s", err.Error())
	}
	log.Infof("Received push message: MsgId: %d, Type: %d,  Appid: %s, Msg: %s\n",
		dataMsg.MsgId, dataMsg.MsgType, dataMsg.AppId, dataMsg.Msg)

	dataAck := packet.PktDataACK{
		MsgId: dataMsg.MsgId,
		AppId: dataMsg.AppId,
		RegId: agent.RegIds[dataMsg.AppId],
	}

	pktAck, err := packet.Pack(packet.PKT_ACK, 0, dataAck)
	if err != nil {
		log.Errorf("Pack error: %s", err.Error())
		return
	}

	agent.SendPkt(pktAck)

	if agent.OnReceiveMsg != nil {
		agent.OnReceiveMsg(dataMsg.AppId, dataMsg.MsgId, dataMsg.MsgType, dataMsg.Msg)
	}
}
