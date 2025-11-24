package svc

import (
	"cv2/internal/config"
	"cv2/internal/infra/mongo"
)

func newMongo(c config.Config) (*mongo.Client, error) {
	client, err := mongo.NewClient(c.Mongo.URI, c.Mongo.Database)
	if err != nil {
		return nil, err
	}
	return client, nil
}
