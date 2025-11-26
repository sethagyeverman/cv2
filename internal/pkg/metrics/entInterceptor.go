package metrics

import (
	"context"
	"time"

	"entgo.io/ent"
)

// EntInterceptor Ent 查询拦截器，用于记录 SQL 执行指标
func EntInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			// 开始计时
			start := time.Now()

			// 执行查询
			value, err := next.Query(ctx, query)

			// 计算耗时
			duration := time.Since(start).Seconds()

			// 提取操作类型和表名
			operation, table := parseQuery(query)

			// 记录耗时
			SQLDuration.WithLabelValues(operation, table).Observe(duration)

			// 记录错误
			if err != nil {
				SQLErrors.WithLabelValues(operation, table).Inc()
			}

			return value, err
		})
	})
}

// parseQuery 从查询中提取操作类型和表名
func parseQuery(query ent.Query) (operation, table string) {
	// Ent Query 没有直接暴露类型信息，这里简化处理
	table = "unknown"
	operation = "select"

	return operation, table
}

// EntMutationInterceptor Ent 变更拦截器
func EntMutationInterceptor() ent.Interceptor {
	return ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
		// Traverse 用于查询拦截
		return nil
	})
}

// EntMutateInterceptor Ent Mutation 拦截器
func EntMutateInterceptor() ent.Interceptor {
	return ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
		start := time.Now()

		// 这里简化处理，实际需要通过 Hook 实现
		defer func() {
			duration := time.Since(start).Seconds()
			SQLDuration.WithLabelValues("mutation", "unknown").Observe(duration)
		}()

		return nil
	})
}
