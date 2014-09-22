package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhengzhiren/pushserver/tcpserver"
)

func main() {
	tcpServer := tcpserver.Create()
	go tcpServer.Start()

	// HTTP server
	go StartHttp()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	s := <-ch
	log.Println("Exiting. Received signal:", s)
	tcpServer.Stop()
}
