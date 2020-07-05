package config

import "time"

type LayerEnum int

const (
	RedisLayer LayerEnum = 1
)

type Config struct {
	Layer       LayerEnum
	RedisConfig *RedisConfig
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
