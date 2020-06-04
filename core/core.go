package core

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	FromLocal  = 1
	FromServer = 2
)

type Message struct {
	MessageType int    `json:"message_type"`
	Data        []byte `json:"data"`
}

var DefaultUpgrader = websocket.Upgrader{
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 5 * time.Second,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
