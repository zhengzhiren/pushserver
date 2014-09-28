package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zhengzhiren/pushserver/tcpserver"
)

func main() {
	var port int
	var httpPort int
	flag.IntVar(&port, "p", 9233, "Port for Push server")
	flag.IntVar(&httpPort, "P", 9234, "Port for http server")
	flag.Parse()

	tcpServer := tcpserver.Create(port)
	go tcpServer.Start()

	// HTTP server
	go StartHttp(httpPort)

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	s := <-ch
	log.Println("Exiting. Received signal:", s)
	tcpServer.Stop()
}
