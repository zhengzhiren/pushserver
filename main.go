package main

import (
	"log"
	"os"
	"flag"
	"os/signal"
	"syscall"

	"github.com/zhengzhiren/pushserver/tcpserver"
)

func main() {
	var port int
	flag.IntVar(&port, "p", 9233, "Port for Push server")
	flag.Parse()

	tcpServer := tcpserver.Create(port)
	go tcpServer.Start()

	// HTTP server
	go StartHttp()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	s := <-ch
	log.Println("Exiting. Received signal:", s)
	tcpServer.Stop()
}
