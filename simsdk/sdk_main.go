package main

import (
	"flag"
	"fmt"
	//"log"
	"math/rand"
	"net"
	"net/rpc"
	//"os"
	"strconv"
	"time"

	log "github.com/cihub/seelog"

	"github.com/zhengzhiren/pushserver/simapp/receiverrpc"
	"github.com/zhengzhiren/pushserver/simsdk/agent"
	"github.com/zhengzhiren/pushserver/simsdk/sdkrpc"
)

var (
	DeviceId = ""
	RpcPort  int
	sdk      *sdkrpc.SDK
)

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("simsdk <-i device_id> [-p rpc_port] <ip:port>\n")
}

func main() {
	flag.StringVar(&DeviceId, "i", "", "Device Id of this simsdk")
	flag.IntVar(&RpcPort, "p", 9988, "RPC listen port for App")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Printf("missing ip:port\n")
		usage()
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
		log.Errorf("Unknown address: %v", err)
		return
	}
	agent := agent.NewAgent(DeviceId, raddr)
	agent.OnReceiveMsg = OnReceiveMsg
	agent.OnRegResponse = OnRegResponse
	go agent.Run()

	//go RunRPC(RpcPort)
	RunRPC(RpcPort, agent)

	//	ch := make(chan os.Signal)
	//	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	//	s := <-ch
}

func RunRPC(port int, agent *agent.Agent) {
	log.Info("Starting RPC server\n")
	laddr := net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		log.Errorf("Failed to start RPC server: %v", err)
		return
	}

	sdk = &sdkrpc.SDK{
		Agent:     agent,
		Receivers: make(map[string]*rpc.Client),
	}
	rpc.Register(sdk)
	log.Infof("RPC server is listening on %v\n", laddr.String())

	defer func() {
		// close the listener sock
		log.Debug("Closing listener socket.\n")
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
			log.Errorf("Failed to accept: %s", err.Error())
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func OnRegResponse(appId string, regId string, result int) {
	if result != 0 {
		log.Warnf("RegResponse error. AppId: %s, RegId: %s, Result: %d", appId, regId, result)
		return
	}
	client := sdk.Receivers[appId]
	if client == nil {
		log.Errorf("No client has AppId: %s", appId)
		return
	}
	arg := receiverrpc.ArgOnRegResp{
		AppId:  appId,
		RegId:  regId,
		Result: result,
	}
	reply := receiverrpc.ReplyOnRegResp{}
	err := client.Call("Receiver.OnRegResp", arg, &reply)
	if err != nil {
		log.Errorf("RPC error OnRegResp [AppId: %s]. %v", appId, err)
	}
}

func OnReceiveMsg(appId string, msgId int64, msgType int, msg string) {
	client := sdk.Receivers[appId]
	if client == nil {
		log.Errorf("Received msg bug no client has AppId: %s", appId)
		return
	}
	arg := receiverrpc.ArgOnReceiveMsg{
		MsgId:   msgId,
		MsgType: msgType,
		Msg:     msg,
	}
	reply := receiverrpc.ReplyOnReceiveMsg{}
	err := client.Call("Receiver.OnReceiveMsg", arg, &reply)
	if err != nil {
		log.Errorf("Failed to send message [AppId: %s, MsgId: %d]. %v", appId, msgId, err)
	}
}
