package main

import (
	"log"
	"net/http"
	"vscode-cde/pkg/cde"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  64 * 1024,
	WriteBufferSize: 64 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serve(w http.ResponseWriter, r *http.Request) {
	log.Println("serve")
	server := cde.NewServer()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()
	defer func() {
		if err := recover(); err != nil {
			log.Println("error serve connection: ", err)
		}
	}()

	server.OnConnection(conn)
}

func main() {
	http.HandleFunc("/serve", serve)
	log.Println("Server started on :5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
