package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type errorResponse struct {
	Error string `json:"error"`
}

func httpErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	msg := "erreur interne"
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		}
	}

	_ = c.JSON(code, errorResponse{Error: msg})
}
