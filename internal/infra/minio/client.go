package minio

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client     *minio.Client
	bucketName string
}

// NewClient 创建 MinIO 客户端
func NewClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool, bucketName string) (*Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &Client{
		client:     minioClient,
		bucketName: bucketName,
	}, nil
}

// GetPresignedUploadURL 获取预签名上传 URL
func (c *Client) GetPresignedUploadURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedPutObject(ctx, c.bucketName, objectName, expires)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// GetPresignedDownloadURL 获取预签名下载 URL
func (c *Client) GetPresignedDownloadURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(ctx, c.bucketName, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

// GetPresignedPostPolicy 获取预签名 POST policy（用于表单上传）
func (c *Client) GetPresignedPostPolicy(ctx context.Context, objectName string, expires time.Duration, maxFileSize int64) (*minio.PostPolicy, map[string]string, error) {
	policy := minio.NewPostPolicy()

	// 设置 bucket
	policy.SetBucket(c.bucketName)

	// 设置对象名称
	policy.SetKey(objectName)

	// 设置过期时间
	policy.SetExpires(time.Now().UTC().Add(expires))

	// 设置文件大小限制（可选）
	if maxFileSize > 0 {
		policy.SetContentLengthRange(1, maxFileSize)
	}

	// 生成预签名 POST 表单数据
	url, formData, err := c.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, nil, err
	}

	formData["url"] = url.String()

	return policy, formData, nil
}

// BucketName 获取 bucket 名称
func (c *Client) BucketName() string {
	return c.bucketName
}
