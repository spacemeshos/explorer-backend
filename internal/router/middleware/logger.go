package logger

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spacemeshos/go-spacemesh/log"
)

// New creates a new middleware handler.
func New(config ...Config) fiber.Handler {
	// Set default config
	cfg := configDefault(config...)

	// Set variables
	var (
		once       sync.Once
		errHandler fiber.ErrorHandler
	)

	// Return new handler
	return func(ctx *fiber.Ctx) (err error) {
		// Don't execute middleware if Next returns true
		if cfg.Next != nil && cfg.Next(ctx) {
			return ctx.Next()
		}

		once.Do(func() {
			errHandler = ctx.App().ErrorHandler
		})

		// Handle request, store err for logging
		chainErr := ctx.Next()

		// Manually call error handler
		if chainErr != nil {
			_ = errHandler(ctx, chainErr)
		}
		log.Info("%s [%d] - %s", time.Now().Format(time.RFC3339), ctx.Response().StatusCode(), ctx.Request().URI().PathOriginal())
		return nil
	}
}
