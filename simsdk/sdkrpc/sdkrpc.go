package sdkrpc

import (
	//"log"
	"net"
	"net/rpc"

	log "github.com/cihub/seelog"

	"github.com/zhengzhiren/pushserver/simsdk/agent"
)

type SDK struct {
	Agent     *agent.Agent
	Receivers map[string]*rpc.Client // appid as key
}

type ArgRegist struct {
	AppId   string
	AppKey  string
	RegId   string
	Token   string
	AppAddr *net.UnixAddr // RPC address for Receiver
}

type ReplyRegist struct {
}

func (this *SDK) Regist(arg ArgRegist, reply *ReplyRegist) error {
	log.Infof("RPC: Regist. AppId: %s, AppKey: %s, RegId: %s, AppAddr: %v",
		arg.AppId, arg.AppKey, arg.RegId, arg.AppAddr)

	conn, err := net.DialUnix("unix", nil, arg.AppAddr)
	if err != nil {
		log.Errorf("Dial error: %s", err.Error())
		return err
	}

	this.Receivers[arg.AppId] = rpc.NewClient(conn)
	if this.Receivers[arg.AppId] == nil {
		log.Errorf("client nil")
		return nil
	}

	regId := arg.RegId
	if regId == "" {
		regId = this.Agent.RegIds[arg.AppId]
	}
	this.Agent.Regist(arg.AppId, arg.AppKey, regId, arg.Token)
	return nil
}

type ArgUnregist struct {
	AppId  string
	AppKey string
	RegId  string
	Token  string
}

type ReplyUnregist struct {
}

func (this *SDK) Unregist(arg ArgUnregist, reply *ReplyUnregist) error {
	log.Infof("RPC: Unregist")
	this.Agent.Unregist(arg.AppId, arg.AppKey, arg.RegId, arg.Token)
	return nil
}
