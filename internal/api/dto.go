package api

import (
	"time"

	"github.com/katchalamele/velodispo/internal/store"
)

type StatusResponse struct {
	Time           time.Time  `json:"time"`
	BikesAvailable int        `json:"bikes_available"`
	DocksAvailable int        `json:"docks_available"`
	BikesDisabled  int        `json:"bikes_disabled"`
	DocksDisabled  int        `json:"docks_disabled"`
	IsInstalled    bool       `json:"is_installed"`
	IsRenting      bool       `json:"is_renting"`
	IsReturning    bool       `json:"is_returning"`
	LastReported   *time.Time `json:"last_reported,omitempty"`
}

type StationResponse struct {
	ID        int64           `json:"id"`
	StationID string          `json:"station_id"`
	City      string          `json:"city"`
	Name      string          `json:"name"`
	Lat       float64         `json:"lat"`
	Lon       float64         `json:"lon"`
	Address   string          `json:"address"`
	Capacity  int             `json:"capacity"`
	Status    *StatusResponse `json:"status"`
}

type Pagination struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type ListResponse struct {
	Data       []StationResponse `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

type HistoryResponse struct {
	StationID int64            `json:"station_id"`
	From      time.Time        `json:"from"`
	To        time.Time        `json:"to"`
	Data      []StatusResponse `json:"data"`
}

func statusResponse(s *store.Snapshot) *StatusResponse {
	if s == nil {
		return nil
	}
	r := StatusResponse{
		Time:           s.Time.UTC(),
		BikesAvailable: s.BikesAvailable,
		DocksAvailable: s.DocksAvailable,
		BikesDisabled:  s.BikesDisabled,
		DocksDisabled:  s.DocksDisabled,
		IsInstalled:    s.IsInstalled,
		IsRenting:      s.IsRenting,
		IsReturning:    s.IsReturning,
	}
	if !s.LastReported.IsZero() {
		lr := s.LastReported.UTC()
		r.LastReported = &lr
	}
	return &r
}

func stationResponse(v store.StationView) StationResponse {
	return StationResponse{
		ID:        v.ID,
		StationID: v.StationID,
		City:      v.CityName,
		Name:      v.Name,
		Lat:       v.Lat,
		Lon:       v.Lon,
		Address:   v.Address,
		Capacity:  v.Capacity,
		Status:    statusResponse(v.Status),
	}
}
