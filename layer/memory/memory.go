package memory

import (
	"context"
	"sync"
	"ws-channels/config"
	"ws-channels/layer/common"
)

type Layer struct {
	GroupExpiry     int
	ReceiverMessage chan common.ReceiverLayerMessage

	groups *sync.Map
}

func (l Layer) Send(message interface{}, channels ...string) error {
	for _, channel := range channels {
		l.ReceiverMessage <- common.ReceiverLayerMessage{
			Message:  message,
			Channels: []string{channel},
		}
	}
	return nil
}

func (l Layer) getChanelsMap(group string) (map[string]bool, error) {
	if v, loaded := l.groups.Load(group); loaded {
		if value, ok := v.(map[string]bool); ok {
			return value, nil
		}
	}
	return map[string]bool{}, nil
}

func (l Layer) GetChannels(group string) ([]string, error) {
	if value, err := l.getChanelsMap(group); value != nil {
		result := make([]string, 0, len(value))
		for key, _ := range value {
			result = append(result, key)
		}
		return result, nil
	} else if err != nil {
		return []string{}, err
	} else {
		return []string{}, nil
	}

}

func (l Layer) GroupAdd(channel string, groups ...string) error {
	for _, group := range groups {
		if v, loaded := l.groups.LoadOrStore(group, map[string]bool{channel: true}); loaded {
			if value, ok := v.(map[string]bool); ok {
				value[channel] = true
			}
		}
	}

	return nil
}

func (l Layer) GroupDiscard(channel string, groups ...string) error {
	for _, group := range groups {
		if v, loaded := l.groups.Load(group); loaded {
			if value, ok := v.(map[string]bool); ok {
				delete(value, channel)
				if len(value) == 0 {
					l.groups.Delete(group)
				}
			}
		}
	}

	return nil
}

func (l Layer) GroupSend(message interface{}, groups ...string) error {
	channelsMap := make(map[string]bool)
	for _, group := range groups {
		channels, _ := l.getChanelsMap(group)
		for k, _ := range channels {
			channelsMap[k] = true
		}

	}
	channels := make([]string, 0, len(channelsMap))
	for k, _ := range channelsMap {
		channels = append(channels, k)
	}
	l.ReceiverMessage <- common.ReceiverLayerMessage{
		Message:  message,
		Channels: channels,
	}

	return nil
}

func (Layer) NewChannel(user string) string {
	if user == "" {
		user = common.RandomString(8)
	}
	return user
}
func (l Layer) Run(ctx context.Context) error {
	return nil
}

func NewLayer(ReceiverMessage chan common.ReceiverLayerMessage, c *config.MemoryConfig) *Layer {
	layer := &Layer{
		GroupExpiry:     86400,
		ReceiverMessage: ReceiverMessage,
		groups:          new(sync.Map),
	}

	return layer
}
