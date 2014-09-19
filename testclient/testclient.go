package main

import (
	"net"
	"log"
	"math/rand"
	"strconv"
	"time"
	"io"

	"pushserver/packet"
	"pushserver/tcpserver"
)

var pktHandlers = map[uint8]tcpserver.PktHandler {}

func main() {
	raddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	if err != nil {
		log.Printf("Unknown address: %s", err.Error())
		return
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		log.Printf("Dial error: %s", err.Error())
		return
	}
	defer func() {
		conn.Close()
	}()

	dataRegist := packet.PktDataRegist {
		DevId: strconv.Itoa(rand.Int() % 10000),		// a random Id
	}

	registPkt, err := packet.Pack(packet.PKT_Regist, uint32(rand.Int()), dataRegist)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	b, err := registPkt.Serialize()
	if err != nil {
		log.Printf("Serialize error: %s", err.Error())
	}
	conn.Write(b)

	var bufHeader = make([]byte, packet.PKT_HEADER_SIZE)
	for {
		//// check if we are exiting
		//select {
		//case <-this.exitChan:
		//	log.Printf("Closing connection from %s.\n", conn.RemoteAddr().String())
		//	return
		//default:
		//	// continue read
		//}

		const readTimeout = 100 * time.Millisecond
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		// read the packet header
		nbytes, err := io.ReadFull(conn, bufHeader)
		if err != nil {
			if err == io.EOF {
				log.Printf("read EOF, closing connection")
				return
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// just read timeout, not an error
				continue
			}
			log.Printf("read error: %s\n", err.Error())
			continue
		}
		log.Printf("%d bytes packet header read\n", nbytes)

		var pkt = packet.Pkt {
			Data : nil,
		}
		pkt.Header.Deserialize(bufHeader)

		// read the packet data
		if pkt.Header.Len > 0 {
			log.Printf("expecting data size: %d\n", pkt.Header.Len)
			var bufData = make([]byte, pkt.Header.Len)
			nbytes, err := io.ReadFull(conn, bufData)
			if err != nil {
				if err == io.EOF {
					log.Printf("read EOF, closing connection")
					return
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// read timeout
					//TODO
					log.Printf("read error: %s\n", err.Error())
					continue
				}
				log.Printf("read error: %s\n", err.Error())
				continue
			}
			log.Printf("%d bytes packet data read\n", nbytes)
			pkt.Data = bufData
		}

		handlePacket(conn, &pkt)
	}
}

func handlePacket(conn *net.TCPConn, pkt *packet.Pkt) {
	handler, ok := pktHandlers[pkt.Header.Type];
	if ok {
		handler(conn, pkt)
	} else {
		log.Printf("Unknown packet type: %d", pkt.Header.Type)
	}
}

func init() {
	pktHandlers[packet.PKT_Push] = HandlePush
	pktHandlers[packet.PKT_ACK] = HandleACK
}

func HandlePush(conn *net.TCPConn, pkt *packet.Pkt) {
}

func HandleACK(conn *net.TCPConn, pkt *packet.Pkt) {
}
