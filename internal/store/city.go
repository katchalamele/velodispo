package store

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"
)

func (s *Store) UpsertCity(ctx context.Context, name string) (City, error) {
	city := City{Name: name}
	err := s.GORM.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
		}).
		Create(&city).Error
	if err != nil {
		return City{}, fmt.Errorf("upsert ville %q: %w", name, err)
	}
	return city, nil
}
