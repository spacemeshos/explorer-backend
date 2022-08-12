package testserver

import "github.com/spacemeshos/explorer-backend/api"

// TestCollectorService wrapper over fake collector service.
type TestCollectorService struct {
	server *api.Server
	port   int
}
