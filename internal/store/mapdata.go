package store

import (
	"context"
	"fmt"
)

func (s *Store) MapStations(ctx context.Context) ([]StationView, error) {
	rows, err := s.Pool.Query(ctx, stationSelect+` ORDER BY s.id`)
	if err != nil {
		return nil, fmt.Errorf("map stations: %w", err)
	}
	defer rows.Close()

	return scanStationViews(rows)
}
