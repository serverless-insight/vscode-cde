package main

import (
	"log"
	"net/url"
	"vscode-cde/pkg/cde"

	"github.com/gorilla/websocket"
)

// const HOST = "localhost:9000"

// const HOST = "cde-server-edeirgdpjq.cn-hangzhou.fcapp.run"
// const HOST = "cde-begxigkief.cn-hongkong.fcapp.run"
const HOST = "1270939-proxy-9000.dsw-gateway-cn-hangzhou.data.aliyun.com"

func main() {
	u := url.URL{Scheme: "wss", Host: HOST, Path: "/serve"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	client := cde.NewClient()
	client.Run(c)
}
