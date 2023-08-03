package api

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spacemeshos/explorer-backend/internal/api/handler"
	"github.com/spacemeshos/explorer-backend/internal/api/router"
	"github.com/spacemeshos/explorer-backend/internal/service"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Api struct {
	Echo *echo.Echo
}

func Init(appService service.AppService, allowedOrigins []string, debug bool) *Api {
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &handler.ApiContext{
				Context: c,
				Service: appService,
			}
			return next(cc)
		}
	})
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
	}))
	handler.Upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("origin")
		for _, val := range allowedOrigins {
			if origin == val || val == "*" {
				return true
			}
		}
		return false
	}

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
