package svc

import (
	"cv2/internal/config"
	"cv2/internal/infra/minio"
)

func newMinIO(c config.Config) (*minio.Client, error) {
	client, err := minio.NewClient(
		c.MinIO.Endpoint,
		c.MinIO.AccessKeyID,
		c.MinIO.SecretAccessKey,
		c.MinIO.UseSSL,
		c.MinIO.BucketName,
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
