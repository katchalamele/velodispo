package api

import (
	"time"

	"github.com/katchalamele/velodispo/internal/store"
)

type MapStationResponse struct {
	ID             int64   `json:"id"`
	City           string  `json:"city"`
	Name           string  `json:"name"`
	Lat            float64 `json:"lat"`
	Lon            float64 `json:"lon"`
	Capacity       int     `json:"capacity"`
	BikesAvailable *int    `json:"bikes_available"`
	DocksAvailable *int    `json:"docks_available"`
}

func mapStationResponse(v store.StationView) MapStationResponse {
	r := MapStationResponse{
		ID:       v.ID,
		City:     v.CityName,
		Name:     v.Name,
		Lat:      v.Lat,
		Lon:      v.Lon,
		Capacity: v.Capacity,
	}
	if v.Status != nil {
		bikes := v.Status.BikesAvailable
		docks := v.Status.DocksAvailable
		r.BikesAvailable = &bikes
		r.DocksAvailable = &docks
	}
	return r
}

type SlotResponse struct {
	Slot     int     `json:"slot"`
	Minutes  int     `json:"minutes"`
	AvgBikes float64 `json:"avg_bikes"`
	AvgDocks float64 `json:"avg_docks"`
	Samples  int     `json:"samples"`
}

type PredictedPoint struct {
	BikesAvailable int `json:"bikes_available"`
	DocksAvailable int `json:"docks_available"`
	Samples        int `json:"samples"`
}

type PredictionResponse struct {
	StationID int64          `json:"station_id"`
	At        time.Time      `json:"at"`
	Weekday   int            `json:"weekday"`
	Slot      int            `json:"slot"`
	Predicted PredictedPoint `json:"predicted"`
	Profile   []SlotResponse `json:"profile"`
}

const slotsPerDay = 48

func predictionResponse(id int64, at time.Time, weekday, slot int, profile []store.SlotAvg) PredictionResponse {
	dense := make([]SlotResponse, slotsPerDay)
	for i := range dense {
		dense[i] = SlotResponse{Slot: i, Minutes: i * 30}
	}
	for _, a := range profile {
		if a.Slot < 0 || a.Slot >= slotsPerDay {
			continue
		}
		dense[a.Slot] = SlotResponse{
			Slot:     a.Slot,
			Minutes:  a.Slot * 30,
			AvgBikes: a.AvgBikes,
			AvgDocks: a.AvgDocks,
			Samples:  a.Samples,
		}
	}

	cur := dense[slot]
	return PredictionResponse{
		StationID: id,
		At:        at.UTC(),
		Weekday:   weekday,
		Slot:      slot,
		Predicted: PredictedPoint{
			BikesAvailable: int(cur.AvgBikes + 0.5),
			DocksAvailable: int(cur.AvgDocks + 0.5),
			Samples:        cur.Samples,
		},
		Profile: dense,
	}
}
