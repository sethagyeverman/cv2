# CV2 指标监控说明

## 配置

在 `etc/cv2.yaml` 中已启用 Prometheus：

```yaml
Prometheus:
  Host: 192.168.1.35
  Port: 9091
  Path: /metrics
```

访问地址：`http://192.168.1.35:9091/metrics`

## 已上报的指标

### 1. Go-Zero 默认指标

#### HTTP Server
- `http_server_requests_duration_ms` - HTTP 请求耗时（毫秒）
  - Labels: `path`, `method`, `code`
- `http_server_requests_code_total` - HTTP 响应状态码计数
  - Labels: `path`, `method`, `code`

#### Go Runtime
- `go_goroutines` - Goroutine 数量
- `go_memstats_alloc_bytes` - 已分配内存
- `go_gc_duration_seconds` - GC 耗时

### 2. 自定义指标

#### SQL (Ent)
- `cv2_sql_duration_seconds` - SQL 执行耗时
  - Labels: `operation`, `table`
- `cv2_sql_errors_total` - SQL 错误计数
  - Labels: `operation`, `table`

#### Redis
- `cv2_redis_duration_seconds` - Redis 命令耗时
  - Labels: `command`
- `cv2_redis_errors_total` - Redis 错误计数
  - Labels: `command`

#### 外部 API
- `cv2_external_api_duration_seconds` - 外部 API 调用耗时
  - Labels: `service`, `endpoint`, `status_code`
- `cv2_external_api_errors_total` - 外部 API 错误计数
  - Labels: `service`, `endpoint`

#### MongoDB
- `cv2_mongo_duration_seconds` - MongoDB 操作耗时
  - Labels: `operation`, `collection`
- `cv2_mongo_errors_total` - MongoDB 错误计数
  - Labels: `operation`, `collection`

#### MinIO
- `cv2_minio_duration_seconds` - MinIO 操作耗时
  - Labels: `operation` (upload/download/delete)
- `cv2_minio_errors_total` - MinIO 错误计数
  - Labels: `operation`

#### 业务指标
- `cv2_business_resume_generation_total` - 简历生成任务总数
  - Labels: `status` (success/failed/pending)
- `cv2_business_resume_generation_duration_seconds` - 简历生成耗时
- `cv2_business_article_views_total` - 文章访问量
  - Labels: `article_id`

## 监控实现

### Ent (数据库)
通过 `ent.Interceptor` 拦截所有查询和变更操作：
- 文件：`internal/pkg/metrics/ent_interceptor.go`
- 注册：`internal/svc/db.go`

### Redis
通过 `redis.Hook` 拦截所有 Redis 命令：
- 文件：`internal/pkg/metrics/redis_hook.go`
- 注册：`internal/svc/redis.go`

### 外部服务
需要在各个客户端中手动调用 metrics 记录：
```go
import "cv2/internal/pkg/metrics"

start := time.Now()
resp, err := client.Do(req)
duration := time.Since(start).Seconds()

metrics.ExternalAPIDuration.WithLabelValues("shiji", "/article/list", "200").Observe(duration)
if err != nil {
    metrics.ExternalAPIErrors.WithLabelValues("shiji", "/article/list").Inc()
}
```

## Prometheus 配置示例

在 Prometheus 的 `prometheus.yml` 中添加：

```yaml
scrape_configs:
  - job_name: 'cv2'
    static_configs:
      - targets: ['192.168.1.35:9091']
```

## Grafana 仪表盘

### 推荐面板

1. **HTTP 请求监控**
   - QPS: `rate(http_server_requests_code_total[1m])`
   - P99 延迟: `histogram_quantile(0.99, rate(http_server_requests_duration_ms_bucket[1m]))`
   - 错误率: `rate(http_server_requests_code_total{code=~"5.."}[1m])`

2. **数据库监控**
   - SQL 耗时: `cv2_sql_duration_seconds`
   - SQL 错误: `rate(cv2_sql_errors_total[1m])`

3. **Redis 监控**
   - 命令耗时: `cv2_redis_duration_seconds`
   - 命令 QPS: `rate(cv2_redis_duration_seconds_count[1m])`

4. **系统资源**
   - Goroutine: `go_goroutines`
   - 内存: `go_memstats_alloc_bytes`
   - GC 频率: `rate(go_gc_duration_seconds_count[1m])`

## 告警规则示例

```yaml
groups:
  - name: cv2_alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_server_requests_code_total{code=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "高错误率告警"

      - alert: SlowSQL
        expr: histogram_quantile(0.99, rate(cv2_sql_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "SQL 查询过慢"

      - alert: HighGoroutines
        expr: go_goroutines > 10000
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Goroutine 泄漏风险"
```

## 下一步

- [ ] 为外部 API 客户端（Algorithm、Shiji）添加指标记录
- [ ] 为 MongoDB 和 MinIO 客户端添加指标记录
- [ ] 添加业务指标（简历生成、文章访问等）
- [ ] 配置 Grafana 仪表盘
- [ ] 配置告警规则
