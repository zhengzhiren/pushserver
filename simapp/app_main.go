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
	deviceId string
	appId    string
	appKey   string
	token    string
	appAddr  *net.UnixAddr
)

func usage() {
	fmt.Printf("Usage:\n")
	fmt.Printf("simapp -t [token] <device_id> <app_id> <app_key>\n")
}

func main() {

	flag.StringVar(&token, "t", "", "user sso token")
	flag.Parse()
	if flag.NArg() != 3 {
		usage()
		return
	}

	deviceId := flag.Args()[0]
	appId := flag.Args()[1]
	appKey := flag.Args()[2]

	appDomain := "/tmp/simsdk_" + deviceId + "_simapp_" + appId
	deviceDomain := "/tmp/simsdk_" + deviceId
	var err error
	appAddr, err = net.ResolveUnixAddr("unix", appDomain)
	raddr, err := net.ResolveUnixAddr("unix", deviceDomain)
	if err != nil {
		log.Errorf("ResolveUnixAddr error: %s", err.Error())
		return
	}

	go RunReceiverRPC()

	// sleep to get RPC prepared
	time.Sleep(5 * time.Second)

	conn, err := net.DialUnix("unix", nil, raddr)
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
		AppId:   appId,
		AppKey:  appKey,
		AppAddr: appAddr,
		Token:   token,
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
		Token:  token,
	}
	replyUnregist := sdkrpc.ReplyUnregist{}
	err = rpcClient.Call("SDK.Unregist", argUnregist, &replyUnregist)
	os.Remove(appDomain)
}

func RunReceiverRPC() {
	log.Info("Starting Receiver RPC\n")
	ln, err := net.ListenUnix("unix", appAddr)
	if err != nil {
		log.Errorf("Failed to start RPC server: %v", err)
		return
	}

	Receiver = receiverrpc.Receiver{}
	rpc.Register(&Receiver)
	log.Infof("RPC server is listening on %s\n", appAddr.String())

	defer func() {
		// close the listener sock
		log.Info("Closing listener socket.\n")
		ln.Close()
	}()

	for {
		ln.SetDeadline(time.Now().Add(time.Second))
		conn, err := ln.AcceptUnix()
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
