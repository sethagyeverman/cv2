package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewClient 创建 MongoDB 客户端
func NewClient(uri, database string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &Client{
		client:   client,
		database: client.Database(database),
	}, nil
}

// Close 关闭 MongoDB 连接
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// Database 获取数据库实例
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection 获取集合
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// ResumeContentCollection 获取简历内容集合
func (c *Client) ResumeContentCollection() *mongo.Collection {
	return c.Collection("resume_content")
}
