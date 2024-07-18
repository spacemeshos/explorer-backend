package main

import (
	"fmt"
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/explorer-backend/api"
	"github.com/spacemeshos/explorer-backend/api/cache"
	"github.com/spacemeshos/explorer-backend/api/storage"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	version string
	commit  string
	branch  string
)

var (
	listenStringFlag     string
	testnetBoolFlag      bool
	allowedOrigins       = cli.NewStringSlice("*")
	debug                bool
	sqlitePathStringFlag string
	layersPerEpoch       int64
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

		db, err := storage.Setup(sqlitePathStringFlag)
		if err != nil {
			log.Info("SQLite storage open error %v", err)
			return err
		}
		dbClient := &storage.Client{}

		server := api.Init(db, dbClient, allowedOrigins.Value(), debug, layersPerEpoch, c)

		log.Info(fmt.Sprintf("starting server on %s", listenStringFlag))
		server.Run(listenStringFlag)

		log.Info("server is shutdown")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Info("%v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
