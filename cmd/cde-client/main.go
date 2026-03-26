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
	client := cde.NewClient()
	for {
		// 1s 自动重连
		run(&client)
		time.Sleep(1 * time.Second)
	}
}

func run(client *cde.Client) {
	host := os.Getenv("HOST")
	u := url.URL{Scheme: "wss", Host: host, Path: "/serve"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	defer c.Close()

	if err := client.Run(c); err != nil {
		log.Println("client run error:", err)
	}
}
