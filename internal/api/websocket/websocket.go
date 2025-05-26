package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func checkOrigin(r *http.Request) bool {
	// if config.IsDevelopment() {
	// 	return true
	// }

	// origin := r.Header.Get("Origin")

	// origins := map[string]bool{
	// 	"http://localhost:3000": true,
	// 	"http://localhost:5000": true,
	// }

	// allowed, ok := origins[origin]
	// return allowed && ok
	return true
}

type Connection interface {
	WriteMessage(messageType int, data []byte) error
}

const TextMessage = websocket.TextMessage
