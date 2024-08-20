package main

import (
	"errors"
	"fmt"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/explorer-backend/api"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/router"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/timesync"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	version string
	commit  string
	branch  string
)

var (
	listenStringFlag        string
	refreshListenStringFlag string
	testnetBoolFlag         bool
	allowedOrigins          = cli.NewStringSlice("*")
	debug                   bool
	sqlitePathStringFlag    string
	layersPerEpoch          int64
	genesisTimeStringFlag   string
	layerDuration           time.Duration
	labelsPerUnit           uint64
	metricsPortFlag         string
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:        "listen",
		Usage:       "Explorer API listen string in format <host>:<port>",
		Required:    false,
		Destination: &listenStringFlag,
		Value:       ":5000",
		EnvVars:     []string{"SPACEMESH_API_LISTEN"},
	},
	&cli.StringFlag{
		Name:        "listen-refresh",
		Usage:       "Explorer refresh API listen string in format <host>:<port>",
		Required:    false,
		Destination: &refreshListenStringFlag,
		Value:       ":5050",
		EnvVars:     []string{"SPACEMESH_REFRESH_API_LISTEN"},
	},
	&cli.BoolFlag{
		Name:        "testnet",
		Usage:       `Use this flag to enable testnet preset ("stest" instead of "sm" for wallet addresses)`,
		Required:    false,
		Destination: &testnetBoolFlag,
		EnvVars:     []string{"SPACEMESH_TESTNET"},
	},
	&cli.StringSliceFlag{
		Name:        "allowed-origins",
		Usage:       `Use this flag to set allowed origins for CORS (default: "*")`,
		Destination: allowedOrigins,
		EnvVars:     []string{"ALLOWED_ORIGINS"},
	},
	&cli.BoolFlag{
		Name:        "debug",
		Usage:       "Use this flag to enable echo debug option along with logger middleware",
		Required:    false,
		Destination: &debug,
		EnvVars:     []string{"DEBUG"},
	},
	&cli.StringFlag{
		Name:        "sqlite",
		Usage:       "Path to node sqlite file",
		Required:    false,
		Destination: &sqlitePathStringFlag,
		Value:       "explorer.sql",
		EnvVars:     []string{"SPACEMESH_SQLITE"},
	},
	&cli.Int64Flag{
		Name:        "layers-per-epoch",
		Usage:       "Number of layers per epoch",
		Required:    false,
		Destination: &layersPerEpoch,
		Value:       4032,
		EnvVars:     []string{"SPACEMESH_LAYERS_PER_EPOCH"},
	},
	&cli.StringFlag{
		Name:        "genesis-time",
		Usage:       "Genesis time in RFC3339 format",
		Required:    true,
		Destination: &genesisTimeStringFlag,
		Value:       "2024-06-21T13:00:00.000Z",
		EnvVars:     []string{"SPACEMESH_GENESIS_TIME"},
	},
	&cli.DurationFlag{
		Name:        "layer-duration",
		Usage:       "Duration of a single layer",
		Required:    false,
		Destination: &layerDuration,
		Value:       30 * time.Second,
		EnvVars:     []string{"SPACEMESH_LAYER_DURATION"},
	},
	&cli.Uint64Flag{
		Name:        "labels-per-unit",
		Usage:       "Number of labels per unit",
		Required:    false,
		Destination: &labelsPerUnit,
		Value:       1024,
		EnvVars:     []string{"SPACEMESH_LABELS_PER_UNIT"},
	},
	&cli.StringFlag{
		Name:        "metricsPort",
		Usage:       ``,
		Required:    false,
		Value:       ":5070",
		Destination: &metricsPortFlag,
		EnvVars:     []string{"SPACEMESH_METRICS_PORT"},
	},
	&cli.DurationFlag{
		Name:        "cache-ttl",
		Usage:       "Cache TTL for resources like overview, epochs, cumulative stats etc.",
		Required:    false,
		Value:       0,
		Destination: &cache.Expiration,
		EnvVars:     []string{"SPACEMESH_CACHE_TTL"},
	},
	&cli.DurationFlag{
		Name:        "short-cache-ttl",
		Usage:       "Short Cache TTL for resources like layers, accounts etc.",
		Required:    false,
		Value:       5 * time.Minute,
		Destination: &cache.ShortExpiration,
		EnvVars:     []string{"SPACEMESH_SHORT_CACHE_TTL"},
	},
	&cli.StringFlag{
		Name:        "redis",
		Usage:       "Redis address for cache / if not set memory cache will be used",
		Required:    false,
		Value:       "",
		Destination: &cache.RedisAddress,
		EnvVars:     []string{"SPACEMESH_REDIS"},
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "Spacemesh Explorer REST API Server"
	app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
	app.Flags = flags
	app.Writer = os.Stderr

	app.Action = func(ctx *cli.Context) error {
		if testnetBoolFlag {
			address.SetAddressConfig("stest")
			types.SetNetworkHRP("stest")
			log.Info(`network HRP set to "stest"`)
		}
		log.Info("layers per epoch: %d", layersPerEpoch)
		log.Info("debug: %v", debug)
		log.Info("sqlite path: %s", sqlitePathStringFlag)

		c := cache.New()

		gTime, err := time.Parse(time.RFC3339, genesisTimeStringFlag)
		if err != nil {
			return fmt.Errorf("cannot parse genesis time %s: %w", genesisTimeStringFlag, err)
		}

		clock, err := timesync.NewClock(
			timesync.WithLayerDuration(layerDuration),
			timesync.WithTickInterval(1*time.Second),
			timesync.WithGenesisTime(gTime),
			timesync.WithLogger(zap.NewNop()),
		)
		if err != nil {
			return fmt.Errorf("cannot create clock: %w", err)
		}

		db, err := storage.Setup(sqlitePathStringFlag)
		if err != nil {
			log.Info("SQLite storage open error %v", err)
			return err
		}
		dbClient := &storage.Client{
			NodeClock:     clock,
			Testnet:       testnetBoolFlag,
			LabelsPerUnit: labelsPerUnit,
			BitsPerLabel:  128,
		}

		var wg sync.WaitGroup
		wg.Add(3)
		// start api server
		server := api.Init(db,
			dbClient,
			allowedOrigins.Value(),
			debug,
			layersPerEpoch,
			c,
			router.Router)
		go func() {
			defer wg.Done()
			log.Info(fmt.Sprintf("starting api server on %s", listenStringFlag))
			server.Run(listenStringFlag)
		}()

		// start refresh api server
		refreshServer := api.Init(db,
			dbClient,
			allowedOrigins.Value(),
			debug,
			layersPerEpoch,
			c,
			router.RefreshRouter)
		go func() {
			defer wg.Done()
			log.Info(fmt.Sprintf("starting refresh api server on %s", refreshListenStringFlag))
			refreshServer.Run(refreshListenStringFlag)
		}()

		go func() {
			defer wg.Done()
			metrics := echo.New()
			metrics.GET("/metrics", echoprometheus.NewHandler())
			if err := metrics.Start(metricsPortFlag); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal("%v", err)
			}
		}()

		wg.Wait()

		log.Info("server is shutdown")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Info("%v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
