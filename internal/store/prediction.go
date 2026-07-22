package store

import (
	"context"
	"fmt"
)

type SlotAvg struct {
	Slot     int
	AvgBikes float64
	AvgDocks float64
	Samples  int
}

func (s *Store) PredictionProfile(ctx context.Context, stationPK int64, isodow int) ([]SlotAvg, error) {
	const q = `
		SELECT (EXTRACT(hour FROM t) * 2 + floor(EXTRACT(minute FROM t) / 30))::int AS slot,
		       avg(bikes_available), avg(docks_available), count(*)
		FROM (
			SELECT time AT TIME ZONE 'Europe/Paris' AS t, bikes_available, docks_available
			FROM station_status
			WHERE station_pk = $1
		) s
		WHERE EXTRACT(isodow FROM t)::int = $2
		GROUP BY slot
		ORDER BY slot`

	rows, err := s.Pool.Query(ctx, q, stationPK, isodow)
	if err != nil {
		return nil, fmt.Errorf("profil prédiction (station %d): %w", stationPK, err)
	}
	defer rows.Close()

	var out []SlotAvg
	for rows.Next() {
		var a SlotAvg
		if err := rows.Scan(&a.Slot, &a.AvgBikes, &a.AvgDocks, &a.Samples); err != nil {
			return nil, fmt.Errorf("scan profil: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours profil: %w", err)
	}
	return out, nil
}
