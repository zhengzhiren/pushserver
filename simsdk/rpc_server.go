package main

import (
	"net"
	"net/rpc"
	"time"
	"log"
)

func RunRPC(port int) {
	log.Printf("Starting RPC server\n")
	laddr := net.TCPAddr {
		IP: net.ParseIP("0.0.0.0"),
		Port: port,
	}
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		log.Printf("Failed to start RPC server: %s", err.Error())
		return
	}

	sdk := new(SDK)
	rpc.Register(sdk)
	log.Printf("RPC server is listening on %s\n", laddr.String())

	defer func() {
		// close the listener sock
		log.Printf("Closing listener socket.\n")
		ln.Close()
	}()

	for {
		ln.SetDeadline(time.Now().Add(time.Second))
		conn, err := ln.AcceptTCP()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// just accept timeout, not an error
				continue
			}
			log.Printf("Failed to accept: %s", err.Error())
			continue
		}
		go rpc.ServeConn(conn)
	}
}
