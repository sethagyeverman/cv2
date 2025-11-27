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
	OAuth2 struct {
		Shiji struct {
			ClientID     string // OAuth2 客户端ID
			ClientSecret string // OAuth2 客户端密钥
		}
	}
	DefaultConfig struct {
		DefaultCoverImage string // 默认封面图URL
	}
	Pay struct {
		ServiceURL         string // 支付微服务地址
		BuySlotNotifyURL   string // 席位购买回调地址
		NotifySecret       string // 回调签名密钥
		OrderExpireMinutes int    // 订单超时时间（分钟）
	}
}
