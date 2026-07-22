package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	defaultLimit    = 50
	maxLimit        = 200
	defaultHistory  = 24 * time.Hour
	maxHistoryRange = 31 * 24 * time.Hour
)

func parsePagination(c echo.Context) (limit, offset int, err error) {
	limit = defaultLimit
	if raw := c.QueryParam("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "limit invalide")
		}
		if limit > maxLimit {
			limit = maxLimit
		}
	}
	if raw := c.QueryParam("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "offset invalide")
		}
	}
	return limit, offset, nil
}

func parseID(c echo.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "id invalide")
	}
	return id, nil
}

func parseAt(c echo.Context, now time.Time) (time.Time, error) {
	raw := c.QueryParam("at")
	if raw == "" {
		return now, nil
	}
	at, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, echo.NewHTTPError(http.StatusBadRequest, "at invalide (RFC3339 attendu)")
	}
	return at, nil
}

func parseInterval(c echo.Context, now time.Time) (from, to time.Time, err error) {
	to = now.UTC()
	from = to.Add(-defaultHistory)

	if raw := c.QueryParam("from"); raw != "" {
		from, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, echo.NewHTTPError(http.StatusBadRequest, "from invalide (RFC3339 attendu)")
		}
	}
	if raw := c.QueryParam("to"); raw != "" {
		to, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, echo.NewHTTPError(http.StatusBadRequest, "to invalide (RFC3339 attendu)")
		}
	}
	if !from.Before(to) {
		return time.Time{}, time.Time{}, echo.NewHTTPError(http.StatusBadRequest, "from doit précéder to")
	}
	if to.Sub(from) > maxHistoryRange {
		return time.Time{}, time.Time{}, echo.NewHTTPError(http.StatusBadRequest, "plage trop large (max 31 jours)")
	}
	return from.UTC(), to.UTC(), nil
}
