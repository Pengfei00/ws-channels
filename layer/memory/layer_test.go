package memory

import (
	"context"
	"fmt"
	"testing"
	"ws-channels/config"
	"ws-channels/layer/common"
)

var layer *Layer = newLayer()
var groups []string = []string{"groupA", "groupB", "groupC"}
var channels []string = []string{layer.NewChannel(""), layer.NewChannel(""), layer.NewChannel("")}

func newLayer() *Layer {
	c := &config.MemoryConfig{}
	receiverMessage := make(chan common.ReceiverLayerMessage, 500)
	layer := NewLayer(receiverMessage, c)
	layer.Run(context.Background())
	return layer
}

func TestLayer(t *testing.T) {
	for i := 0; i < 2; i++ {
		GroupAdd(t)
		GetChannels(t)
		GroupSend(t)
		Send(t)
		GroupDiscard(t)
	}
}

func GroupAdd(t *testing.T) {
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
func GroupDiscard(t *testing.T) {
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

func GetChannels(t *testing.T) {
	if channels, err := layer.GetChannels(groups[0]); err != nil || len(channels) != 2 {
		t.Error(err)
	}
}
func GroupSend(t *testing.T) {
	if err := layer.GroupSend("测试消息", groups[0]); err != nil {
		t.Error(err)
	}
	d := <-layer.ReceiverMessage
	if len(d.Channels) != 2 || d.Message != "测试消息" {
		t.Error("error", d.Channels)
	}

	if err := layer.GroupSend("测试消息", groups...); err != nil {
		t.Error(err)
	}

	d = <-layer.ReceiverMessage
	if len(d.Channels) != 3 || d.Message != "测试消息" {
		t.Error("error", d.Channels)
	}

	if err := layer.GroupSend("测试消息", groups[1]); err != nil {
		t.Error(err)
	}
	d = <-layer.ReceiverMessage
	if len(d.Channels) != 1 {
		t.Error("error")
	}
}

func Send(t *testing.T) {
	data := "测试单发消息"
	if err := layer.Send(data, channels[0]); err != nil {
		t.Error("error")
	}
	d := <-layer.ReceiverMessage
	if d.Message != data || d.Channels[0] != channels[0] {
		t.Error("error")
	}

}

func BenchmarkLayer_GroupAdd(b *testing.B) {
	layer = newLayer()
	group := groups[0]
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		layer.GroupAdd(fmt.Sprintf("%d", i), group)
	}
}

func BenchmarkLayer_GroupSend(b *testing.B) {
	layer = newLayer()
	group := groups[0]
	layer.GroupAdd(channels[0], group)
	b.ResetTimer()
	i := 0
	for i = 0; i < b.N; i++ {
		layer.GroupSend("测试消息", group)
		<-layer.ReceiverMessage
	}
}
func BenchmarkLayer_GroupSend10000(b *testing.B) {
	layer = newLayer()
	group := groups[0]
	for i := 0; i < 100000; i++ {
		layer.GroupAdd(layer.NewChannel(""), group)
	}
	b.ResetTimer()
	i := 0
	for i = 0; i < b.N; i++ {
		layer.GroupSend("测试消息", group)
		<-layer.ReceiverMessage
	}
}

func BenchmarkLayer_GetChannels100000(b *testing.B) {
	layer = newLayer()
	group := groups[0]
	for i := 0; i < 100000; i++ {
		layer.GroupAdd(layer.NewChannel(""), group)
	}
	b.ResetTimer()
	i := 0
	for i = 0; i < b.N; i++ {
		layer.GetChannels(group)
	}
}
