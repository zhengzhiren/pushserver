package sdkrpc

import (
	"log"
	"net"
	"net/rpc"

	"github.com/zhengzhiren/pushserver/simsdk/agent"
)

type SDK struct {
	Agent     *agent.Agent
	Receivers map[string]*rpc.Client // appid as key
}

type ArgRegist struct {
	AppId        string
	AppKey       string
	RegId        string
	ReceiverAddr net.TCPAddr // RPC address for Receiver
}

type ReplyRegist struct {
}

func (this *SDK) Regist(arg ArgRegist, reply *ReplyRegist) error {
	log.Printf("RPC: Regist")
	this.Agent.Regist(arg.AppId, arg.AppKey, arg.RegId)

	conn, err := net.DialTCP("tcp", nil, &arg.ReceiverAddr)
	if err != nil {
		log.Printf("Dial error: %s", err.Error())
		return err
	}

	this.Receivers[arg.AppId] = rpc.NewClient(conn)
	if this.Receivers[arg.AppId] == nil {
		log.Fatal("client nil")
	}
	return nil
}

type ArgUnregist struct {
	AppId  string
	AppKey string
	RegId  string
}

type ReplyUnregist struct {
}

func (this *SDK) Unregist(arg ArgUnregist, reply *ReplyUnregist) error {
	log.Printf("RPC: Unregist")
	this.Agent.Unregist(arg.AppId, arg.AppKey, arg.RegId)
	return nil
}
