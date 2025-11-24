package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Storages struct {
		Driver string
		DSN    string
	}
}
