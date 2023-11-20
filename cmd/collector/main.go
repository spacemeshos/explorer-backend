package main

import (
	"context"
	"fmt"
	"github.com/spacemeshos/address"
	"github.com/spacemeshos/explorer-backend/collector"
	"github.com/spacemeshos/explorer-backend/collector/sql"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/urfave/cli/v2"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	version string
	commit  string
	branch  string
)

var (
	nodePublicAddressStringFlag  string
	nodePrivateAddressStringFlag string
	mongoDbUrlStringFlag         string
	mongoDbNameStringFlag        string
	testnetBoolFlag              bool
	syncFromLayerFlag            int
	syncMissingLayersBoolFlag    bool
	sqlitePathStringFlag         string
)

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:        "node-public",
		Usage:       "Spacemesh public node API address string in format <host>:<port>",
		Required:    false,
		Destination: &nodePublicAddressStringFlag,
		Value:       "localhost:9092",
		EnvVars:     []string{"SPACEMESH_NODE_PUBLIC"},
	},
	&cli.StringFlag{
		Name:        "node-private",
		Usage:       "Spacemesh private node API address string in format <host>:<port>",
		Required:    false,
		Destination: &nodePrivateAddressStringFlag,
		Value:       "localhost:9093",
		EnvVars:     []string{"SPACEMESH_NODE_PRIVATE"},
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
	&cli.IntFlag{
		Name:        "syncFromLayer",
		Usage:       ``,
		Required:    false,
		Value:       0,
		Destination: &syncFromLayerFlag,
		EnvVars:     []string{"SPACEMESH_SYNC_FROM_LAYER"},
	},
	&cli.BoolFlag{
		Name:        "syncMissingLayers",
		Usage:       `Use this flag to disable missing layers sync`,
		Required:    false,
		Destination: &syncMissingLayersBoolFlag,
		Value:       true,
		EnvVars:     []string{"SPACEMESH_SYNC_MISSING_LAYERS"},
	},
	&cli.StringFlag{
		Name:        "sqlite",
		Usage:       "Path to node sqlite file",
		Required:    false,
		Destination: &sqlitePathStringFlag,
		Value:       "explorer.sql",
		EnvVars:     []string{"SPACEMESH_SQLITE"},
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

		db, err := sql.Setup(sqlitePathStringFlag)
		if err != nil {
			log.Info("SQLite storage open error %v", err)
			return err
		}
		dbClient := &sql.Client{}

		c := collector.NewCollector(nodePublicAddressStringFlag, nodePrivateAddressStringFlag,
			syncMissingLayersBoolFlag, syncFromLayerFlag, mongoStorage, db, dbClient)
		mongoStorage.AccountUpdater = c

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		pidFile, err = os.OpenFile("/var/run/explorer-collector", os.O_RDWR|os.O_CREATE, 0644)
		if err == nil {
			_, err := pidFile.Write([]byte("started"))
			if err != nil {
				return err
			}
			err = pidFile.Close()
			if err != nil {
				return err
			}
		}

		go func() {
			<-sigs
			os.Remove("/var/run/explorer-collector")
			os.Exit(0)
		}()

		go func() {
			for {
				if err := c.Run(); err != nil {
					fmt.Println(err)
					time.Sleep(5 * time.Second)
				}
			}
		}()

		select {}
	}

	if err := app.Run(os.Args); err != nil {
		log.Info("%+v", err)
		os.Exit(1)
	}

	os.Exit(0)
}
