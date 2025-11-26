package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// SQL 执行耗时
	SQLDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "sql",
			Name:      "duration_seconds",
			Help:      "SQL execution duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation", "table"},
	)

	// SQL 错误计数
	SQLErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "sql",
			Name:      "errors_total",
			Help:      "Total number of SQL errors",
		},
		[]string{"operation", "table"},
	)

	// Redis 命令耗时
	RedisDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "redis",
			Name:      "duration_seconds",
			Help:      "Redis command duration in seconds",
			Buckets:   []float64{.0001, .0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"command"},
	)

	// Redis 错误计数
	RedisErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "redis",
			Name:      "errors_total",
			Help:      "Total number of Redis errors",
		},
		[]string{"command"},
	)

	// 外部 API 调用耗时
	ExternalAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "external_api",
			Name:      "duration_seconds",
			Help:      "External API call duration in seconds",
			Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10, 30},
		},
		[]string{"service", "endpoint", "status_code"},
	)

	// 外部 API 错误计数
	ExternalAPIErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "external_api",
			Name:      "errors_total",
			Help:      "Total number of external API errors",
		},
		[]string{"service", "endpoint"},
	)

	// 业务指标：简历生成任务
	ResumeGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "business",
			Name:      "resume_generation_total",
			Help:      "Total number of resume generation tasks",
		},
		[]string{"status"}, // success, failed, pending
	)

	// 业务指标：简历生成耗时
	ResumeGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "business",
			Name:      "resume_generation_duration_seconds",
			Help:      "Resume generation duration in seconds",
			Buckets:   []float64{1, 5, 10, 30, 60, 120, 300},
		},
	)

	// 业务指标：文章访问量
	ArticleViews = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "business",
			Name:      "article_views_total",
			Help:      "Total number of article views",
		},
		[]string{"article_id"},
	)

	// MongoDB 操作耗时
	MongoDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "mongo",
			Name:      "duration_seconds",
			Help:      "MongoDB operation duration in seconds",
			Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"operation", "collection"},
	)

	// MongoDB 错误计数
	MongoErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "mongo",
			Name:      "errors_total",
			Help:      "Total number of MongoDB errors",
		},
		[]string{"operation", "collection"},
	)

	// MinIO 操作耗时
	MinIODuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cv2",
			Subsystem: "minio",
			Name:      "duration_seconds",
			Help:      "MinIO operation duration in seconds",
			Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"operation"}, // upload, download, delete
	)

	// MinIO 错误计数
	MinIOErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "cv2",
			Subsystem: "minio",
			Name:      "errors_total",
			Help:      "Total number of MinIO errors",
		},
		[]string{"operation"},
	)
)
