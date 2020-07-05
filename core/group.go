package core

import (
	"sync"
)

type MemoryLayer struct {
	clients map[string]*Client
	groups  *sync.Map
}

func (l MemoryLayer) getChanelsMap(group string) (map[string]bool, error) {
	if v, loaded := l.groups.Load(group); loaded {
		if value, ok := v.(map[string]bool); ok {
			return value, nil
		}
	}
	return map[string]bool{}, nil
}

func (l MemoryLayer) GetChannels(group string) ([]string, error) {
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

func (l MemoryLayer) GroupAdd(channel string, groups ...string) error {
	for _, group := range groups {
		if v, loaded := l.groups.LoadOrStore(group, map[string]bool{channel: true}); loaded {
			if value, ok := v.(map[string]bool); ok {
				value[channel] = true
			}
		}
	}

	return nil
}

func (l MemoryLayer) GroupDiscard(channel string, groups ...string) error {
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
