package receiverrpc

import (
	"log"
)

type Receiver struct {
}

type ArgOnReceiveMsg struct {
	MsgId   int64
	MsgType int
	Msg     string
}

type ReplyOnReceiveMsg struct {
}

func (this *Receiver) OnReceiveMsg(arg ArgOnReceiveMsg, reply *ReplyOnReceiveMsg) error {
	log.Printf("RPC: OnReceiveMsg")
	log.Printf("MsgId: %d, MsgType: %d, Msg: %s", arg.MsgId, arg.MsgType, arg.Msg)
	return nil
}
