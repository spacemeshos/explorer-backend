package handler

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

func Search(c echo.Context) error {
	cc := c.(*ApiContext)

	search := strings.ToLower(c.Param("id"))
	redirectURL, err := cc.Service.Search(context.TODO(), search)
	if err != nil {
		return fmt.Errorf("error search `%s`: %w", search, err)
	}

	return c.JSON(http.StatusOK, RedirectResponse{
		Redirect: redirectURL,
	})
}
