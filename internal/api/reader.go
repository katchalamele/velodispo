package api

import (
	"context"
	"time"

	"github.com/katchalamele/velodispo/internal/store"
)

type StationReader interface {
	ListStations(ctx context.Context, f store.StationFilter) ([]store.StationView, int, error)
	GetStation(ctx context.Context, id int64) (store.StationView, error)
	History(ctx context.Context, stationPK int64, from, to time.Time) ([]store.Snapshot, error)
	PredictionProfile(ctx context.Context, stationPK int64, isodow int) ([]store.SlotAvg, error)
	MapStations(ctx context.Context) ([]store.StationView, error)
}
