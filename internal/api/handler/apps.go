package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/spacemeshos/explorer-backend/model"
)

func Apps(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	apps, total, err := cc.Service.GetApps(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get apps info: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       apps,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func App(c echo.Context) error {
	cc := c.(*ApiContext)
	app, err := cc.Service.GetApp(context.TODO(), c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get app `%s` info: %w", c.Param("id"), err)
	}

	return c.JSON(http.StatusOK, DataResponse{
		Data: []*model.App{app},
	})
}
