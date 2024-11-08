package di

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Sender   string `json:"sender"` 
}

var PatientConnections = make(map[*websocket.Conn]bool) 
var CustomerConnections = make(map[*websocket.Conn]bool) 
