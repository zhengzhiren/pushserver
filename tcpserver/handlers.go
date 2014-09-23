package tcpserver

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(conn *net.TCPConn, pkt *packet.Pkt)

func HandleRegist(conn *net.TCPConn, pkt *packet.Pkt) *Client {
	log.Printf("Handling packet type: Regist")
	var dataRegist = packet.PktDataRegist{}
	err := packet.Unpack(pkt, &dataRegist)
	if err != nil {
		log.Printf("Failed to Unpack: %s", err.Error())
		return nil
	}

	// TODO: check if the Id already exist

	client := Client{
		Conn:          conn,
		Id:            dataRegist.DevId,
		PktChan:       make(chan *packet.Pkt),
		AppIds:        dataRegist.AppIds,
		LastHeartbeat: time.Now(),
	}

	log.Printf("New Device is online, Id: %s, AppIds: %s", client.Id, client.AppIds)

	go func() {
		for {
			select {
			case pkt := <-client.PktChan:
				log.Printf("pkt to send")
				b, err := pkt.Serialize()
				if err != nil {
					log.Printf("Error on serializing pkt: %s", err.Error())
					continue
				}
				var n int
				n, err = conn.Write(b)
				if err != nil {
					log.Printf("Error on sending pkt: %s", err.Error())
					continue
				}
				log.Printf("Write successfully: %d", n)
			}
		}
	}()

	ClientMapLock.Lock()
	ClientMap[client.Id] = &client
	ClientMapLock.Unlock()

	SendOfflineMsg(&client)

	return &client
}

func SendOfflineMsg(client *Client) {
	// connect to Redis
	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Printf("Dial redix error: %s", err.Error())
		return
	}

	// get the timestamp
	var reply []interface{}
	reply, err = redis.Values(redisConn.Do("TIME"))
	if err != nil {
		log.Printf("Error on TIME: %s", err.Error())
		return
	}
	var current_time int64
	_, err = redis.Scan(reply, &current_time)

	// get offline message for each App on this device
	for _, appid := range client.AppIds {
		key := "broadcast_msg:" + appid
		reply, err = redis.Values(redisConn.Do("ZRANGE", key, 0, -1))
		if err != nil {
			log.Printf("Error on ZRANGE: %s", err.Error())
			goto Out
		}
		var msg_id int64
		for len(reply) > 0 {
			reply, err = redis.Scan(reply, &msg_id)
			if err != nil {
				log.Printf("Error on Scan ZRANGE reply: %s", err.Error())
				goto Out
			}
			log.Printf("offline msg_id: %d", msg_id)
			key = "msg:" + strconv.FormatInt(msg_id, 10)
			var reply_msg []interface{}
			reply_msg, err = redis.Values(redisConn.Do("HMGET", key, "msg", "expire_time"))
			if err != nil {
				log.Printf("Error on HMGET: %s", err.Error())
				goto Out
			}
			var msg string
			var expire_time int64
			_, err = redis.Scan(reply_msg, &msg, &expire_time)
			if err != nil {
				log.Printf("Error on Scan HMGET reply: %s", err.Error())
				goto Out
			}
			log.Printf("expire_time: %d, msg: %s", expire_time, msg)
			if expire_time > current_time {
				// message hasn't expired, need to send it
				client.SendMsg(msg)
			}
		}
	}

Out:
	redisConn.Close()
}

func HandleHeartbeat(conn *net.TCPConn, pkt *packet.Pkt) {
}

func HandleACK(conn *net.TCPConn, pkt *packet.Pkt) {
}
