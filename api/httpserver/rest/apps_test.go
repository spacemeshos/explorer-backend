package rest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApps(t *testing.T) { // "/apps"
	t.Parallel()
	res := apiServer.Get(t, apiPrefix+"/apps?pagesize=1000")
	res.RequireOK(t)
	var resp appResp
	res.RequireUnmarshal(t, &resp)
	require.Equal(t, len(generator.Apps), len(resp.Data))
	for _, app := range resp.Data {
		generatedAp, ok := generator.Apps[app.Address]
		require.True(t, ok)
		require.Equal(t, generatedAp, app)
	}
}

func TestApp(t *testing.T) { // "/apps/{id}"
	t.Parallel()
	for _, app := range generator.Apps {
		res := apiServer.Get(t, apiPrefix+"/apps/"+app.Address)
		res.RequireOK(t)
	}
}
