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
		BaseURL string
	}
}
