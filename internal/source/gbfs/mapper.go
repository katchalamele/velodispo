package gbfs

import (
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

func toStation(city string, s stationInformation) domain.Station {
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

func toStatus(s stationStatus) domain.Status {
	var reported time.Time
	if s.LastReported > 0 {
		reported = time.Unix(s.LastReported, 0).UTC()
	}
	return domain.Status{
		StationID:      s.StationID,
		BikesAvailable: s.NumBikesAvailable,
		DocksAvailable: s.NumDocksAvailable,
		BikesDisabled:  s.NumBikesDisabled,
		DocksDisabled:  s.NumDocksDisabled,
		IsInstalled:    bool(s.IsInstalled),
		IsRenting:      bool(s.IsRenting),
		IsReturning:    bool(s.IsReturning),
		LastReported:   reported,
	}
}
