package model

import (
    "go.mongodb.org/mongo-driver/bson"
)

type App struct {
    Address	string
}

type AppService interface {
    GetAccount(ctx context.Context, query *bson.D) (*App, error)
    GetApps(ctx context.Context, query *bson.D) ([]*App, error)
    SaveApp(ctx context.Context, in *App) error
}
