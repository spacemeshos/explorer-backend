package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"

	"github.com/spacemeshos/explorer-backend/model"
)

func Transactions(c echo.Context) error {
	cc := c.(*ApiContext)
	pageNum, pageSize := GetPagination(c)
	txs, total, err := cc.Service.GetTransactions(context.TODO(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get transactions list: %w", err)
	}

	return c.JSON(http.StatusOK, PaginatedDataResponse{
		Data:       txs,
		Pagination: GetPaginationMetadata(total, pageNum, pageSize),
	})
}

func Transaction(c echo.Context) error {
	cc := c.(*ApiContext)
	tx, err := cc.Service.GetTransaction(context.TODO(), c.Param("id"))
	if err != nil {
		return fmt.Errorf("failed to get transaction %s list: %s", c.Param("id"), err)
	}

	return c.JSON(http.StatusOK, DataResponse{Data: []*model.Transaction{tx}})
}
