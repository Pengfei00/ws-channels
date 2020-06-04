package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"runtime/debug"
	"strings"
	"time"
	"ws-channels/config"
	"ws-channels/layer/common"
)

type sendLayerGroupMessage struct {
	Groups  []string    `json:"groups"`
	Message interface{} `json:"message"`
}

type Layer struct {
	GroupExpiry     int
	ReceiverTaskNum int
	SendTaskNum     int

	client []*redis.Pool

	ReceiverMessage chan common.ReceiverLayerMessage

	clientPrefix     string
	sendGroupMessage chan sendLayerGroupMessage

	MustSendRemote bool
}

func (layer Layer) GetChannels(group string) ([]string, error) {
	client := layer.getPool().Get()
	defer client.Close()
	key := layer.groupKey(group)
	return redis.Strings(client.Do("SMEMBERS", key))
}

func (layer Layer) GroupAdd(channel string, groups ...string) error {
	client := layer.getPool().Get()
	defer client.Close()

	for _, group := range groups {
		key := layer.groupKey(group)
		_, err := client.Do("SADD", key, channel)
		if err != nil {
			return err
		}
		_ = client.Send("EXPIRE", key, layer.GroupExpiry)
	}

	return nil
}

func (layer Layer) GroupDiscard(channel string, groups ...string) error {
	client := layer.getPool().Get()
	defer client.Close()
	for _, group := range groups {
		key := layer.groupKey(group)
		_, err := client.Do("SREM", key, channel)
		if err != nil {
			return err
		}
	}

	return nil
}

func (layer Layer) GroupSend(message interface{}, groups ...string) error {
	layer.sendGroupMessage <- sendLayerGroupMessage{
		Groups:  groups,
		Message: message,
	}
	return nil
}

func (layer *Layer) Run(ctx context.Context) error {
	if len(layer.client) < 1 {
		return errors.New("未配置redis")
	}
	for i := 0; i < layer.ReceiverTaskNum; i++ {
		go layer.receiverTask(ctx)
	}
	for i := 0; i < layer.ReceiverTaskNum; i++ {
		go layer.sendTask(ctx)
	}
	return nil
}

func (layer Layer) NewChannel(user string) string {
	if user == "" {
		user = common.RandomString(8)
	}
	return layer.clientPrefix + "!" + user
}

func (layer *Layer) newClient(config *config.RedisConfig) {
	layer.client = []*redis.Pool{
		{
			Dial: func() (conn redis.Conn, err error) {
				return redis.Dial("tcp", config.Addr, redis.DialPassword(config.Password), redis.DialDatabase(config.DB))
			},
			MaxActive:   config.MaxActive,
			MaxIdle:     config.MaxIdle,
			IdleTimeout: config.IdleTimeout,
			Wait:        config.Wait,
		},
	}

}

func (Layer) groupKey(group string) string {
	return "group:" + group
}

func (layer Layer) getPool() *redis.Pool {
	return layer.client[0]
}

func (layer Layer) Send(message interface{}, channels ...string) error {
	for _, channel := range channels {
		noneLocalName := layer.noneLocalName(channel)
		d := common.ReceiverLayerMessage{
			Message:  message,
			Channels: []string{channel},
		}
		layer.sendToRedis(nil, noneLocalName, d)
	}
	return nil
}

func (layer Layer) userKeysMessageToMap(channelMap []string, message interface{}) map[string]*common.ReceiverLayerMessage {
	result := make(map[string]*common.ReceiverLayerMessage)
	for _, channel := range channelMap {
		noneLocalName := layer.noneLocalName(channel)
		if _, ok := result[noneLocalName]; ok {
			result[noneLocalName].Channels = append(result[noneLocalName].Channels, channel)
		} else {
			result[noneLocalName] = &common.ReceiverLayerMessage{
				Message:  message,
				Channels: []string{channel},
			}
		}
	}
	return result
}

func (layer Layer) noneLocalName(channel string) string {
	if position := strings.Index(channel, "!"); position > -1 {
		return channel[:position]
	} else {
		return channel
	}
}

func (layer *Layer) receiverTask(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("receiver stop:", string(debug.Stack()))
			fmt.Println("重新启动")
			time.Sleep(10 * time.Second)
			go layer.receiverTask(ctx)
		}
	}()

	client := layer.getPool().Get()
	defer client.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			rawData, err := client.Do("BRPOP", layer.clientPrefix, 5)
			if rawData == nil {
				continue
			}
			data, err := redis.ByteSlices(rawData, err)
			if err != nil {
				continue
			}
			if data != nil {
				var object json.RawMessage
				msg := common.ReceiverLayerMessage{Message: &object}
				_ = json.Unmarshal(data[1], &msg)
				layer.ReceiverMessage <- msg
			}
		}
	}
}

func (layer Layer) sendToRedis(client redis.Conn, serverKey string, data common.ReceiverLayerMessage) {
	if client == nil {
		client = layer.getPool().Get()
		defer client.Close()
	}
	if serverKey != layer.clientPrefix && !layer.MustSendRemote {
		layer.ReceiverMessage <- data
	} else {
		d, _ := json.Marshal(data)
		_, _ = client.Do("LPUSH", serverKey, d)
		_, _ = client.Do("EXPIRE", serverKey, layer.GroupExpiry)
	}
}

func (layer *Layer) sendTask(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("send stop:", string(debug.Stack()))
			fmt.Println("重新启动")
			time.Sleep(10 * time.Second)
			go layer.sendTask(ctx)
		}
	}()

	client := layer.getPool().Get()
	defer client.Close()
	for {
		select {
		case data := <-layer.sendGroupMessage:
			var channelMap []string
			var err error
			groups := data.Groups
			message := data.Message
			keys := make([]interface{}, len(groups))
			for i := range groups {
				keys[i] = layer.groupKey(groups[i])
			}
			if len(keys) == 1 {
				channelMap, err = redis.Strings(client.Do("SMEMBERS", keys[0]))
			} else {
				channelMap, err = redis.Strings(client.Do("SUNION", keys...))
			}
			if err != nil {
				continue
			}
			for serverKey, data := range layer.userKeysMessageToMap(channelMap, message) {
				layer.sendToRedis(client, serverKey, *data)
			}
		case <-ctx.Done():
			return
		}
	}
}

func NewLayer(receiverMessage chan common.ReceiverLayerMessage, c *config.RedisConfig) *Layer {
	layer := &Layer{
		GroupExpiry:      86400,
		ReceiverTaskNum:  5,
		SendTaskNum:      5,
		client:           nil,
		clientPrefix:     common.RandomString(8),
		sendGroupMessage: make(chan sendLayerGroupMessage, 500),
		ReceiverMessage:  receiverMessage,
	}
	layer.newClient(c)

	return layer
}
