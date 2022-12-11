package main

import (
	"context"
	"fmt"
	"github.com/spacemeshos/address"
	"os"
	"os/signal"
	"syscall"

	"github.com/spacemeshos/explorer-backend/collector"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/urfave/cli/v2"
)

var (
	version string
	commit  string
	branch  string
)

var (
	nodeAddressStringFlag string
	mongoDbUrlStringFlag  string
	mongoDbNameStringFlag string
	testnetBoolFlag       bool
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:        "node",
		Usage:       "Spacemesh node API address string in format <host>:<port>",
		Required:    false,
		Destination: &nodeAddressStringFlag,
		Value:       "localhost:9092",
		EnvVars:     []string{"SPACEMESH_NODE"},
	},
	&cli.StringFlag{
		Name:        "mongodb",
		Usage:       "Explorer MongoDB Uri string in format mongodb://<host>:<port>",
		Required:    false,
		Destination: &mongoDbUrlStringFlag,
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
	app.Name = "Spacemesh Explorer Collector"
	app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
	app.Flags = flags
	app.Writer = os.Stderr

	app.Action = func(ctx *cli.Context) error {
		var pidFile *os.File

		if testnetBoolFlag {
			address.SetAddressConfig("stest")
			log.Info(`Network HRP set to "stest"`)
		}

		mongoStorage, err := storage.New(context.Background(), mongoDbUrlStringFlag, mongoDbNameStringFlag)
		if err != nil {
			log.Info("MongoDB storage open error %v", err)
			return err
		}

		c := collector.NewCollector(nodeAddressStringFlag, mongoStorage)
		mongoStorage.AccountUpdater = c

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		pidFile, err = os.OpenFile("/var/run/explorer-collector", os.O_RDWR|os.O_CREATE, 0644)
		if err == nil {
			pidFile.Write([]byte("started"))
			pidFile.Close()
		}

		go func() {
			_ = <-sigs
			os.Remove("/var/run/explorer-collector")
			os.Exit(0)
		}()

		c.Run()

		os.Remove("/var/run/explorer-collector")
		log.Info("Collector is shutdown")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Info("%+v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
