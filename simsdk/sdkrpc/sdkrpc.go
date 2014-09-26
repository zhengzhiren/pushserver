package sdkrpc

import (
	"log"
)

type SDK int

type ArgRegist struct {
	AppId string
}

type ReplyRegist struct {
}

func (t *SDK) Regist(arg ArgRegist, reply *ReplyRegist) error {
	log.Printf("RPC: Regist")
	return nil
}
