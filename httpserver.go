package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strconv"

	"github.com/garyburd/redigo/redis"
	"github.com/zhengzhiren/pushserver/tcpserver"
	"github.com/zhengzhiren/pushserver/packet"
)

func StartHttp() {
	addr := "localhost:8080"
	log.Printf("Starting HTTP server on %s\n", addr)
	http.HandleFunc("/", rootHandler)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf(err.Error())
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// output
	fmt.Fprintln(w, "\n", time.Now())
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Host: %q\n", r.Host)
	fmt.Fprintf(w, "URL: %q\n", r.URL.Path)
	fmt.Fprintf(w, "Method: %q\n", r.Method)
	fmt.Fprintf(w, "Proto: %q\n", r.Proto)

	// forms
	r.ParseForm()
	clientid := r.FormValue("clientid")
	msg := r.FormValue("msg")
	expire_str := r.FormValue("expire")
	var expire = 0
	var err error
	if expire_str != "" {
		expire, err = strconv.Atoi(expire_str)
		if err != nil {
			log.Printf("expire format error: %s", err.Error())
			return
		}
	}

	fmt.Fprintf(w, "\nForm values:\n")
	fmt.Fprintf(w, "clientid: %s\n", clientid)
	fmt.Fprintf(w, "msg: %s\n", msg)
	fmt.Fprintf(w, "expire: %d (s)\n", expire)

	// connect to Redis
	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Printf("Dial redix error: %s", err.Error())
	}

	// get new message Id
	n, err := redisConn.Do("INCR", "msg_id")
	if err != nil {
		log.Printf("Error on save message to redis: %s", err.Error())
		return
	}
	msg_id, ok := n.(int64)
	if (!ok) {
		log.Printf("Error on msg_id")
		return
	}
	fmt.Fprintf(w, "msd_id: %d\n", msg_id)

	// store in HashMap
	key := "msg:" + strconv.FormatInt(msg_id, 10)
	_, err = redisConn.Do("HMSET", key, "msg", msg, "clientid", clientid, "expire", expire)
	if err != nil {
		log.Printf("Error on saving message to redis: %s", err.Error())
		return
	}

	pktDataMsg := packet.PktDataMessage {
		Msg : msg,
	}

	var pkt *packet.Pkt
	pkt, err = packet.Pack(packet.PKT_Push, 0, &pktDataMsg)
	if err != nil {
		log.Printf("Error on pack message: %s", err.Error())
		return
	}

	clientCount := 0
	tcpserver.ClientMapLock.RLock();
	fmt.Fprintf(w, "\nClients online: %d\n", len(tcpserver.ClientMap))
	for _, client := range tcpserver.ClientMap {
		fmt.Fprintf(w, "Client Id: %s\n", client.Id);
		if clientid == "" || clientid == client.Id {
			client.PktChan <- pkt
			fmt.Fprintf(w, "Message has been pushed to %s\n", client.Conn.RemoteAddr().String())
			clientCount++
		}
	}
	tcpserver.ClientMapLock.RUnlock();
	fmt.Fprintf(w, "\nMessage has been pushed to %d clients\n", clientCount)
}
