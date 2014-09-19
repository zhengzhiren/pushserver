package tcpserver

import (
	"net"
	"sync"
)

type Client struct {
	Conn *net.TCPConn
	Id string
	MsgChan chan *Message
}

var (
	ClientMap = make(map[string]*Client)
	ClientMapLock = sync.RWMutex {}	//lock for the ClientMap
)

