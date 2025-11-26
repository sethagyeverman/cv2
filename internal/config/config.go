package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Storages struct {
		Driver string
		DSN    string
	}
	Mongo struct {
		URI      string
		Database string
	}
	MinIO struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		BucketName      string
	}
	Algorithm struct {
		GenerateURL string // 简历生成服务 URL
		DataURL     string // 数据服务 URL
		ScoreURL    string // 评分服务 URL
	}
	JWT struct {
		SecretKey        string // JWT 密钥
		TokenRefreshURL  string // Token 刷新地址
		RefreshThreshold int    // Token 刷新阈值（分钟）
	}
	Shiji struct {
		BaseURL string // 世纪服务器地址
	}
}
