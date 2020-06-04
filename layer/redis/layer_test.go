package redis

import (
	"context"
	"testing"
	"ws-channels/config"
	"ws-channels/layer/common"
)

type msg struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}

func newLayer() *Layer {
	c := &config.RedisConfig{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}
	receiverMessage := make(chan common.ReceiverLayerMessage, 50)
	layer := NewLayer(receiverMessage, c)
	layer.Run(context.Background())
	return layer
}

func TestLayer(t *testing.T) {
	layer := newLayer()
	groups := []string{"groupA", "groupB", "groupC"}
	channels := []string{layer.NewChannel(""), layer.NewChannel(""), layer.NewChannel("")}
	for i := 0; i < 2; i++ {
		NoneLocalName(layer, t)
		GroupAdd(layer, channels, groups, t)
		GetChannels(layer, groups, t)
		GroupSend(layer, groups, t)
		Send(layer, channels, t)
		GroupDiscard(layer, channels, groups, t)
		layer.MustSendRemote = true
	}
}

func NoneLocalName(layer *Layer, t *testing.T) {
	if layer.noneLocalName("aaa") != "aaa" {
		t.Error("error")
	}
	if layer.noneLocalName("aaa!bbb") != "aaa" {
		t.Error("error", )
	}
}

func GroupAdd(layer *Layer, channels, groups []string, t *testing.T) {
	if err := layer.GroupAdd(channels[0], groups...); err != nil {
		t.Error("error")
	}
	if err := layer.GroupAdd(channels[1], groups[0]); err != nil {
		t.Error("error")
	}
	if err := layer.GroupAdd(channels[2], groups[2]); err != nil {
		t.Error("error")
	}
}
func GroupDiscard(layer *Layer, channels, groups []string, t *testing.T) {
	if err := layer.GroupDiscard(channels[0], groups...); err != nil {
		t.Error("error")
	}
	if channels, err := layer.GetChannels(groups[0]); err != nil || len(channels) != 1 {
		t.Error("error", channels)
	}
	if err := layer.GroupDiscard(channels[1], groups[0]); err != nil {
		t.Error("error")
	}
	if channels, err := layer.GetChannels(groups[0]); err != nil || len(channels) != 0 {
		t.Error("error")
	}
	if err := layer.GroupDiscard(channels[2], groups[2]); err != nil {
		t.Error("error")
	}
	if channels, err := layer.GetChannels(groups[2]); err != nil || len(channels) != 0 {
		t.Error("error")
	}
}

func GetChannels(layer *Layer, groups []string, t *testing.T) {
	if channels, err := layer.GetChannels(groups[0]); err != nil || len(channels) != 2 {
		t.Error(err)
	}
}

func GroupSend(layer *Layer, groups []string, t *testing.T) {
	{
		sendData := msg{
			Id:      1,
			Content: "测试群发结构体",
		}
		if err := layer.GroupSend(sendData, groups[0]); err != nil {
			t.Error(err)
		}
		d := <-layer.ReceiverMessage
		d.SerializerMessage(&msg{})
		if len(d.Channels) != 2 || d.Message != sendData {
			t.Error("error", d.Channels)
		}
	}
	{
		sendData := "测试文本消息"
		if err := layer.GroupSend(sendData, groups...); err != nil {
			t.Error(err)
		}
		d := <-layer.ReceiverMessage
		d.SerializerMessage("")

		if len(d.Channels) != 3 || d.Message != sendData {
			t.Error("error", d.Channels)
		}
	}

	{
		sendData := "测试文本消息"
		if err := layer.GroupSend(sendData, groups[1]); err != nil {
			t.Error(err)
		}
		d := <-layer.ReceiverMessage
		d.SerializerMessage("")
		if len(d.Channels) != 1 {
			t.Error("error")
		}
	}

}

func Send(layer *Layer, channels []string, t *testing.T) {
	data := "测试单发消息"
	if err := layer.Send(data, channels[0]); err != nil {
		t.Error("error")
	}
	d := <-layer.ReceiverMessage
	d.SerializerMessage("")
	if d.Message != data || d.Channels[0] != channels[0] {
		t.Error("error")
	}

}
