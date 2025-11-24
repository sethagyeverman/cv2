package svc

import (
	"cv2/internal/config"
	"cv2/internal/ent"

	_ "github.com/go-sql-driver/mysql"
)

func newDB(c config.Config) (*ent.Client, error) {
	client, err := ent.Open(c.Storages.Driver, c.Storages.DSN)
	if err != nil {
		return nil, err
	}
	return client, nil
}
