package model

import (
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type App struct {
    Address	string
}

type AppService interface {
    GetAccount(ctx context.Context, query *bson.D) (*App, error)
    GetApps(ctx context.Context, query *bson.D, opts ...*options.FindOptions) ([]*App, error)
    SaveApp(ctx context.Context, in *App) error
}
