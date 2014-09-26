package main

import (
	"fmt"
	"io"
	"flag"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"encoding/json"

	"github.com/zhengzhiren/pushserver/packet"
)

const CHAN_LEN = 10

var (
	RegIds   = make(map[string]string)
	DeviceId = ""
	OutPkt   = make(chan *packet.Pkt, CHAN_LEN)
	RpcPort int
)

func init() {
	PktHandlers[packet.PKT_Init_Resp] = HandleInit_Resp
	PktHandlers[packet.PKT_Regist_Resp] = HandleRegist_Resp
	PktHandlers[packet.PKT_Push] = HandlePush
}

func main() {
	log.SetPrefix(os.Args[0])

	flag.StringVar(&DeviceId, "i", "", "Device Id")
	flag.IntVar(&RpcPort, "p", 9988, "Port for RPC")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Printf("missing ip:port\n")
		return
	}

	dst := flag.Args()[0]

	if DeviceId == "" {
		rand.Seed(time.Now().Unix())
		DeviceId = strconv.Itoa(rand.Int() % 10000) // a random Id
	}
	fmt.Printf("Device Id: [%s], RPC port: [%d]\n", DeviceId, RpcPort)

	//go RunRPC(RpcPort)
	RunRPC(RpcPort)

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
				log.Printf("Write data: %s\n", b)
				conn.Write(b)
			case <-timer.C:
				conn.Write(heartbeat)
			}
		}
	}()

	go func() {
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
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	for {
		s := <-ch
		log.Println("Received signal:", s)
		if len(RegIds) > 0 {
			for appid, regid := range RegIds {
				log.Printf("Unregist AppId: [%s], RegId: [%s]", appid, regid)
				Unregist(appid)
				break
			}
		} else {
			log.Printf("Stopped")
			break
		}
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

func Unregist(appid string) {
	dataUnregist := packet.PktDataUnregist{
		AppId:  appid,
		AppKey: "tempkey",
		RegId:  RegIds[appid],
	}
	pktUnregist, err := packet.Pack(packet.PKT_Unregist, 0, dataUnregist)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return
	}
	OutPkt <- pktUnregist
	delete(RegIds, appid)
}

func SaveRegIds() {
	file, err := os.OpenFile("RegIds.txt", os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		log.Printf("OpenFile error: %s", err.Error())
		return
	}
	b, err := json.Marshal(RegIds)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
		file.Close()
		return
	}
	file.Write(b)
	file.Close()
}

func LoadRegIds() {
	file, err := os.Open("RegIds.txt")
	if err != nil {
		log.Printf("Open error: %s", err.Error())
		return
	}
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil {
		log.Printf("Read file error: %s", err.Error())
		file.Close()
		return
	}

	log.Printf("%s", buf)

	regIds := map[string]string {}
	err = json.Unmarshal(buf[:n], &regIds)
	if err != nil {
		log.Printf("Unarshal error: %s", err.Error())
		file.Close()
		return
	}
	for appid, _ := range RegIds {
		RegIds[appid] = regIds[appid]
	}

	log.Printf("RegIds: %s", RegIds)
	log.Printf("regIds: %s", regIds)
}
