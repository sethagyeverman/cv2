package svc

import (
	"cv2/internal/config"

	"github.com/redis/go-redis/v9"
)

func newRedis(c config.Config) *redis.Client {
	// TODO: 从配置文件读取 Redis 配置
	return redis.NewClient(&redis.Options{
		Addr:     "192.168.1.35:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}
