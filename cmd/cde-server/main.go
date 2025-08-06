package main

import (
	"log"
	"net/http"
	"vscode-cde/pkg/cde"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serve(w http.ResponseWriter, r *http.Request) {
	server := cde.NewServer()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	server.OnConnection(conn)
}

func main() {
	http.HandleFunc("/serve", serve)
	log.Println("Server started on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
