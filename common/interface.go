package common

import "context"

type LayerInterface interface {
	GroupAdd(channel string, groups ...string) error
	GroupDiscard(channel string, groups ...string) error
	GroupSend(message Message, groups ...string) error
	Send(message Message, channels ...string) error
	GetChannels(group string) ([]string, error)
	NewChannel(user string) string
	Run(ctx context.Context) error
}
