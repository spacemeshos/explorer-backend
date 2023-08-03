package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/spacemeshos/go-spacemesh/log"

	"github.com/spacemeshos/explorer-backend/model"
)

func Smeshers(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	smeshersList, total, err := cc.Service.GetSmeshers(context.TODO(), pageNum, pageSize)
	if err != nil {
		log.Err(fmt.Errorf("failed to get smeshers list: %s", err))
		return err
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       smeshersList,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Smesher(c echo.Context) error {
	cc := c.(*ApiContext)
	smesher, err := cc.Service.GetSmesher(context.TODO(), c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get smesher: %w", err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Smesher{smesher}})
}

func SmesherDetails(c echo.Context) error {
	cc := c.(*ApiContext)
	var (
		response interface{}
		err      error
		total    int64
	)
	pageNum, pageSize := GetPagination(c)
	switch c.Param("entity") {
	case atxs:
		response, total, err = cc.Service.GetSmesherActivations(context.TODO(), c.Param("id"), pageNum, pageSize)
	case rewards:
		response, total, err = cc.Service.GetSmesherRewards(context.TODO(), c.Param("id"), pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		log.Err(fmt.Errorf("failed to get smesher entity `%s` details: %s", c.Param("entity"), err))
		return err
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       response,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}
