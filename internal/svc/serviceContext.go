package svc

import (
	"cv2/internal/config"
	"cv2/internal/infra/algorithm"
	"cv2/internal/infra/ent"
	"cv2/internal/infra/minio"
	"cv2/internal/infra/mongo"
	"cv2/internal/middleware"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config    config.Config
	Auth      rest.Middleware
	Ent       *ent.Client
	Mongo     *mongo.Client
	MinIO     *minio.Client
	Redis     *redis.Client
	Algorithm *algorithm.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	client, err := newDB(c)
	if err != nil {
		panic(err)
	}

	mongoClient, err := newMongo(c)
	if err != nil {
		panic(err)
	}

	minioClient, err := newMinIO(c)
	if err != nil {
		panic(err)
	}

	redisClient := newRedis(c)
	algClient := algorithm.NewClient(c.Algorithm.BaseURL)

	return &ServiceContext{
		Config:    c,
		Auth:      middleware.NewAuthMiddleware().Handle,
		Ent:       client,
		Mongo:     mongoClient,
		MinIO:     minioClient,
		Redis:     redisClient,
		Algorithm: algClient,
	}
}
