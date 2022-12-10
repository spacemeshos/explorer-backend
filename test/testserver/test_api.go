package testserver

import (
	"bytes"
	"context"
	"fmt"
	"github.com/spacemeshos/explorer-backend/test/testutils"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"

	"github.com/spacemeshos/explorer-backend/api"
	"github.com/spacemeshos/explorer-backend/storage"
)

const (
	testAPIServiceDB = "explorer_test"
)

// TestAPIService wrapper over fake api service.
type TestAPIService struct {
	Storage *storage.Storage
	server  *api.Server
	port    int
}

// StartTestAPIService start test api service.
func StartTestAPIService(dbPort int) (*TestAPIService, error) {
	appPort, err := freeport.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %s", err)
	}
	println("starting test api service on port", appPort)
	server, err := api.New(context.TODO(), &api.Config{
		ListenOn: ":" + fmt.Sprint(appPort),
		DbName:   testAPIServiceDB,
		DbUrl:    fmt.Sprintf("mongodb://localhost:%d", dbPort),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create api server: %s", err)
	}
	go server.Run()
	return &TestAPIService{
		server: server,
		port:   appPort,
	}, nil
}

// Get allow to execute GET request to the fake server.
func (tx *TestAPIService) Get(t *testing.T, path string) *testutils.TestResponse {
	t.Helper()

	path = strings.TrimLeft(path, "/")
	url := fmt.Sprintf("http://localhost:%d/%s", tx.port, path)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer([]byte{}))
	require.NoError(t, err, "failed to construct new request for url %s: %s", url, err)
	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}
	res, err := client.Do(req)
	require.NoError(t, err, "failed to make request to %s: %s", url, err)
	t.Cleanup(func() {
		require.NoError(t, res.Body.Close())
	})
	return &testutils.TestResponse{Res: res}
}
