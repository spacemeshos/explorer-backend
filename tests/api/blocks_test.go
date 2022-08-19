package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBlocks(t *testing.T) { // /blocks/{id}
	t.Parallel()
	for _, epoch := range generator.Epochs {
		for _, block := range epoch.Blocks {
			res := apiServer.Get(t, apiPrefix+"/blocks/"+block.Id)
			res.RequireOK(t)
			var resp blockResp
			res.RequireUnmarshal(t, &resp)
			require.Equal(t, 1, len(resp.Data))
			require.Equal(t, block, &resp.Data[0])
		}
	}
}
