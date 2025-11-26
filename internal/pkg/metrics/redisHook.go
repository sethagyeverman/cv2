package metrics

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisHook Redis 命令钩子，用于记录 Redis 指标
type RedisHook struct{}

var _ redis.Hook = (*RedisHook)(nil)

func NewRedisHook() *RedisHook {
	return &RedisHook{}
}

func (h *RedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (h *RedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()

		err := next(ctx, cmd)

		duration := time.Since(start).Seconds()

		// 提取命令名称
		cmdName := strings.ToUpper(cmd.Name())

		// 记录耗时
		RedisDuration.WithLabelValues(cmdName).Observe(duration)

		// 记录错误
		if err != nil && err != redis.Nil {
			RedisErrors.WithLabelValues(cmdName).Inc()
		}

		return err
	}
}

func (h *RedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()

		err := next(ctx, cmds)

		duration := time.Since(start).Seconds()

		// Pipeline 命令统一记录为 PIPELINE
		RedisDuration.WithLabelValues("PIPELINE").Observe(duration)

		if err != nil {
			RedisErrors.WithLabelValues("PIPELINE").Inc()
		}

		return err
	}
}
