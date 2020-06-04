package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type ReceiverLayerMessage struct {
	Message  interface{}
	Channels []string
	Groups   []string // todo
}

func (r *ReceiverLayerMessage) SerializerMessage(messageType interface{}) error {
	resValueOf := reflect.ValueOf(messageType)
	isPtr := resValueOf.Kind() == reflect.Ptr
	if resValueOf.Kind() == reflect.Struct {
		return errors.New("need ptr")
	}

	switch value := r.Message.(type) {
	case *json.RawMessage:
		if isPtr {
			if err := json.Unmarshal(*value, messageType); err != nil {
				return err
			}
		} else {
			if err := json.Unmarshal(*value, &messageType); err != nil {
				fmt.Println(err)
				return err
			}
		}
	default:
		messageType = value
	}
	if isPtr {
		r.Message = resValueOf.Elem().Interface()
	} else {
		r.Message = messageType
	}
	return nil
}
