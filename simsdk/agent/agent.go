package agent

import (
	"encoding/json"
	"io"
	//"log"
	"net"
	"os"
	"time"

	log "github.com/cihub/seelog"

	"github.com/zhengzhiren/pushserver/packet"
)

const CHAN_LEN = 10

type OnRegResponse func(appId string, regId string, result int)
type OnReceiveMsg func(appId string, msgId int64, msgType int, msg string)

type Agent struct {
	deviceId      string
	serverAddr    *net.TCPAddr // Push server address
	pktHandlers   map[uint8]PktHandler
	outPkt        chan *packet.Pkt
	OnRegResponse OnRegResponse
	OnReceiveMsg  OnReceiveMsg
	RegIds        map[string]string
}

func NewAgent(devId string, serverAddr *net.TCPAddr) *Agent {
	agent := Agent{
		deviceId:    devId,
		serverAddr:  serverAddr,
		pktHandlers: map[uint8]PktHandler{},
		outPkt:      make(chan *packet.Pkt, CHAN_LEN),
		RegIds:      make(map[string]string),
	}
	agent.pktHandlers[packet.PKT_Init_Resp] = HandleInit_Resp
	agent.pktHandlers[packet.PKT_Regist_Resp] = HandleRegist_Resp
	agent.pktHandlers[packet.PKT_Unregist_Resp] = HandleUnregist_Resp
	agent.pktHandlers[packet.PKT_Push] = HandlePush
	return &agent
}

func (this *Agent) Run() {
	this.LoadRegIds()

	conn, err := net.DialTCP("tcp", nil, this.serverAddr)
	if err != nil {
		log.Errorf("Dial %v error: %s", this.serverAddr, err.Error())
		log.Flush()
		os.Exit(1)
		return
	}
	defer func() {
		conn.Close()
	}()

	dataInit := packet.PktDataInit{
		DevId: this.deviceId,
	}

	initPkt, err := packet.Pack(packet.PKT_Init, 0, dataInit)
	if err != nil {
		log.Error("Pack error: %s", err.Error())
		return
	}

	b, err := initPkt.Serialize()
	if err != nil {
		log.Error("Serialize error: %s", err.Error())
	}
	log.Debug(string(initPkt.Data))
	conn.Write(b)

	go func() {
		timer := time.NewTicker(20 * time.Second)
		hbPkt, _ := packet.Pack(packet.PKT_Heartbeat, 0, nil)
		heartbeat, _ := hbPkt.Serialize()
		for {
			select {
			//case <- done:
			//	break
			case pkt := <-this.outPkt:
				b, err := pkt.Serialize()
				if err != nil {
					log.Errorf("Serialize error: %s", err.Error())
				}
				log.Debugf("Write data: %s\n", b)
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

		const readTimeout = 300 * time.Second
		conn.SetReadDeadline(time.Now().Add(readTimeout))

		// read the packet header
		nbytes, err := io.ReadFull(conn, bufHeader)
		if err != nil {
			if err == io.EOF {
				log.Warn("read EOF, closing connection")
				return
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// just read timeout, not an error
				continue
			}
			log.Errorf("read error: %s\n", err.Error())
			return
		}
		log.Debugf("%d bytes packet header read\n", nbytes)

		var pkt = packet.Pkt{
			Data: nil,
		}
		pkt.Header.Deserialize(bufHeader)

		// read the packet data
		if pkt.Header.Len > 0 {
			log.Debugf("expecting data size: %d\n", pkt.Header.Len)
			var bufData = make([]byte, pkt.Header.Len)
			nbytes, err := io.ReadFull(conn, bufData)
			if err != nil {
				if err == io.EOF {
					log.Warn("read EOF, closing connection")
					return
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// read timeout
					//TODO: exit?
					log.Errorf("Timeout on receiving packet data. error: %s\n", err.Error())
					continue
				}
				log.Errorf("Error on receiving packet data. error: %s\n", err.Error())
				return
			}
			log.Debugf("%d bytes packet data read: %s\n", nbytes, bufData)
			pkt.Data = bufData
		}

		this.handlePacket(&pkt)
	}
}

func (this *Agent) SendPkt(pkt *packet.Pkt) {
	this.outPkt <- pkt
}

func (this *Agent) handlePacket(pkt *packet.Pkt) {
	handler, ok := this.pktHandlers[pkt.Header.Type]
	if ok {
		handler(this, pkt)
	} else {
		log.Warnf("Unknown packet type: %d", pkt.Header.Type)
	}
}

func (this *Agent) Regist(appid string, appkey string, regid string) {
	log.Infof("Regist AppId: %s, AppKey: %s, RegId: %s", appid, appkey, regid)
	dataRegist := packet.PktDataRegist{
		AppId:  appid,
		RegId:  regid,
		AppKey: appkey,
	}
	pktRegist, err := packet.Pack(packet.PKT_Regist, 0, &dataRegist)
	if err != nil {
		log.Errorf("Pack error: %s", err.Error())
		return
	}
	this.SendPkt(pktRegist)
}

func (this *Agent) Unregist(appid string, appkey string, regid string) {
	log.Infof("Unregist AppId: %s, AppKey: %s, RegId: %s", appid, appkey, regid)
	dataUnregist := packet.PktDataUnregist{
		AppId:  appid,
		AppKey: appkey,
		RegId:  regid,
	}
	pktUnregist, err := packet.Pack(packet.PKT_Unregist, 0, dataUnregist)
	if err != nil {
		log.Errorf("Pack error: %s", err.Error())
		return
	}
	this.SendPkt(pktUnregist)
}

func (this *Agent) SaveRegIds() {
	file, err := os.OpenFile("RegIds."+this.deviceId, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Errorf("OpenFile error: %s", err.Error())
		return
	}
	b, err := json.Marshal(this.RegIds)
	if err != nil {
		log.Errorf("Marshal error: %s", err.Error())
		file.Close()
		return
	}
	file.Write(b)
	file.Close()
}

func (this *Agent) LoadRegIds() {
	fileName := "RegIds." + this.deviceId
	file, err := os.Open(fileName)
	if err != nil {
		log.Warnf("Open error: %s", err.Error())
		return
	}
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil {
		log.Errorf("Read file error: %s", err.Error())
		file.Close()
		return
	}

	//log.Debugf("content in %s: %s", fileName, buf)

	err = json.Unmarshal(buf[:n], &this.RegIds)
	if err != nil {
		log.Errorf("Unarshal %s error: %s", fileName, err.Error())
		file.Close()
		return
	}

	log.Debugf("RegIds: %s", this.RegIds)
}
