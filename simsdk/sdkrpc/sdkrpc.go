package sdkrpc

import (
	"log"

	"github.com/zhengzhiren/pushserver/simsdk/agent"
)

type SDK struct {
	Agent *agent.Agent
}

type ArgRegist struct {
	AppId string
	AppKey string
	RegId string
}

type ReplyRegist struct {
}

func (this *SDK) Regist(arg ArgRegist, reply *ReplyRegist) error {
	log.Printf("RPC: Regist")
	this.Agent.Regist(arg.AppId, arg.AppKey, arg.RegId)
	return nil
}
