package rest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const apiPrefix = "" // will be replaced to v2 after some endpoints will be refactored.

func TestNetworkInfoHandler(t *testing.T) {
	type resp struct {
		Network struct {
			Netid              int  `json:"netid"`
			Genesis            int  `json:"genesis"`
			Layers             int  `json:"layers"`
			Maxtx              int  `json:"maxtx"`
			Duration           int  `json:"duration"`
			Lastlayer          int  `json:"lastlayer"`
			Lastlayerts        int  `json:"lastlayerts"`
			Lastapprovedlayer  int  `json:"lastapprovedlayer"`
			Lastconfirmedlayer int  `json:"lastconfirmedlayer"`
			Connectedpeers     int  `json:"connectedpeers"`
			Issynced           bool `json:"issynced"`
			Syncedlayer        int  `json:"syncedlayer"`
			Toplayer           int  `json:"toplayer"`
			Verifiedlayer      int  `json:"verifiedlayer"`
		} `json:"network"`
	}

	res := apiServer.Get(t, apiPrefix+"/network-info")
	var networkInfo resp
	res.RequireOK(t)
	res.RequireUnmarshal(t, &networkInfo)
	require.Equal(t, 123, networkInfo.Network.Netid)
	require.Equal(t, 2, networkInfo.Network.Genesis)
	require.Equal(t, 10, networkInfo.Network.Layers)
	require.Equal(t, 100, networkInfo.Network.Maxtx)
	require.Equal(t, 10, networkInfo.Network.Duration)
}

func TestSyncedHandler(t *testing.T) {
	res := apiServer.Get(t, apiPrefix+"/synced")
	res.RequireTooEarly(t)
	apiServer.Storage.OnNodeStatus(10, true, 11, 12, 13)
	res = apiServer.Get(t, apiPrefix+"/synced")
	res.RequireOK(t)
}
