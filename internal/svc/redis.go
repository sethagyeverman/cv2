package svc

import (
	"cv2/internal/config"
	"cv2/internal/pkg/metrics"

	"github.com/redis/go-redis/v9"
)

func newRedis(c config.Config) *redis.Client {
	// TODO: 从配置文件读取 Redis 配置
	client := redis.NewClient(&redis.Options{
		Addr:     "192.168.1.35:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// 添加 Hook 监控
	client.AddHook(metrics.NewRedisHook())

	return client
}
