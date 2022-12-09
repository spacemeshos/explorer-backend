package testutils

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResponse wrapper for http.Response.
type TestResponse struct {
	Res *http.Response
}

// RequireUnmarshal try unmarshal response body to given struct.
func (r *TestResponse) RequireUnmarshal(t *testing.T, dst interface{}) {
	t.Helper()
	err := json.NewDecoder(r.Res.Body).Decode(dst)
	require.NoError(t, err)
}

// RequireTooEarly check that response code is 429 Too Early.
func (r *TestResponse) RequireTooEarly(t *testing.T) {
	t.Helper()
	r.requireStatus(t, http.StatusTooEarly)
}

// RequireOK check that response code is 200 OK.
func (r *TestResponse) RequireOK(t *testing.T) {
	t.Helper()
	r.requireStatus(t, http.StatusOK)
}

func (r *TestResponse) requireStatus(t *testing.T, status int) {
	t.Helper()
	require.NotNil(t, r.Res, "response is nil")
	require.Equal(t, status, r.Res.StatusCode, "invalid response status code")
}
