package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spacemeshos/explorer-backend/internal/service"
	"github.com/spacemeshos/explorer-backend/model"
	"net/http"
)

func Accounts(c echo.Context) error {
	cc := c.(*ApiContext)

	pageNum, pageSize := GetPagination(c)
	accounts, total, err := cc.Service.GetAccounts(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get accounts list: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       accounts,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Account(c echo.Context) error {
	cc := c.(*ApiContext)

	account, err := cc.Service.GetAccount(context.TODO(), c.Param("id"))
	if err != nil {
		if err == service.ErrNotFound {
			return echo.ErrNotFound
		}
		return fmt.Errorf("failed to get account `%s` info: %w", c.Param("id"), err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Account{account}})
}

func AccountDetails(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(cc)
	accountID := c.Param("id")
	var (
		response interface{}
		err      error
		total    int64
	)

	switch c.Param("entity") {
	case txs:
		response, total, err = cc.Service.GetAccountTransactions(context.TODO(), accountID, pageNum, pageSize)
	case rewards:
		response, total, err = cc.Service.GetAccountRewards(context.TODO(), accountID, pageNum, pageSize)
	default:
		return echo.NewHTTPError(http.StatusNotFound, "entity not found")
	}
	if err != nil {
		if err == service.ErrNotFound {
			return echo.ErrNotFound
		}
		return fmt.Errorf("failed to get account entity `%s` list: %w", c.Param("entity"), err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       response,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}
