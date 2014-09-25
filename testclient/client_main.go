package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/zhengzhiren/pushserver/packet"
)

const CHAN_LEN = 10

var (
	AppIds   []string
	DeviceId = ""
	OutPkt = make(chan *packet.Pkt, CHAN_LEN)
)

func init() {
	PktHandlers[packet.PKT_Init_Resp] = HandleInit_Resp
	PktHandlers[packet.PKT_Regist_Resp] = HandleRegist_Resp
	PktHandlers[packet.PKT_Push] = HandlePush
}

func main() {
	log.SetPrefix("testclient ")

	dst := ""

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-i": // device id
			i++
			if i >= len(os.Args) {
				fmt.Printf("missing argument for \"-i\"\n")
				return
			}
			DeviceId = os.Args[i]
		case "-a": // appid
			i++
			if i >= len(os.Args) {
				fmt.Printf("missing argument for \"-a\"\n")
				return
			}
			AppIds = append(AppIds, os.Args[i])
		default:
			if dst == "" {
				dst = os.Args[i]
			} else {
				fmt.Printf("unknown argument %s\n", os.Args[i])
				return
			}
		}
	}

	if dst == "" {
		fmt.Printf("no destination address\n")
		return
	}
	if len(AppIds) == 0 {
		fmt.Printf("no AppId on this device\n")
		return
	}
	if DeviceId == "" {
		rand.Seed(time.Now().Unix())
		DeviceId = strconv.Itoa(rand.Int() % 10000) // a random Id
	}

	raddr, err := net.ResolveTCPAddr("tcp", dst)
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

	dataInit := packet.PktDataInit{
		DevId: DeviceId,
	}

	initPkt, err := packet.Pack(packet.PKT_Init, uint32(rand.Int()), dataInit)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}

	b, err := initPkt.Serialize()
	if err != nil {
		log.Printf("Serialize error: %s", err.Error())
	}
	log.Printf(string(initPkt.Data))
	conn.Write(b)

	go func() {
		timer := time.NewTicker(20 * time.Second)
		hbPkt, _ := packet.Pack(packet.PKT_Heartbeat, 0, nil)
		heartbeat, _ := hbPkt.Serialize()
		for {
			select {
			//case <- done:
			//	break
			case pkt := <-OutPkt:
				b, err := pkt.Serialize()
				if err != nil {
					log.Printf("Serialize error: %s", err.Error())
				}
				conn.Write(b)
			case <-timer.C:
				conn.Write(heartbeat)
			}
		}
	}()

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

		var pkt = packet.Pkt{
			Data: nil,
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
			log.Printf("%d bytes packet data read: %s\n", nbytes, bufData)
			pkt.Data = bufData
		}

		handlePacket(conn, &pkt)
	}
}

func handlePacket(conn *net.TCPConn, pkt *packet.Pkt) {
	handler, ok := PktHandlers[pkt.Header.Type]
	if ok {
		handler(conn, pkt)
	} else {
		log.Printf("Unknown packet type: %d", pkt.Header.Type)
	}
}
