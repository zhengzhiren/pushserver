package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/zhengzhiren/pushserver/tcpserver"
)

func StartHttp(port int) {
	addr := "localhost:" + strconv.Itoa(port)
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

	// form values
	r.ParseForm()
	deviceid := r.FormValue("deviceid")
	msg := r.FormValue("msg")
	appid := r.FormValue("appid")
	expire_str := r.FormValue("expire")
	var expire int64 = 0
	var err error
	if expire_str != "" {
		expire, err = strconv.ParseInt(expire_str, 10, 64)
		if err != nil {
			log.Printf("expire format error: %s", err.Error())
			return
		}
	}

	fmt.Fprintf(w, "\nForm values:\n")
	fmt.Fprintf(w, "\tdeviceid: %s\n", deviceid)
	fmt.Fprintf(w, "\tmsg: %s\n", msg)
	fmt.Fprintf(w, "\tappid: %s\n", appid)
	fmt.Fprintf(w, "\texpire: %d (s)\n", expire)

	// connect to Redis
	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Printf("Dial redix error: %s", err.Error())
		return
	}

	defer func() {
		redisConn.Close()
	}()

	// get new message Id
	msg_id, err := redis.Int64(redisConn.Do("INCR", "msg_id"))
	if err != nil {
		log.Printf("Error on INCR msg_id: %s", err.Error())
		return
	}
	fmt.Fprintf(w, "msd_id: %d\n", msg_id)

	// get the timestamp
	var reply []interface{}
	reply, err = redis.Values(redisConn.Do("TIME"))
	if err != nil {
		log.Printf("Error on TIME: %s", err.Error())
		return
	}
	var create_time int64
	_, err = redis.Scan(reply, &create_time)
	if err != nil {
		log.Printf("Error on Scan TIME reply: %s", err.Error())
		return
	}
	expire_time := create_time + expire

	// store in HashMap
	key := "msg:" + strconv.FormatInt(msg_id, 10)
	_, err = redisConn.Do("HMSET", key,
		"msg", msg,
		"deviceid", deviceid,
		"appid", appid,
		"create_time", create_time,
		"expire_time", expire_time)

	if err != nil {
		log.Printf("Error on saving message to redis HMSET: %s", err.Error())
		return
	}

	if deviceid != "" {
		// message to one device
		key = "device:" + deviceid + ":" + appid
		_, err = redisConn.Do("RPUSH", key, msg_id)
		if err != nil {
			log.Printf("Error on saving message to redis RPUSH: %s", err.Error())
			return
		}
	} else {
		// broadcast message
		key = "broadcast_msg:" + appid
		_, err = redisConn.Do("ZADD", key, msg_id, msg_id)
		if err != nil {
			log.Printf("Error on saving message to redis ZADD: %s", err.Error())
			return
		}
	}

	// send the message to online devices
	sendCount := 0
	tcpserver.ClientMapLock.RLock()
	fmt.Fprintf(w, "\nOnline devices: %d\n", len(tcpserver.ClientMap))
	for _, client := range tcpserver.ClientMap {
		fmt.Fprintf(w, "Device Id: %s\n", client.Id)
		if deviceid == "" || deviceid == client.Id {
			for _, v := range client.AppIds {
				if appid == v {
					err = client.SendMsg(msg, appid)
					if err != nil {
						log.Printf("Error on sending message: %s", err.Error())
						return
					}
					fmt.Fprintf(w, "Message has been pushed to %s\n", client.Conn.RemoteAddr().String())
					sendCount++
					break
				}
			}
		}
	}
	tcpserver.ClientMapLock.RUnlock()
	fmt.Fprintf(w, "\nMessage has been pushed to %d clients\n", sendCount)
}
