package redis

import (
	"encoding/json"
	"fmt"
	"ws-channels/layer/common"
)

func MessageToData(receiverMessage common.ReceiverLayerMessage, result interface{}) error {
	switch value := receiverMessage.Message.(type) {
	case map[string]interface{}:
		body, _ := json.Marshal(value)
		return json.Unmarshal(body, result)
	default:
		fmt.Println(value)
		result = receiverMessage.Message
		return nil
	}
}
