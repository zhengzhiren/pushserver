package main

import (
	"fmt"
	"net/rpc"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhengzhiren/pushserver/simsdk/sdkrpc"
)

func main() {
	log.SetPrefix(os.Args[0])

	var(
		rpcPort int
		ip string
	)

	flag.StringVar(&ip, "d", "127.0.0.1", "Dest IP")
	flag.IntVar(&rpcPort, "p", 9988, "Dest port for RPC")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Printf("need AppId and AppKey\n")
		return
	}

	appId := flag.Args()[0]
	appKey := flag.Args()[1]

	raddr := net.TCPAddr {
		IP: net.ParseIP(ip),
		Port: rpcPort,
	}
	if raddr.IP == nil {
		fmt.Printf("Invalid IP address\n")
		return
	}

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
	args := sdkrpc.ArgRegist{
		AppId: appId,
		AppKey: appKey,
	}
	reply := sdkrpc.ReplyRegist{
	}
	err = rpcClient.Call("SDK.Regist", args, &reply)
	if err != nil {
		log.Printf("RPC error: %s", err.Error())
		return
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)

	s := <-ch
	log.Println("Received signal:", s)
	//err = rpcClient.Call("SDK.Unregist", args, &reply)
}
