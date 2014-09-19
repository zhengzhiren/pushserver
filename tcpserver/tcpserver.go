package tcpserver

import (
	"net"
	"log"
	"time"
	"sync"
	"io"

	"pushserver/packet"
)


type TcpServer struct {
	exitChan chan bool
	waitGroup *sync.WaitGroup
	pktHandlers map[uint8]PktHandler
}

func Create() *TcpServer{
	server := &TcpServer {
		exitChan: make(chan bool),
		waitGroup: &sync.WaitGroup{},
		pktHandlers: map[uint8]PktHandler{},
	}
	server.pktHandlers[packet.PKT_Regist] = HandleRegist
	server.pktHandlers[packet.PKT_Heartbeat] = HandleHeartbeat
	server.pktHandlers[packet.PKT_ACK] = HandleACK
	return server
}

func (this *TcpServer) Start() {
	log.Printf("Starting TcpServer\n")
	laddr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:9999")
	ln, err := net.ListenTCP("tcp", laddr)
	if (err != nil) {
		log.Printf("Failed to start TcpServer: %s", err.Error())
		return
	}

	log.Printf("TcpServer is listening on %s\n", laddr.String())

	this.waitGroup.Add(1)

	defer func() {
		// close the listener sock
		log.Printf("Closing listener socket.\n")
		ln.Close()
		this.waitGroup.Done()
	}()

	for {
		// check if we are exiting
		select {
		case <-this.exitChan:
			log.Printf("Stopping TcpServer.\n")
			return
		default:
			// continue accept new connection
		}

		ln.SetDeadline(time.Now().Add(time.Second))
		conn, err := ln.AcceptTCP()
		if (err != nil) {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// just accept timeout, not an error
				continue
			}
			log.Printf("Failed to accept: %s", err.Error())
			continue
		}

		// handle new connection
		go this.handleConn(conn)
	}
}

func (this *TcpServer) Stop() {
	close(this.exitChan)
	this.waitGroup.Wait()
}

func (this *TcpServer) handleConn(conn *net.TCPConn) {
	this.waitGroup.Add(1)
	log.Printf("New conn accepted from %s\n", conn.RemoteAddr().String())

	defer func() {
		conn.Close()
		//FIXME: remove conn
		//ClientMapLock.Lock()
		//delete(ClientMap, client.Id)
		//ClientMapLock.Unlock()
		this.waitGroup.Done()
	}()

	var bufHeader = make([]byte, packet.PKT_HEADER_SIZE)
	for {
		// check if we are exiting
		select {
		case <-this.exitChan:
			log.Printf("Closing connection from %s.\n", conn.RemoteAddr().String())
			return
		default:
			// continue read
		}

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
			//FIXME: check the Len
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

		this.handlePacket(conn, &pkt)
	}
}

func (this *TcpServer) handlePacket(conn *net.TCPConn, pkt *packet.Pkt) {
	handler, ok := this.pktHandlers[pkt.Header.Type];
	if ok {
		handler(conn, pkt)
	} else {
		log.Printf("Unknown packet type: %d", pkt.Header.Type)
	}
}
