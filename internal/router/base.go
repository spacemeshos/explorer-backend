package router

import (
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/spacemeshos/go-spacemesh/log"

	logger "github.com/spacemeshos/explorer-backend/internal/router/middleware"
	"github.com/spacemeshos/explorer-backend/internal/service"
)

const (
	// list of items to search from GET request.
	txs      = "txs"
	atxs     = "atxs"
	blocks   = "blocks"
	layers   = "layers"
	rewards  = "rewards"
	smeshers = "smeshers"
)

// Config is the configuration of the server.
type Config struct {
	ListenOn string
}

// AppRouter is the main router for the app.
type AppRouter struct {
	routerTimeout time.Duration
	conf          *Config
	appService    service.AppService
	fiber.Router
	FiberApp *fiber.App
}

// InitAppRouter initializes the app router.
func InitAppRouter(conf *Config, appService service.AppService) *AppRouter {
	fiberApp := fiber.New(
		fiber.Config{
			ErrorHandler: func(ctx *fiber.Ctx, err error) error {
				log.Error("error get info. path: `%s`, err: %s", ctx.Request().URI().PathOriginal(), err)
				if errors.Is(err, service.ErrNotFound) {
					return ctx.Status(http.StatusNotFound).JSON(fiber.Map{"error": "not found"})
				} else if errors.Is(err, fiber.ErrBadRequest) {
					return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "bad request"})
				}
				return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
			},
			DisableStartupMessage: true,
			ReadBufferSize:        4096 * 16,
			WriteBufferSize:       4096 * 16,
			BodyLimit:             100 * 1024 * 1024,
		},
	)

	fiberApp.Use(recover.New())
	fiberApp.Use(logger.New())

	app := &AppRouter{
		routerTimeout: 5 * time.Second,
		appService:    appService,
		conf:          conf,
		FiberApp:      fiberApp,
		Router:        fiberApp,
	}
	app.initRoutes()
	return app
}

func (a *AppRouter) initRoutes() {
	a.FiberApp.Get("/healthz", a.healthzHandler)
	a.FiberApp.Get("/synced", a.synced)

	a.FiberApp.Get("/network-info", a.networkInfo)
	a.FiberApp.Get("/ws/network-info", websocket.New(a.networkInfoWS))

	a.FiberApp.Get("/epochs", a.epochs)
	a.FiberApp.Get("/epochs/:id", a.epoch)
	a.FiberApp.Get("/epochs/:id/:entity", a.epochDetails) // layers, txs, smeshers, rewards, atxs.

	a.FiberApp.Get("/layers", a.layers)
	a.FiberApp.Get("/layers/:id", a.layer)
	a.FiberApp.Get("/layers/:id/:entity", a.layerDetails) // txs, smeshers, blocks, rewards, atxs

	a.FiberApp.Get("/smeshers", a.smeshers)
	a.FiberApp.Get("/smeshers/:id", a.smesher)
	a.FiberApp.Get("/smeshers/:id/:entity", a.smesherDetails) // atx, rewards

	a.FiberApp.Get("/apps", a.apps)
	a.FiberApp.Get("/apps/:id", a.app)

	a.FiberApp.Get("/atxs", a.activations)
	a.FiberApp.Get("/atxs/:id", a.activation)

	a.FiberApp.Get("/txs", a.transactions)
	a.FiberApp.Get("/txs/:id", a.transaction)

	a.FiberApp.Get("/rewards", a.rewards)
	a.FiberApp.Get("/rewards/:id", a.reward)

	a.FiberApp.Get("/accounts", a.accounts)
	a.FiberApp.Get("/accounts/:id", a.account)
	a.FiberApp.Get("/accounts/:id/:entity", a.accountDetails) // txs, rewards

	a.FiberApp.Get("/blocks/:id", a.block)

	a.FiberApp.Get("/search/:id", a.search)
}

func (a *AppRouter) getPagination(ctx *fiber.Ctx) (pageNumber, pageSize int64) {
	pageNumber = 1
	pageSize = 20
	if page := ctx.Query("page"); page != "" {
		pageNumber, _ = strconv.ParseInt(page, 10, 32)
		if pageNumber <= 0 {
			pageNumber = 1
		}
	}
	if size := ctx.Query("pagesize"); size != "" {
		pageSize, _ = strconv.ParseInt(size, 10, 32)
		if pageSize <= 0 {
			pageSize = 20
		}
	}
	return pageNumber, pageSize
}

type pagination struct {
	TotalCount  int64 `json:"totalCount"`
	PageCount   int64 `json:"pageCount"`
	PerPage     int64 `json:"perPage"`
	Next        int64 `json:"next"`
	HasNext     bool  `json:"hasNext"`
	HasPrevious bool  `json:"hasPrevious"`
	Current     int64 `json:"current"`
	Previous    int64 `json:"previous"`
}

func (a *AppRouter) setPaginationResponse(total int64, pageNumber int64, pageSize int64) pagination {
	pageCount := (total + pageSize - 1) / pageSize
	result := pagination{
		TotalCount: total,
		PageCount:  pageNumber,
		PerPage:    pageSize,
		Next:       pageCount,
		Current:    pageNumber,
		Previous:   1,
	}
	if pageNumber < pageCount {
		result.Next = pageNumber + 1
		result.HasNext = true
	}
	if pageNumber > 1 {
		result.Previous = pageNumber - 1
		result.HasPrevious = true
	}
	return result
}

// Run starts the server.
func (a *AppRouter) Run() error {
	log.Info("Server is running. For exit <CTRL-c>")
	if err := a.FiberApp.Listen(a.conf.ListenOn); err != nil {
		log.Error("server stopped: %s", err)
	}

	syscalCh := make(chan os.Signal, 1)
	signal.Notify(syscalCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case s := <-syscalCh:
		log.Info("Exiting, got signal %v", s)
		if err := a.Shutdown(); err != nil {
			log.Error("Error on shutdown: %v", err)
		}
		return nil
	}
}

// Shutdown gracefully shuts down the server.
func (a *AppRouter) Shutdown() error {
	return a.FiberApp.Shutdown()
}
