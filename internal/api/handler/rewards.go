package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/spacemeshos/explorer-backend/model"
)

func Rewards(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	rewardsList, total, err := cc.Service.GetRewards(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get rewards info: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       rewardsList,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Reward(c echo.Context) error {
	cc := c.(*ApiContext)
	reward, err := cc.Service.GetReward(context.TODO(), c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get reward `%s` info: %w", c.Param("id"), err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Reward{reward}})
}
