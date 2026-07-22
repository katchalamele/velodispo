package api

import (
	"net/http"

	_ "github.com/katchalamele/velodispo/internal/api/docs"
	"github.com/katchalamele/velodispo/internal/web"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func New(reader StationReader) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = httpErrorHandler

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	h := NewHandler(reader)

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/stations", h.ListStations)
	e.GET("/stations/:id", h.GetStation)
	e.GET("/stations/:id/history", h.GetStationHistory)
	e.GET("/stations/:id/prediction", h.GetPrediction)
	e.GET("/map", h.GetMap)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.StaticFS("/", echo.MustSubFS(web.Assets, "static"))

	return e
}
