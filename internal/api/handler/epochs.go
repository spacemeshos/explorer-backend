package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func Epochs(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	epochs, total, err := cc.Service.GetEpochs(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get epoch list: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       epochs,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Epoch(c echo.Context) error {
	cc := c.(*ApiContext)
	layerNum, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}
	epochs, err := cc.Service.GetEpoch(context.TODO(), layerNum)
	if err != nil {
		return fmt.Errorf("failed to get epoch info: %w", err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Epoch{epochs}})
}

func EpochDetails(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	epochID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("wrong epoch id: %w", err)
	}
	var (
		response interface{}
		total    int64
	)

	switch c.Param("entity") {
	case layers:
		response, total, err = cc.Service.GetEpochLayers(context.TODO(), epochID, pageNum, pageSize)
	case txs:
		response, total, err = cc.Service.GetEpochTransactions(context.TODO(), epochID, pageNum, pageSize)
	case smeshers:
		response, total, err = cc.Service.GetEpochSmeshers(context.TODO(), epochID, pageNum, pageSize)
	case rewards:
		response, total, err = cc.Service.GetEpochRewards(context.TODO(), epochID, pageNum, pageSize)
	case atxs:
		response, total, err = cc.Service.GetEpochActivations(context.TODO(), epochID, pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get epoch entity `%s` list: %w", c.Param("entity"), err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       response,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}
