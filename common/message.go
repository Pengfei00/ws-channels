package common

type Message struct {
	MessageType int    `json:"message_type"`
	Data        []byte `json:"data"`
}

type ReceiverLayerMessage struct {
	Message  Message  `json:"message"` //  Message struct
	Channels []string `json:"channels"`
	Groups   []string `json:"groups"` // todo
}
