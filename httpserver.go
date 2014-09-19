package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"pushserver/tcpserver"
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
	var clientid, msg string = "", ""
	r.ParseForm()
	fmt.Fprintf(w, "\nForm values: %d\n", len(r.Form))
	for key, value := range r.Form {
		fmt.Fprintf(w, "%s: %s\n", key, value)
		switch key {
		case "clientid":
			clientid = value[0]
		case "msg":
			msg = value[0]
		default:

		}
	}
	log.Printf("clientid: %s, msg: %s\n", clientid, msg)

	clientCount := 0
	tcpserver.ClientMapLock.RLock();
	fmt.Fprintf(w, "\nClients online: %d\n", len(tcpserver.ClientMap))
	for _, client := range tcpserver.ClientMap {
		fmt.Fprintf(w, "Client Id: %s\n", client.Id);
		if clientid == "" || clientid == client.Id {
			client.Conn.Write([]byte(msg))
			fmt.Fprintf(w, "Message has been pushed to %s\n", client.Conn.RemoteAddr().String())
			clientCount++
		}
	}
	tcpserver.ClientMapLock.RUnlock();
	fmt.Fprintf(w, "\nMessage has been pushed to %d clients\n", clientCount)
}
