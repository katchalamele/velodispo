package store

import (
	"context"
	"fmt"

	"github.com/katchalamele/velodispo/internal/domain"
	"gorm.io/gorm/clause"
)

func (s *Store) UpsertStations(ctx context.Context, cityID int64, stations []domain.Station) error {
	if len(stations) == 0 {
		return nil
	}

	models := make([]Station, 0, len(stations))
	for _, st := range stations {
		models = append(models, stationModel(cityID, st))
	}

	err := s.GORM.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "city_id"}, {Name: "station_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "lat", "lon", "address", "capacity", "updated_at",
			}),
		}).
		Create(&models).Error
	if err != nil {
		return fmt.Errorf("upsert stations (ville %d): %w", cityID, err)
	}
	return nil
}

func (s *Store) StationsByCity(ctx context.Context, cityID int64) ([]Station, error) {
	var models []Station
	if err := s.GORM.WithContext(ctx).Where("city_id = ?", cityID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("lecture stations (ville %d): %w", cityID, err)
	}
	return models, nil
}
