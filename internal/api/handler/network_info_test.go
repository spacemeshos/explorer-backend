package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const apiPrefix = "" // will be replaced to v2 after some endpoints will be refactored.
type networkResp struct {
	Network struct {
		Genesisid          string `json:"genesisid"`
		Genesis            uint64 `json:"genesis"`
		Layers             uint32 `json:"layers"`
		Maxtx              uint64 `json:"maxtx"`
		Duration           uint64 `json:"duration"`
		Lastlayer          int    `json:"lastlayer"`
		Lastlayerts        int    `json:"lastlayerts"`
		Lastapprovedlayer  int    `json:"lastapprovedlayer"`
		Lastconfirmedlayer int    `json:"lastconfirmedlayer"`
		Connectedpeers     int    `json:"connectedpeers"`
		Issynced           bool   `json:"issynced"`
		Syncedlayer        int    `json:"syncedlayer"`
		Toplayer           int    `json:"toplayer"`
		Verifiedlayer      int    `json:"verifiedlayer"`
	} `json:"network"`
}

func TestNetworkInfoHandler(t *testing.T) {
	res := apiServer.Get(t, apiPrefix+"/network-info")
	var networkInfo networkResp
	res.RequireOK(t)
	res.RequireUnmarshal(t, &networkInfo)
	compareNetworkInfo(t, networkInfo)
}

func TestSyncedHandler(t *testing.T) {
	res := apiServer.Get(t, apiPrefix+"/synced")
	res.RequireTooEarly(t)
	apiServer.Storage.OnNodeStatus(10, true, 11, 12, 13)
	require.Eventually(t, func() bool {
		res = apiServer.Get(t, apiPrefix+"/synced")
		return res.Res.StatusCode == http.StatusOK
	}, 4*time.Second, 1*time.Second)
}

func TestWSNetworkInfoHandler(t *testing.T) {
	t.Parallel()
	res := apiServer.GetReadWS(t, apiPrefix+"/ws/network-info")
	counter := 0
loop:
	for {
		select {
		case c := <-res:
			var networkInfo networkResp
			require.NoError(t, json.Unmarshal(c, &networkInfo))
			compareNetworkInfo(t, networkInfo)
			counter++

		case <-time.After(2 * time.Second):
			break loop
		}
	}
	require.NotZero(t, counter)
}

func compareNetworkInfo(t *testing.T, networkInfo networkResp) {
	require.Equal(t, string(seed.GenesisID), networkInfo.Network.Genesisid)
	require.Equal(t, seed.GenesisTime, networkInfo.Network.Genesis)
	require.Equal(t, seed.EpochNumLayers, networkInfo.Network.Layers)
	require.Equal(t, seed.MaxTransactionPerSecond, networkInfo.Network.Maxtx)
	require.Equal(t, seed.LayersDuration, networkInfo.Network.Duration)
}
