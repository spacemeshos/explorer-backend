package router

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/handler"
)

func Router(e *echo.Echo) {
	e.GET("/layer/:id", handler.Layer)
	e.GET("/epoch/:id", handler.Epoch)
	e.GET("/epoch/:id/decentral", handler.EpochDecentral)
	e.GET("/account/:address", handler.Account)
	e.GET("/smeshers/:epoch", handler.SmeshersByEpoch)
	e.GET("/smeshers", handler.Smeshers)
	e.GET("/smesher/:smesherId", handler.Smesher)
	e.GET("/overview", handler.Overview)
	e.GET("/circulation", handler.Circulation)
}

func RefreshRouter(e *echo.Echo) {
	g := e.Group("/refresh")
	g.GET("/epoch/:id", handler.EpochRefresh)
	e.GET("/epoch/:id/decentral", handler.EpochDecentralRefresh)
	g.GET("/overview", handler.OverviewRefresh)
	g.GET("/smeshers/:epoch", handler.SmeshersByEpochRefresh)
	g.GET("/smeshers", handler.SmeshersRefresh)
}
