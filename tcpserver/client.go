package tcpserver

import (
	"net"
	"sync"
	"time"

	"github.com/zhengzhiren/pushserver/packet"
)

type Client struct {
	Conn          *net.TCPConn
	Id            string
	AppIds        []string
	PktChan       chan *packet.Pkt
	LastHeartbeat time.Time
}

var (
	ClientMap     = make(map[string]*Client)
	ClientMapLock = sync.RWMutex{} //lock for the ClientMap
)

func (this *Client) SendMsg(msg string, appid string) error {
	pktDataMsg := packet.PktDataMessage{
		Msg:   msg,
		AppId: appid,
	}

	pkt, err := packet.Pack(packet.PKT_Push, 0, &pktDataMsg)
	if err != nil {
		return err
	}
	this.PktChan <- pkt
	return nil
}
