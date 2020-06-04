package config

import "time"

type LayerEnum int

const (
	RedisLayer  LayerEnum = 1
	MemoryLayer           = 2
)

type Config struct {
	Layer        LayerEnum
	RedisConfig  *RedisConfig
	MemoryConfig *MemoryConfig
}

type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	MaxActive   int
	MaxIdle     int
	IdleTimeout time.Duration
	Wait        bool
}

type MemoryConfig struct {
}
