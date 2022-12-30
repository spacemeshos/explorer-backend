package testserver

import (
	"bytes"
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/require"

	apiv2 "github.com/spacemeshos/explorer-backend/internal/api"
	service2 "github.com/spacemeshos/explorer-backend/internal/service"
	"github.com/spacemeshos/explorer-backend/internal/storage/storagereader"
	"github.com/spacemeshos/explorer-backend/storage"
	"github.com/spacemeshos/explorer-backend/test/testutils"
)

// TestAPIService wrapper over fake api service.
type TestAPIService struct {
	Storage *storage.Storage
	port    int
}

// StartTestAPIServiceV2 start test api service with refacored router.
func StartTestAPIServiceV2(db *storage.Storage, dbReader *storagereader.Reader) (*TestAPIService, error) {
	appPort, err := freeport.GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get free port: %s", err)
	}
	println("starting test api service on port", appPort)

	api := apiv2.Init(service2.NewService(dbReader, time.Second), []string{"*"}, false)
	go api.Run(fmt.Sprintf(":%d", appPort))
	return &TestAPIService{
		Storage: db,
		port:    appPort,
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

// GetReadWS allow to execute WS read-only request to the fake server.
func (tx *TestAPIService) GetReadWS(t *testing.T, path string) <-chan []byte {
	t.Helper()

	path = strings.TrimLeft(path, "/")
	origin := fmt.Sprintf("http://localhost:%d", tx.port)
	conn, err := websocket.Dial(fmt.Sprintf("ws://localhost:%d/%s", tx.port, path), "", origin)
	if err != nil {
		t.Fatal("can't connect to websocket:", err)
	}
	result := make(chan []byte, 10)
	go func() {
		defer conn.Close()
		for {
			data := make([]byte, 1024)
			readLen, err := conn.Read(data)
			if err != nil {
				log.Println(err)
				return
			}
			result <- data[:readLen]
		}
	}()
	return result
}
