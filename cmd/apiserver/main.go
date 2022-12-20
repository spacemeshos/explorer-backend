package main

import (
	"context"
	"fmt"
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/explorer-backend/internal/api"
	appService "github.com/spacemeshos/explorer-backend/internal/service"
	"github.com/spacemeshos/explorer-backend/internal/storage/storagereader"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var (
	version string
	commit  string
	branch  string
)

var (
	listenStringFlag      string
	mongoDbURLStringFlag  string
	mongoDbNameStringFlag string
	testnetBoolFlag       bool
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
		Name:        "mongodb",
		Usage:       "Explorer MongoDB Uri string in format mongodb://<host>:<port>",
		Required:    false,
		Destination: &mongoDbURLStringFlag,
		Value:       "mongodb://localhost:27017",
		EnvVars:     []string{"SPACEMESH_MONGO_URI"},
	},
	&cli.StringFlag{
		Name:        "db",
		Usage:       "MongoDB Explorer database name string",
		Required:    false,
		Destination: &mongoDbNameStringFlag,
		Value:       "explorer",
		EnvVars:     []string{"SPACEMESH_MONGO_DB"},
	},
	&cli.BoolFlag{
		Name:        "testnet",
		Usage:       `Use this flag to enable testnet preset ("stest" instead of "sm" for wallet addresses)`,
		Required:    false,
		Destination: &testnetBoolFlag,
		EnvVars:     []string{"SPACEMESH_TESTNET"},
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
			log.Info(`network HRP set to "stest"`)
		}

		dbReader, err := storagereader.NewStorageReader(context.Background(), mongoDbURLStringFlag, mongoDbNameStringFlag)
		if err != nil {
			return fmt.Errorf("error init storage reader: %w", err)
		}

		service := appService.NewService(dbReader, time.Minute)
		server := api.Init(service)

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
