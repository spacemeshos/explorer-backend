package api

import (
    "context"
    "errors"
    "os"
    "os/signal"
    "syscall"

    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/explorer-backend/api/httpserver"
    "github.com/spacemeshos/explorer-backend/storage"
)

type Server struct {
    ctx    context.Context
    cancel context.CancelFunc
    cfg    *Config

    httpServer      *httpserver.Server
    storage         *storage.Storage
}

type Config struct {
    ListenOn       string
    DbUrl          string
    DbName         string

    httpServerCfg  httpserver.Config
}

func New(ctx context.Context, cfg *Config) (*Server, error) {
    var err error

    log.Info("Create new server")

    if cfg == nil {
        return nil, errors.New("Empty server config")
    }

    server := &Server{
        cfg: cfg,
    }

    if ctx == nil {
        server.ctx, server.cancel = context.WithCancel(context.Background())
    } else {
        server.ctx, server.cancel = context.WithCancel(ctx)
    }

    server.cfg.httpServerCfg.ListenOn = server.cfg.ListenOn

    if server.storage, err = storage.New(server.ctx, server.cfg.DbUrl, server.cfg.DbName); err != nil {
        return nil, err
    }

    if server.httpServer, err = httpserver.New(server.ctx, &server.cfg.httpServerCfg, server.storage); err != nil {
        return nil, err
    }

    log.Info("New API server is created")

    return server, nil
}

func (server *Server) Run() error {
    log.Info("Starting server")

    log.Info("Server is running. For exit <CTRL-c>")
    go func() { server.httpServer.Run() } ()

    syscalCh := make(chan os.Signal, 1)
    signal.Notify(syscalCh, syscall.SIGINT, syscall.SIGTERM)

    select {
    case s := <-syscalCh:
        log.Info("Exiting, got signal %v", s)
        server.Shutdown()
        return nil
    }
}

func (s *Server) Shutdown() {
    log.Info("Shutting down server...")

    defer s.cancel()

    if err := s.httpServer.Shutdown(); err != nil {
        log.Info("API Server shutdown error %v", err)
    }

    log.Info("Server is shutdown")
}
