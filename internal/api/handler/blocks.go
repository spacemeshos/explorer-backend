package handler

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/go-spacemesh/log"
	"net/http"

	"github.com/spacemeshos/explorer-backend/model"
)

func Block(c echo.Context) error {
	cc := c.(*ApiContext)
	block, err := cc.Service.GetBlock(context.TODO(), c.Param("id"))
	if err != nil {
		log.Error("failed to get block `%s` info: %s", block, err)
		return err
	}
	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Block{block}})
}
