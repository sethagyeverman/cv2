package svc

import (
	"context"
	"cv2/internal/config"
	"cv2/internal/infra/ent"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func newDB(c config.Config) (*ent.Client, error) {
	client, err := ent.Open(c.Storages.Driver, c.Storages.DSN)
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表结构
	// 如果表不存在，会创建新表
	// 如果表已存在，会更新表结构以匹配 schema 定义（添加新列、索引等）
	// 注意：不会删除已存在的列或数据
	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
