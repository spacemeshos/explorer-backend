package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/spacemeshos/explorer-backend/model"
)

func Activations(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	atxs, total, err := cc.Service.GetActivations(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get apps info: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       atxs,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Activation(c echo.Context) error {
	cc := c.(*ApiContext)
	atx, err := cc.Service.GetActivation(context.TODO(), c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get activation %s info: %w", c.Param("id"), err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Activation{atx}})
}
