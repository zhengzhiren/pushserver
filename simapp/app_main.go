package main

import (
	"flag"
	"fmt"
	//"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/cihub/seelog"

	"github.com/zhengzhiren/pushserver/simapp/receiverrpc"
	"github.com/zhengzhiren/pushserver/simsdk/sdkrpc"
)

var (
	Receiver receiverrpc.Receiver
)

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("simapp [-d sdk_ip] [-p sdk_rpc_port] [-r receive_port] <app_id> <app_key>\n")
}

func main() {
	var (
		rpcPort      int
		ip           string
		receiverPort int
	)

	flag.StringVar(&ip, "d", "127.0.0.1", "SDK IP")
	flag.IntVar(&rpcPort, "p", 9988, "Dest SDK port for RPC")
	flag.IntVar(&receiverPort, "r", 9888, "Receiver port")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Printf("need AppId and AppKey\n")
		usage()
		return
	}

	appId := flag.Args()[0]
	appKey := flag.Args()[1]

	receiverAddr := net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: receiverPort,
	}

	raddr := net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: rpcPort,
	}
	if raddr.IP == nil {
		fmt.Printf("Invalid IP address\n")
		return
	}

	go RunReceiverRPC(receiverPort)

	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		log.Errorf("Dial error: %s", err.Error())
		return
	}
	defer func() {
		conn.Close()
	}()

	log.Info("Registing")

	rpcClient := rpc.NewClient(conn)

	// call regist
	argRegist := sdkrpc.ArgRegist{
		AppId:        appId,
		AppKey:       appKey,
		ReceiverAddr: receiverAddr,
	}
	replyRegist := sdkrpc.ReplyRegist{}
	err = rpcClient.Call("SDK.Regist", argRegist, &replyRegist)
	if err != nil {
		log.Errorf("RPC error: %s", err.Error())
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)

	s := <-ch
	log.Infof("Received signal: %v", s)
	log.Info("Unregisting")
	argUnregist := sdkrpc.ArgUnregist{
		AppId:  appId,
		AppKey: appKey,
		RegId:  Receiver.RegId,
	}
	replyUnregist := sdkrpc.ReplyUnregist{}
	err = rpcClient.Call("SDK.Unregist", argUnregist, &replyUnregist)
}

func RunReceiverRPC(port int) {
	log.Info("Starting Receiver RPC\n")
	laddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		log.Errorf("Failed to start RPC server: %s", err.Error())
		return
	}

	Receiver = receiverrpc.Receiver{}
	rpc.Register(&Receiver)
	log.Infof("RPC server is listening on %s\n", laddr.String())

	defer func() {
		// close the listener sock
		log.Info("Closing listener socket.\n")
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
			log.Error("Failed to accept: %s", err.Error())
			continue
		}
		rpc.ServeConn(conn)
	}
}
