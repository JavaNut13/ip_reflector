package main

import (
	"bytes"
	"encoding/json"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
  "fmt"
)

const (
	PUSH_URL  = "https://api.pushbullet.com/v2/pushes"
	ORIGIN    = "http://pushbullet.com/"
	SOCKET    = "wss://stream.pushbullet.com/websocket/"
	IP_LOOKUP = "https://freegeoip.net/json/"
)

type Push struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	DeviceId string `json:"source_device_iden"`
}

type Container struct {
	Pushes []Push `json:"pushes"`
}

type IpResponse struct {
	Ip string `json:"ip"`
}

func WatchSocket(origin, url string, wait int, callback func([]byte)) {
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Fatal(err)
	}
	var msg = make([]byte, 512)
	var n int
	for {
		if n, err = ws.Read(msg); err != nil {
			log.Fatal(err)
		} else {
      if string(msg[:n]) != "{\"type\": \"nop\"}" {
  			callback(msg)
    		log.Printf("%s\n", msg[:n])
      }
		}
		time.Sleep(time.Duration(wait) * time.Second)
	}
}

func GetPushes(token string, wait int) []Push {
	client := &http.Client{}
	t := strconv.Itoa(int(time.Now().Unix()) - (2 * wait))
	req, _ := http.NewRequest("GET", PUSH_URL+"?modified_after="+t, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := client.Do(req)

	decoder := json.NewDecoder(res.Body)

	var container Container

	err := decoder.Decode(&container)
	if err != nil {
		panic(err)
	}
	return container.Pushes
}

func GetIp() string {
	res, err := http.Get(IP_LOOKUP)
	decoder := json.NewDecoder(res.Body)

	var ip IpResponse
	err = decoder.Decode(&ip)
	if err != nil {
		panic(err)
	}
	return ip.Ip
}

func DoPush(title, message, device, token string) {
	client := &http.Client{}
	body := "{\"type\": \"note\", \"title\": \"" + title + "\", \"body\": \"" + message + "\", \"device_iden\": \"" + device + "\"}"
	req, _ := http.NewRequest("POST", PUSH_URL, bytes.NewBuffer([]byte(body)))
	req.Header.Set("Authorization", "Bearer " + token)
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)
}

func ReflectIp(pushes []Push, token string) {
	for _, push := range pushes {
		if push.Title == "RPI" && push.Body == "reflect_ip" {
			ip := GetIp()
			DoPush("Reflected IP", ip, push.DeviceId, token)
      break
		}
	}
}

func main() {
  if len(os.Args) < 3 {
    fmt.Println("Useage: reflect_ip <API key> <Sleep time (Seconds)>")
    os.Exit(1)
  }
  token := os.Args[1]
  waitStr := os.Args[2]
  wait, err := strconv.ParseInt(waitStr, 10, 32)
  if err != nil || wait < 0 {
    fmt.Println("Sleep time must be a positive integer.")
    os.Exit(2)
  }
  
  call := func(b []byte) {
    ReflectIp(GetPushes(token, int(wait)), token)
  }
  WatchSocket(ORIGIN, SOCKET + token, int(wait), call)
}
