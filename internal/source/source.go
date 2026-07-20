package source

import (
	"context"

	"github.com/katchalamele/velodispo/internal/domain"
)

type Source interface {
	City() string
	Stations(ctx context.Context) ([]domain.Station, error)
	Statuses(ctx context.Context) ([]domain.Status, error)
}
