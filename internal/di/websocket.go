package di

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}
type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
		Sender   string `json:"sender"` // new field to identify sender
}

var P1atientConnections = make(map[*websocket.Conn]bool)  // Active patient connections
var CustomerConnections = make(map[*websocket.Conn]bool) // Active customer care connections
