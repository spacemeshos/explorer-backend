package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) transactions(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	txs, total, err := a.appService.GetTransactions(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get transactions list: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": txs, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) transaction(ctx *fiber.Ctx) error {
	tx, err := a.appService.GetTransaction(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get transaction %s list: %s", ctx.Params("id"), err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Transaction{tx}})
}
