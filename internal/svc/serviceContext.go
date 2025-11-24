// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"cv2/internal/config"
	"cv2/internal/ent"
	"cv2/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config config.Config
	Auth   rest.Middleware
	Ent    *ent.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := newDB(c)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config: c,
		Auth:   middleware.NewAuthMiddleware().Handle,
		Ent:    client,
	}
}
