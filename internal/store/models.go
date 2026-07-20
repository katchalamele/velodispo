package store

import (
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

type City struct {
	ID        int64 `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (City) TableName() string { return "cities" }

type Station struct {
	ID        int64 `gorm:"primaryKey"`
	CityID    int64
	StationID string
	Name      string
	Lat       float64
	Lon       float64
	Address   string
	Capacity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Station) TableName() string { return "stations" }

func stationModel(cityID int64, s domain.Station) Station {
	return Station{
		CityID:    cityID,
		StationID: s.ID,
		Name:      s.Name,
		Lat:       s.Lat,
		Lon:       s.Lon,
		Address:   s.Address,
		Capacity:  s.Capacity,
	}
}

func (s Station) toDomain(city string) domain.Station {
	return domain.Station{
		ID:       s.StationID,
		City:     city,
		Name:     s.Name,
		Lat:      s.Lat,
		Lon:      s.Lon,
		Address:  s.Address,
		Capacity: s.Capacity,
	}
}
