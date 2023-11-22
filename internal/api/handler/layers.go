package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func Layers(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	layersList, total, err := cc.Service.GetLayers(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get epoch list: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       layersList,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Layer(c echo.Context) error {
	cc := c.(*ApiContext)
	layerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	layer, err := cc.Service.GetLayer(context.TODO(), layerID)
	if err != nil {
		if err == service.ErrNotFound {
			return echo.ErrNotFound
		}
		return fmt.Errorf("failed to get layer info: %w", err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Layer{layer}})
}

func LayerDetails(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	layerID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return fmt.Errorf("wrong layer id: %w", err)
	}
	var (
		response interface{}
		total    int64
	)

	switch c.Param("entity") {
	case blocks:
		response, total, err = cc.Service.GetLayerBlocks(context.TODO(), layerID, pageNum, pageSize)
	case txs:
		response, total, err = cc.Service.GetLayerTransactions(context.TODO(), layerID, pageNum, pageSize)
	case smeshers:
		response, total, err = cc.Service.GetLayerSmeshers(context.TODO(), layerID, pageNum, pageSize)
	case rewards:
		response, total, err = cc.Service.GetLayerRewards(context.TODO(), layerID, pageNum, pageSize)
	case atxs:
		response, total, err = cc.Service.GetLayerActivations(context.TODO(), layerID, pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get layer entity `%s` list: %w", c.Param("entity"), err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       response,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}
