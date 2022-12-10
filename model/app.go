package model

import (
	"context"
)

type App struct {
	Address string `json:"address" bson:"address"`
}

type AppService interface {
	GetApps(ctx context.Context, pageNum, pageSize int64) (apps []*App, total int64, err error)
	GetApp(ctx context.Context, appID string) (*App, error)
}
