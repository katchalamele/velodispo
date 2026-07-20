package gbfs

import (
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

// toStation convertit une station GBFS brute vers le modèle de domaine. Le nom de
// la ville provient de la config, pas du flux.
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

// toStatus convertit un statut GBFS brut vers le modèle de domaine. L'horodatage
// last_reported est un epoch Unix (commun 1.x / 2.x) converti en time.Time UTC ;
// une valeur absente (0) donne un time.Time zéro plutôt qu'une date de 1970.
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
