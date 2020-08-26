package main

import (
    "context"
    "fmt"
    "os"

    "github.com/urfave/cli"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/api"
)

var (
    version string
    commit  string
    branch  string
)

var (
    listenStringFlag      string
    mongoDbUrlStringFlag  string
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
        Destination: &mongoDbUrlStringFlag,
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

    app.Action = func(ctx *cli.Context) (error) {

        serverCfg := &api.Config{
            ListenOn:     listenStringFlag,
            DbUrl:        mongoDbUrlStringFlag,
            DbName:       mongoDbNameStringFlag,
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

        log.Info("Server is shutdown")
        return nil
    }

    if err := app.Run(os.Args); err != nil {
        log.Info("%v", err)
        os.Exit(1)
    }

    os.Exit(0)
}
