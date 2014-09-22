package tcpserver

import (
	"net"
	"log"

	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func (conn *net.TCPConn, pkt *packet.Pkt)

func HandleRegist(conn *net.TCPConn, pkt *packet.Pkt) {
	log.Printf("Handling packet type: Regist")
	var dataRegist = packet.PktDataRegist{}
	err := packet.Unpack(pkt, &dataRegist)
	if err != nil {
		log.Printf("Failed to Unpack: %s", err.Error())
		return
	}

	log.Printf("New Device is online, Id: %s", dataRegist.DevId)

	// TODO: check if the Id already exist

	client := Client{
		Conn: conn,
		Id: dataRegist.DevId,
		PktChan: make(chan *packet.Pkt),
	}

	go func() {
		for {
			select {
			case pkt := <-client.PktChan:
				log.Printf("pkt to send")
				b, err := pkt.Serialize()
				if err != nil {
					log.Printf("Error on serializing pkt: %s", err.Error())
					continue
				}
				var n int
				n, err = conn.Write(b)
				if err != nil {
					log.Printf("Error on sending pkt: %s", err.Error())
					continue
				}
				log.Printf("Write successfully: %d", n)
			}
		}
	}()

	ClientMapLock.Lock()
	ClientMap[client.Id] = &client
	ClientMapLock.Unlock()
}

func HandleHeartbeat(conn *net.TCPConn, pkt *packet.Pkt) {
}

func HandleACK(conn *net.TCPConn, pkt *packet.Pkt) {
}
