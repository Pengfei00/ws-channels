package main

import (
	"context"
	"net/http"
	"time"
	"ws-channels/config"
	"ws-channels/core"
)

func onConnect(resp http.ResponseWriter, req *http.Request, client *core.Client, next func(channelName string) error) {
	err := next("")
	if err != nil {
		return
	}
	_ = client.GroupAdd("all")

}
func onDisconnect(code int, reason string, c *core.Client) {
	_ = c.GroupDiscard("all")
}
func onMessage(messageType int, data []byte, from int, c *core.Client) {
	if from == core.FromLocal {
		_ = c.GroupSend(messageType, data, "all")
	} else {
		c.Send(messageType, data)
	}
}

func main() {
	c := config.Config{
		Layer: config.RedisLayer,
		RedisConfig: &config.RedisConfig{
			Addr:        "127.0.0.1:6379",
			Password:    "",
			DB:          0,
			MaxIdle:     10,
			IdleTimeout: 60 * time.Second,
			Wait:        true,
		},
	}

	server := core.NewServer(&c, context.Background(), onConnect, onDisconnect, onMessage)

	server.Run()

	http.HandleFunc("/ws", server.Handler)
	_ = http.ListenAndServe("0.0.0.0:7777", nil)
}
