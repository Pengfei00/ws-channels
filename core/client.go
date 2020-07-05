package core

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
	"ws-channels/common"
)

type Client struct {
	Channel string
	Req     *http.Request

	isClose  bool
	cancel   context.CancelFunc
	inChan   chan string
	outChan  chan common.Message
	wsSocket *websocket.Conn
	server   *Server
}

func (c *Client) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if messageType, data, err := c.wsSocket.ReadMessage(); err != nil {
				c.Close(websocket.CloseNormalClosure, err.Error())
			} else if c.server.OnMessage != nil {
				c.server.OnMessage(messageType, data, FromLocal, c)
			}
		}

	}

}

func (c Client) writeLoop(ctx context.Context) {
	for {
		select {
		case msg := <-c.outChan:
			if err := c.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				c.Close(websocket.CloseNormalClosure, err.Error())
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c Client) Send(messageType int, data []byte) {
	c.outChan <- common.Message{
		MessageType: messageType,
		Data:        data,
	}
}

func (c Client) GroupAdd(groups ...string) error {
	return c.server.GroupAdd(c.Channel, groups...)
}

func (c Client) GroupDiscard(groups ...string) error {
	return c.server.GroupDiscard(c.Channel, groups...)
}
func (c Client) GroupSend(messageType int, data []byte, groups ...string, ) error {
	return c.server.GroupSend(messageType, data, groups...)
}
