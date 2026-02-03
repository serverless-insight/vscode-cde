package main

import (
	"log"
	"net/url"
	"os"
	"time"
	"vscode-cde/pkg/cde"

	"github.com/gorilla/websocket"
)

func main() {
	for {
		// 1s 自动重连
		run()
		time.Sleep(1 * time.Second)
	}
}

func run() {
	host := os.Getenv("HOST")
	u := url.URL{Scheme: "wss", Host: host, Path: "/serve"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	client := cde.NewClient()
	client.Run(c)
}
