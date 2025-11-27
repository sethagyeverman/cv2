package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
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

// Endpoint 获取 endpoint
func (c *Client) Endpoint() string {
	return c.client.EndpointURL().String()
}

// Upload 上传文件
func (c *Client) Upload(ctx context.Context, objectName string, fileHeader *multipart.FileHeader) error {
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	_, err = c.client.PutObject(ctx, c.bucketName, objectName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: getContentType(fileHeader.Filename),
	})
	return err
}

// UploadBytes 上传字节数据
func (c *Client) UploadBytes(ctx context.Context, objectName string, data []byte, contentType string) error {
	reader := bytes.NewReader(data)
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// UploadReader 上传 Reader
func (c *Client) UploadReader(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := c.client.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// GetPublicURL 获取公开访问 URL（需要 bucket 设置为公开或使用预签名）
func (c *Client) GetPublicURL(objectName string) string {
	return fmt.Sprintf("%s/%s/%s", c.client.EndpointURL().String(), c.bucketName, objectName)
}

// getContentType 根据文件名获取 Content-Type
func getContentType(filename string) string {
	switch {
	case len(filename) > 4 && filename[len(filename)-4:] == ".pdf":
		return "application/pdf"
	case len(filename) > 4 && filename[len(filename)-4:] == ".doc":
		return "application/msword"
	case len(filename) > 5 && filename[len(filename)-5:] == ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case len(filename) > 4 && filename[len(filename)-4:] == ".png":
		return "image/png"
	case len(filename) > 4 && filename[len(filename)-4:] == ".jpg":
		return "image/jpeg"
	case len(filename) > 5 && filename[len(filename)-5:] == ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

// GenerateObjectKey 生成 MinIO object key
func GenerateObjectKey(filename string) string {
	now := time.Now()
	return fmt.Sprintf("%d/%02d/%02d/%s",
		now.Year(), now.Month(), now.Day(), filename)
}
