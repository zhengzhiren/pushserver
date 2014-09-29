package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zhengzhiren/pushserver/simapp/receiverrpc"
	"github.com/zhengzhiren/pushserver/simsdk/sdkrpc"
)

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("simapp [-d sdk_ip] [-p sdk_rpc_port] [-r receive_port] <app_id> <app_key>\n")
}

func main() {
	log.SetPrefix(os.Args[0])

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
		log.Printf("Dial error: %s", err.Error())
		return
	}
	defer func() {
		conn.Close()
	}()

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
		log.Printf("RPC error: %s", err.Error())
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)

	s := <-ch
	log.Println("Received signal:", s)
	argUnregist := sdkrpc.ArgUnregist{
		AppId:  appId,
		AppKey: appKey,
	}
	replyUnregist := sdkrpc.ReplyUnregist{}
	err = rpcClient.Call("SDK.Unregist", argUnregist, &replyUnregist)
}

func RunReceiverRPC(port int) {
	log.Printf("Starting Receiver RPC\n")
	laddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		log.Printf("Failed to start RPC server: %s", err.Error())
		return
	}

	receiver := receiverrpc.Receiver{}
	rpc.Register(&receiver)
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
		rpc.ServeConn(conn)
	}
}
