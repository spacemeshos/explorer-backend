package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/urfave/cli"

	"github.com/spacemeshos/explorer-backend/api"
	"github.com/spacemeshos/explorer-backend/internal/router"
	appService "github.com/spacemeshos/explorer-backend/internal/service"
	"github.com/spacemeshos/explorer-backend/internal/storage/storagereader"
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
)

var flags = []cli.Flag{
	cli.StringFlag{
		Name:        "listen",
		Usage:       "Explorer API listen string in format <host>:<port>",
		Required:    false,
		Destination: &listenStringFlag,
		Value:       ":5000",
	},
	cli.StringFlag{
		Name:        "mongodb",
		Usage:       "Explorer MongoDB Uri string in format mongodb://<host>:<port>",
		Required:    false,
		Destination: &mongoDbURLStringFlag,
		Value:       "mongodb://localhost:27017",
	},
	cli.StringFlag{
		Name:        "db",
		Usage:       "MongoDB Explorer database name string",
		Required:    false,
		Destination: &mongoDbNameStringFlag,
		Value:       "explorer",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "Spacemesh Explorer REST API Server"
	app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
	app.Flags = flags
	app.Writer = os.Stderr

	env, ok := os.LookupEnv("SPACEMESH_API_LISTEN")
	if ok {
		listenStringFlag = env
	}
	env, ok = os.LookupEnv("SPACEMESH_MONGO_URI")
	if ok {
		mongoDbURLStringFlag = env
	}
	env, ok = os.LookupEnv("SPACEMESH_MONGO_DB")
	if ok {
		mongoDbNameStringFlag = env
	}

	flag := true // flag switch old|new router.

	app.Action = func(ctx *cli.Context) error {
		if flag {
			if err := newApp(mongoDbURLStringFlag, mongoDbNameStringFlag, listenStringFlag); err != nil {
				log.Error("error start service", err.Error())
				return err
			}
		} else {
			if err := oldApp(mongoDbURLStringFlag, mongoDbNameStringFlag, listenStringFlag); err != nil {
				log.Error("error start service", err.Error())
				return err
			}
		}

		log.Info("Server is shutdown")
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Info("%v", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func oldApp(mongoDbURL, mongoDbName, listen string) error {
	serverCfg := &api.Config{
		ListenOn: listen,
		DbUrl:    mongoDbURL,
		DbName:   mongoDbName,
	}

	server, err := api.New(context.Background(), serverCfg)
	if err != nil {
		log.Info("%+v", err)
		return err
	}

	if err = server.Run(); err != nil {
		log.Info("%+v", err)
		return err
	}
	return nil
}

func newApp(mongoDbURL, mongoDbName, listen string) error {
	dbReader, err := storagereader.NewStorageReader(context.Background(), mongoDbURL, mongoDbName)
	if err != nil {
		return fmt.Errorf("error init storage reader: %w", err)
	}
	service := appService.NewService(dbReader, time.Minute)
	app := router.InitAppRouter(&router.Config{
		ListenOn: listen,
	}, service)
	log.Info(fmt.Sprintf("starting server on %s", listen))
	if err = app.Run(); err != nil {
		return fmt.Errorf("error start service: %w", err)
	}
	return nil
}
