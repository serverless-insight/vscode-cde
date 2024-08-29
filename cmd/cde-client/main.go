package main

import (
	"log"
	"net/url"
	"vscode-cde/internal/cde"

	"github.com/gorilla/websocket"
)

const HOST = "localhost:9000"

// const HOST = "cde-server-edeirgdpjq.cn-hangzhou.fcapp.run"

func main() {
	u := url.URL{Scheme: "ws", Host: HOST, Path: "/serve"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	client := cde.NewClient()
	client.Run(c)
}
