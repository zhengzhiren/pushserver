package tcpserver

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/zhengzhiren/pushserver/packet"
)

type PktHandler func(client *Client, pkt *packet.Pkt)

func HandleInit(conn *net.TCPConn, pkt *packet.Pkt) *Client {
	log.Printf("Handling packet type: Regist")
	var dataInit = packet.PktDataInit{}
	err := packet.Unpack(pkt, &dataInit)
	if err != nil {
		log.Printf("Failed to Unpack: %s", err.Error())
		return nil
	}

	// TODO: check if the Id already exist

	client := Client{
		Conn:          conn,
		Id:            dataInit.DevId,
		PktChan:       make(chan *packet.Pkt),
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

	// send Response for the Init packet
	dataInitResp := packet.PktDataInitResp{}

	initRespPkt, err := packet.Pack(packet.PKT_Init_Resp, 0, dataInitResp)
	if err != nil {
		log.Printf("Pack error: %s", err.Error())
		return nil
	}

	b, err := initRespPkt.Serialize()
	if err != nil {
		log.Printf("Serialize error: %s", err.Error())
		return nil
	}
	conn.Write(b)

	ClientMapLock.Lock()
	ClientMap[client.Id] = &client
	ClientMapLock.Unlock()

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
				client.SendMsg(msg, appid)
			}
		}
	}

Out:
	redisConn.Close()
}

func HandleRegist(client *Client, pkt *packet.Pkt) {
	var dataRegist = packet.PktDataRegist{}
	err := packet.Unpack(pkt, &dataRegist)
	if err != nil {
		log.Printf("Failed to Unpack: %s", err.Error())
	}
	log.Printf("Device [%s] regist %s", client.Id, dataRegist.AppIds)
	client.AppIds = append(client.AppIds, dataRegist.AppIds...)
	SendOfflineMsg(client)
}

func HandleHeartbeat(client *Client, pkt *packet.Pkt) {
}

func HandleACK(client *Client, pkt *packet.Pkt) {
}
