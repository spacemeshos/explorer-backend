package router

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/api/handler"
)

func Init(e *echo.Echo) {
	e.GET("/stats/layer/:id", handler.LayerStats)
	e.GET("/stats/epoch/:id", handler.EpochStats)
	e.GET("/smeshers/:epoch", handler.SmeshersByEpoch)
	e.GET("/smeshers", handler.Smeshers)
	e.GET("/smesher/:smesherId", handler.Smesher)
	e.GET("/overview", handler.Overview)
}
