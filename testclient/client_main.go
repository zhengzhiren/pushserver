package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"encoding/json"

	"github.com/zhengzhiren/pushserver/simsdk/agent"
)


var (
	RegIds   = make(map[string]string)
	DeviceId = ""
)

func main() {
	log.SetPrefix("testclient ")

	dst := ""

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-i": // device id
			i++
			if i >= len(os.Args) {
				fmt.Printf("missing argument for \"-i\"\n")
				return
			}
			DeviceId = os.Args[i]
		case "-a": // appid
			i++
			if i >= len(os.Args) {
				fmt.Printf("missing argument for \"-a\"\n")
				return
			}
			RegIds[os.Args[i]] = ""
		default:
			if dst == "" {
				dst = os.Args[i]
			} else {
				fmt.Printf("unknown argument %s\n", os.Args[i])
				return
			}
		}
	}

	if dst == "" {
		fmt.Printf("no destination address\n")
		return
	}
	if len(RegIds) == 0 {
		fmt.Printf("no AppId on this device\n")
		return
	}
	if DeviceId == "" {
		rand.Seed(time.Now().Unix())
		DeviceId = strconv.Itoa(rand.Int() % 10000) // a random Id
	}
	LoadRegIds()

	raddr, err := net.ResolveTCPAddr("tcp", dst)
	if err != nil {
		log.Printf("Unknown address: %s", err.Error())
		return
	}

	agent := agent.NewAgent(DeviceId, raddr)
	go agent.Run()

	for appid, _ := range RegIds {
		agent.Regist(appid, appid+"key", RegIds[appid])
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL)
	for {
		s := <-ch
		log.Println("Received signal:", s)
		if len(RegIds) > 0 {
			for appid, regid := range RegIds {
				log.Printf("Unregist AppId: [%s], RegId: [%s]", appid, regid)
				agent.Unregist(appid, "tempkey", RegIds[appid])
				delete(RegIds, appid)
				break
			}
		} else {
			log.Printf("Stopped")
			break
		}
	}
}

func SaveRegIds() {
	file, err := os.OpenFile("RegIds.txt", os.O_RDWR | os.O_CREATE, 0666)
	if err != nil {
		log.Printf("OpenFile error: %s", err.Error())
		return
	}
	b, err := json.Marshal(RegIds)
	if err != nil {
		log.Printf("Marshal error: %s", err.Error())
		file.Close()
		return
	}
	file.Write(b)
	file.Close()
}

func LoadRegIds() {
	file, err := os.Open("RegIds.txt")
	if err != nil {
		log.Printf("Open error: %s", err.Error())
		return
	}
	buf := make([]byte, 1024)
	n, err := file.Read(buf)
	if err != nil {
		log.Printf("Read file error: %s", err.Error())
		file.Close()
		return
	}

	log.Printf("%s", buf)

	regIds := map[string]string {}
	err = json.Unmarshal(buf[:n], &regIds)
	if err != nil {
		log.Printf("Unarshal error: %s", err.Error())
		file.Close()
		return
	}
	for appid, _ := range RegIds {
		RegIds[appid] = regIds[appid]
	}

	log.Printf("RegIds: %s", RegIds)
	log.Printf("regIds: %s", regIds)
}
