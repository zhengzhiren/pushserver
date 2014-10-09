package receiverrpc

import (
	//"log"

	log "github.com/cihub/seelog"
)

var lastMsgId int64 = 0

type Receiver struct {
	RegId string
}

type ArgOnReceiveMsg struct {
	MsgId   int64
	MsgType int
	Msg     string
}

type ReplyOnReceiveMsg struct {
}

func (this *Receiver) OnReceiveMsg(arg ArgOnReceiveMsg, reply *ReplyOnReceiveMsg) error {
	log.Trace("RPC: OnReceiveMsg")
	log.Infof("Received message. MsgId: %d, MsgType: %d, Msg: %s", arg.MsgId, arg.MsgType, arg.Msg)
	if lastMsgId == 0 {
		lastMsgId = arg.MsgId
	} else if lastMsgId >= arg.MsgId {
		log.Errorf("Received bad message Id: %d, LastMsgId: %d", arg.MsgId, lastMsgId)
	}
	return nil
}

type ArgOnRegResp struct {
	AppId  string
	RegId  string
	Result int
}

type ReplyOnRegResp struct {
}

func (this *Receiver) OnRegResp(arg ArgOnRegResp, reply *ReplyOnRegResp) error {
	log.Trace("RPC: OnRegResp")
	log.Infof("Regist response: RegId: %s, Result: %d", arg.RegId, arg.Result)
	this.RegId = arg.RegId
	if arg.Result != 0 {
		log.Error("Regist failed")
	}
	return nil
}
