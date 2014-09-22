package tcpserver

import (
	"net"
	"sync"

	"github.com/zhengzhiren/pushserver/packet"
)

type Client struct {
	Conn *net.TCPConn
	Id string
	PktChan chan *packet.Pkt
}

var (
	ClientMap = make(map[string]*Client)
	ClientMapLock = sync.RWMutex {}	//lock for the ClientMap
)

