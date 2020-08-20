package httpserver

import (
    "context"
    "errors"
    "fmt"
    "net"
    "net/http"
    "time"

    "github.com/gorilla/mux"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/explorer-backend/api/httpserver/rest"
    "github.com/spacemeshos/explorer-backend/storage"
)

type Server struct {
    ctx       context.Context
    cancel    context.CancelFunc
    config    *Config

    listener    net.Listener
    router      *mux.Router
    httpServer  *http.Server
    restService *rest.Service
}

type Config struct {
    ListenOn        string // listener address string
    ReadTimeout     int    // read timeout - default 60 sec
    WriteTimeout    int    // write timeout - default 60 sec
    IdleTimeout     int    // idle timeout - default 60 sec
    MaxHeaderBytes  int    // max HTTP header length - default 1 MB
    MaxBodyBytes    int    // max HTTP body length - default 0 - unlimited
    ShutdownTimeout int
}

func New(ctx context.Context, cfg *Config, storage *storage.Storage) (*Server, error) {
    var err error

    log.Info("Creating new HTTP server")

    if cfg == nil {
        return nil, errors.New("Empty HTTP server config")
    }
    if cfg.ListenOn == "" {
        return nil, errors.New("Empty HTTP listener address")
    }

    server := &Server{
        config:    cfg,
    }

    if ctx == nil {
        server.ctx, server.cancel = context.WithCancel(context.Background())
    } else {
        server.ctx, server.cancel = context.WithCancel(ctx)
    }

    server.httpServer = &http.Server{
        ReadTimeout:  time.Duration(cfg.ReadTimeout * int(time.Second)),
        WriteTimeout: time.Duration(cfg.WriteTimeout * int(time.Second)),
        IdleTimeout:  time.Duration(cfg.IdleTimeout * int(time.Second)),
    }

    if cfg.MaxHeaderBytes > 0 {
        server.httpServer.MaxHeaderBytes = cfg.MaxHeaderBytes
    }

    server.listener, err = net.Listen("tcp", cfg.ListenOn)
    if err != nil {
        return nil, fmt.Errorf("Failed to create new TCP listener: network = 'tcp', address %v, error %v", cfg.ListenOn, err)
    }

    log.Info("Created new TCP listener: network = 'tcp', address", cfg.ListenOn)

    server.restService, err = rest.New(server.ctx, storage)
    if err != nil {
        return nil, fmt.Errorf("Failed to create new TCP listener: network = 'tcp', address %v, error %v", cfg.ListenOn, err)
    }

    server.router = mux.NewRouter()

    http.Handle("/", server.router)

    epochsRouter := server.router.PathPrefix("/epochs").Subrouter()
    epochsRouter.HandleFunc("/",                       server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}",            server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}/layers",     server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}/txs",        server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}/smeshers",   server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}/rewards",    server.restService.NotImplemented).Methods("GET")
    epochsRouter.HandleFunc("/{id:[0-9]+}/atxs",       server.restService.NotImplemented).Methods("GET")


    layersRouter := server.router.PathPrefix("/layers").Subrouter()
    layersRouter.HandleFunc("/",                       server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}",            server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}/txs",        server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}/smeshers",   server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}/blocks",     server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}/rewards",    server.restService.NotImplemented).Methods("GET")
    layersRouter.HandleFunc("/{id:[0-9]+}/atxs",       server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/smeshers",              server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/smeshers/{id}",         server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/smeshers/{id}/atxs",    server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/smeshers/{id}/rewards", server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/apps",                  server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/apps/{id}",             server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/txs",                   server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/txs/{id}",              server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/rewards",               server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/rewards/{id}",          server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/address/{id}",          server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/address/{id}/txs",      server.restService.NotImplemented).Methods("GET")
    server.router.HandleFunc("/address/{id}/rewards",  server.restService.NotImplemented).Methods("GET")

    server.router.HandleFunc("/blocks/{id}",           server.restService.NotImplemented).Methods("GET")

    log.Info("HTTP server is created")
    return server, nil
}

func (s *Server) Run() (error) {
    return s.httpServer.Serve(s.listener)
}

func (s *Server) Shutdown() (myerr error) {
    log.Info("Waiting for shutdown HTTP Server: sec", s.config.ShutdownTimeout)

    defer s.cancel()

    cancelCtx, cancel := context.WithTimeout(s.ctx, time.Duration(s.config.ShutdownTimeout*int(time.Second)))
    defer cancel()

    if err := s.httpServer.Shutdown(cancelCtx); err != nil {
        return fmt.Errorf("Failed to shutdown HTTP server: %v sec, %v", s.config.ShutdownTimeout, err)
    }

    if err := s.restService.Shutdown(); err != nil {
        return fmt.Errorf("Failed to shutdown REST service: %v sec, %v", s.config.ShutdownTimeout, err)
    }

    log.Info("HTTP Server shutdown successfuly")
    return nil
}
