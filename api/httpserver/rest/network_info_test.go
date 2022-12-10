package rest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const apiPrefix = "" // will be replaced to v2 after some endpoints will be refactored.

func TestNetworkInfoHandler(t *testing.T) {
	type resp struct {
		Network struct {
			Netid              uint64 `json:"netid"`
			Genesis            uint64 `json:"genesis"`
			Layers             uint64 `json:"layers"`
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

	res := apiServer.Get(t, apiPrefix+"/network-info")
	var networkInfo resp
	res.RequireOK(t)
	res.RequireUnmarshal(t, &networkInfo)
	require.Equal(t, seed.NetID, networkInfo.Network.Netid)
	require.Equal(t, seed.GenesisTime, networkInfo.Network.Genesis)
	require.Equal(t, seed.EpochNumLayers, networkInfo.Network.Layers)
	require.Equal(t, seed.MaxTransactionPerSecond, networkInfo.Network.Maxtx)
	require.Equal(t, seed.LayersDuration, networkInfo.Network.Duration)
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
