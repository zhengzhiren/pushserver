package main

import (
	"fmt"
	"flag"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"

	"github.com/zhengzhiren/pushserver/simsdk/agent"
	"github.com/zhengzhiren/pushserver/simsdk/sdkrpc"
)

var (
	RegIds   = make(map[string]string)
	DeviceId = ""
	RpcPort int
)

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

	raddr, err := net.ResolveTCPAddr("tcp", dst)
	if err != nil {
		log.Printf("Unknown address: %s", err.Error())
		return
	}
	agent := agent.NewAgent(DeviceId, raddr)
	go agent.Run()

	//go RunRPC(RpcPort)
	RunRPC(RpcPort, agent)

//	ch := make(chan os.Signal)
//	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
//	s := <-ch
}

func RunRPC(port int, agent *agent.Agent) {
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

	sdk := sdkrpc.SDK {
		Agent: agent,
	}
	rpc.Register(&sdk)
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
