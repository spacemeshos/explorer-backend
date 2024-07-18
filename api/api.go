package api

import (
	"context"
	"fmt"
	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spacemeshos/explorer-backend/api/handler"
	"github.com/spacemeshos/explorer-backend/api/router"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/sql"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Api struct {
	Echo *echo.Echo
}

func Init(db *sql.Database, dbClient storage.DatabaseClient, allowedOrigins []string, debug bool, layersPerEpoch int64,
	marshaler *marshaler.Marshaler) *Api {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &handler.ApiContext{
				Context:        c,
				Storage:        db,
				StorageClient:  dbClient,
				LayersPerEpoch: layersPerEpoch,
				Cache:          marshaler,
			}
			return next(cc)
		}
	})
	e.HideBanner = true
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
	}))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Info("%s [%d] - %s", time.Now().Format(time.RFC3339), c.Response().Status, c.Request().URL.Path)
			return nil
		},
	}))

	if debug {
		e.Debug = true
		e.Use(middleware.Logger())
	}

	router.Init(e)

	return &Api{
		Echo: e,
	}
}

func (a *Api) Run(address string) {
	log.Info("server is running. For exit <CTRL-c>")
	if err := a.Echo.Start(address); err != nil {
		log.Err(fmt.Errorf("server stopped: %s", err))
	}

	sysSignal := make(chan os.Signal, 1)
	signal.Notify(sysSignal, syscall.SIGINT, syscall.SIGTERM)

	s := <-sysSignal
	log.Info("exiting, got signal %v", s)
	if err := a.Shutdown(); err != nil {
		log.Err(fmt.Errorf("error on shutdown: %v", err))
	}
}

func (a *Api) Shutdown() error {
	return a.Echo.Shutdown(context.TODO())
}
