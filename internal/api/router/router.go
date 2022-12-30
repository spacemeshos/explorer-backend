package router

import (
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/internal/api/handler"
)

func Init(e *echo.Echo) {
	e.GET("/healthz", handler.HealthzHandler)
	e.GET("/synced", handler.Synced)

	e.GET("/network-info", handler.NetworkInfo)
	e.GET("/ws/network-info", handler.NetworkInfoWS)

	e.GET("/epochs", handler.Epochs)
	e.GET("/epochs/:id", handler.Epoch)
	e.GET("/epochs/:id/:entity", handler.EpochDetails)

	e.GET("/layers", handler.Layers)
	e.GET("/layers/:id", handler.Layer)
	e.GET("/layers/:id/:entity", handler.LayerDetails)

	e.GET("/smeshers", handler.Smeshers)
	e.GET("/smeshers/:id", handler.Smesher)
	e.GET("/smeshers/:id/:entity", handler.SmesherDetails)

	e.GET("/apps", handler.Apps)
	e.GET("/apps/:id", handler.App)

	e.GET("/atxs", handler.Activations)
	e.GET("/atxs/:id", handler.Activation)

	e.GET("/txs", handler.Transactions)
	e.GET("/txs/:id", handler.Transaction)

	e.GET("/rewards", handler.Rewards)
	e.GET("/rewards/:id", handler.Reward)

	e.GET("/accounts", handler.Accounts)
	e.GET("/accounts/:id", handler.Account)
	e.GET("/accounts/:id/:entity", handler.AccountDetails)

	e.GET("/blocks/:id", handler.Block)

	e.GET("/search/:id", handler.Search)
}
