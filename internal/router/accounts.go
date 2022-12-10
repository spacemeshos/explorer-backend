package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/spacemeshos/explorer-backend/model"
)

func (a *AppRouter) accounts(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	accounts, total, err := a.appService.GetAccounts(ctx.UserContext(), pageNum, pageSize)
	if err != nil {
		return fmt.Errorf("failed to get accounts list: %w", err)
	}
	return ctx.JSON(fiber.Map{"data": accounts, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}

func (a *AppRouter) account(ctx *fiber.Ctx) error {
	account, err := a.appService.GetAccount(ctx.UserContext(), ctx.Params("id"))
	if err != nil {
		return fmt.Errorf("failed to get account `%s` info: %w", ctx.Params("id"), err)
	}
	return ctx.JSON(fiber.Map{"data": []*model.Account{account}})
}

func (a *AppRouter) accountDetails(ctx *fiber.Ctx) error {
	pageNum, pageSize := a.getPagination(ctx)
	accountID := ctx.Params("id")
	var (
		response interface{}
		err      error
		total    int64
	)

	switch ctx.Params("entity") {
	case txs:
		response, total, err = a.appService.GetAccountTransactions(ctx.UserContext(), accountID, pageNum, pageSize)
	case rewards:
		response, total, err = a.appService.GetAccountRewards(ctx.UserContext(), accountID, pageNum, pageSize)
	default:
		return fiber.NewError(fiber.StatusNotFound, "entity not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get account entity `%s` list: %w", ctx.Params("entity"), err)
	}
	return ctx.JSON(fiber.Map{"data": response, "pagination": a.setPaginationResponse(total, pageNum, pageSize)})
}
