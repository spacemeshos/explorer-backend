package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/urfave/cli"
    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/collector"
    "github.com/spacemeshos/explorer-backend/storage"
)

var (
    version string
    commit  string
    branch  string
)

var (
    nodeAddressStringFlag string
    mongoDbUrlStringFlag string
    mongoDbNameStringFlag string
)

var flags = []cli.Flag{
    cli.StringFlag{
        Name:        "node",
        Usage:       "Spacemesh node API address string in format <host>:<port>",
        Required:    false,
        Destination: &nodeAddressStringFlag,
        Value:       "localhost:9092",
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
    app.Name = "Spacemesh Explorer Collector"
    app.Version = fmt.Sprintf("%s, commit '%s', branch '%s'", version, commit, branch)
    app.Flags = flags
    app.Writer = os.Stderr

    app.Action = func(ctx *cli.Context) (error) {
        var pidFile *os.File

        log.InitSpacemeshLoggingSystem("", "spacemesh-explorer-collector.log")

        env, ok := os.LookupEnv("SPACEMESH_NODE")
        if ok {
            nodeAddressStringFlag = env
        }
        env, ok = os.LookupEnv("SPACEMESH_MONGO_URI")
        if ok {
            mongoDbUrlStringFlag = env
        }
        env, ok = os.LookupEnv("SPACEMESH_MONGO_DB")
        if ok {
            mongoDbNameStringFlag = env
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
            pidFile.Write([]byte("started"));
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
