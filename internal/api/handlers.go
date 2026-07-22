package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/katchalamele/velodispo/internal/store"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	reader StationReader
	now    func() time.Time
	loc    *time.Location
}

func NewHandler(reader StationReader) *Handler {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		loc = time.UTC
	}
	return &Handler{reader: reader, now: time.Now, loc: loc}
}

// ListStations godoc
// @Summary  Liste les stations
// @Description Stations paginées, filtrables par ville, avec la dernière disponibilité connue.
// @Tags     stations
// @Produce  json
// @Param    city   query  string  false  "Filtre par nom de ville"
// @Param    limit  query  int     false  "Taille de page (défaut 50, max 200)"
// @Param    offset query  int     false  "Décalage (défaut 0)"
// @Success  200 {object} ListResponse
// @Failure  400 {object} errorResponse
// @Failure  500 {object} errorResponse
// @Router   /stations [get]
func (h *Handler) ListStations(c echo.Context) error {
	limit, offset, err := parsePagination(c)
	if err != nil {
		return err
	}

	views, total, err := h.reader.ListStations(c.Request().Context(), store.StationFilter{
		City:   c.QueryParam("city"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return err
	}

	data := make([]StationResponse, 0, len(views))
	for _, v := range views {
		data = append(data, stationResponse(v))
	}

	return c.JSON(http.StatusOK, ListResponse{
		Data:       data,
		Pagination: Pagination{Limit: limit, Offset: offset, Total: total},
	})
}

// GetStation godoc
// @Summary  Détail d'une station
// @Tags     stations
// @Produce  json
// @Param    id  path  int  true  "Identifiant interne de la station"
// @Success  200 {object} StationResponse
// @Failure  400 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Failure  500 {object} errorResponse
// @Router   /stations/{id} [get]
func (h *Handler) GetStation(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	view, err := h.reader.GetStation(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "station introuvable")
		}
		return err
	}
	return c.JSON(http.StatusOK, stationResponse(view))
}

// GetStationHistory godoc
// @Summary  Historique de disponibilité d'une station
// @Tags     stations
// @Produce  json
// @Param    id    path   int     true   "Identifiant interne de la station"
// @Param    from  query  string  false  "Début (RFC3339, défaut il y a 24h)"
// @Param    to    query  string  false  "Fin (RFC3339, défaut maintenant)"
// @Success  200 {object} HistoryResponse
// @Failure  400 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Failure  500 {object} errorResponse
// @Router   /stations/{id}/history [get]
func (h *Handler) GetStationHistory(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	from, to, err := parseInterval(c, h.now())
	if err != nil {
		return err
	}

	if _, err := h.reader.GetStation(c.Request().Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "station introuvable")
		}
		return err
	}

	snaps, err := h.reader.History(c.Request().Context(), id, from, to)
	if err != nil {
		return err
	}

	data := make([]StatusResponse, 0, len(snaps))
	for i := range snaps {
		data = append(data, *statusResponse(&snaps[i]))
	}

	return c.JSON(http.StatusOK, HistoryResponse{
		StationID: id,
		From:      from,
		To:        to,
		Data:      data,
	})
}

// GetPrediction godoc
// @Summary  Prédiction de disponibilité d'une station
// @Description Moyenne par (jour de semaine, tranche de 30 min) en heure locale Europe/Paris.
// @Tags     stations
// @Produce  json
// @Param    id  path   int     true   "Identifiant interne de la station"
// @Param    at  query  string  false  "Instant cible (RFC3339, défaut maintenant)"
// @Success  200 {object} PredictionResponse
// @Failure  400 {object} errorResponse
// @Failure  404 {object} errorResponse
// @Failure  500 {object} errorResponse
// @Router   /stations/{id}/prediction [get]
func (h *Handler) GetPrediction(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	at, err := parseAt(c, h.now())
	if err != nil {
		return err
	}

	if _, err := h.reader.GetStation(c.Request().Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "station introuvable")
		}
		return err
	}

	local := at.In(h.loc)
	weekday := int(local.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	slot := local.Hour()*2 + local.Minute()/30

	profile, err := h.reader.PredictionProfile(c.Request().Context(), id, weekday)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, predictionResponse(id, at, weekday, slot, profile))
}

// GetMap godoc
// @Summary  Toutes les stations pour la carte
// @Description Position et dernière disponibilité de toutes les stations (sans pagination).
// @Tags     stations
// @Produce  json
// @Success  200 {array} MapStationResponse
// @Failure  500 {object} errorResponse
// @Router   /map [get]
func (h *Handler) GetMap(c echo.Context) error {
	views, err := h.reader.MapStations(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]MapStationResponse, 0, len(views))
	for _, v := range views {
		data = append(data, mapStationResponse(v))
	}
	return c.JSON(http.StatusOK, data)
}
