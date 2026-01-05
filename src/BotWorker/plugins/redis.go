package plugins

import (
	"botworker/internal/redis"
)

var GlobalRedis *redis.Client

func SetGlobalRedis(client *redis.Client) {
	GlobalRedis = client
}
