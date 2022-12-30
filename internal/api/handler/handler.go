package handler

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/internal/service"
)

const (
	// list of items to search from GET request.
	txs      = "txs"
	atxs     = "atxs"
	blocks   = "blocks"
	layers   = "layers"
	rewards  = "rewards"
	smeshers = "smeshers"
)

var Upgrader = websocket.Upgrader{}

type ApiContext struct {
	echo.Context
	Service service.AppService
}

type DataResponse struct {
	Data interface{} `json:"data"`
}

type PaginatedDataResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

type RedirectResponse struct {
	Redirect string `json:"redirect"`
}
