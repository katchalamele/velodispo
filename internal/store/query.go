package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("station introuvable")

type StationView struct {
	Station
	CityName string
	Status   *Snapshot
}

type StationFilter struct {
	City   string
	Limit  int
	Offset int
}

const stationSelect = `
	SELECT s.id, s.city_id, s.station_id, s.name, s.lat, s.lon, s.address,
	       s.capacity, s.created_at, s.updated_at, c.name AS city_name,
	       st.time, st.bikes_available, st.docks_available, st.bikes_disabled,
	       st.docks_disabled, st.is_installed, st.is_renting, st.is_returning,
	       st.last_reported
	FROM stations s
	JOIN cities c ON c.id = s.city_id
	LEFT JOIN LATERAL (
		SELECT * FROM station_status ss
		WHERE ss.station_pk = s.id
		ORDER BY time DESC
		LIMIT 1
	) st ON true`

func (s *Store) ListStations(ctx context.Context, f StationFilter) ([]StationView, int, error) {
	var total int
	countQ := `SELECT count(*) FROM stations s JOIN cities c ON c.id = s.city_id WHERE ($1 = '' OR c.name = $1)`
	if err := s.Pool.QueryRow(ctx, countQ, f.City).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("comptage stations: %w", err)
	}

	q := stationSelect + `
		WHERE ($1 = '' OR c.name = $1)
		ORDER BY s.id
		LIMIT $2 OFFSET $3`

	rows, err := s.Pool.Query(ctx, q, f.City, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("liste stations: %w", err)
	}
	defer rows.Close()

	views, err := scanStationViews(rows)
	if err != nil {
		return nil, 0, err
	}
	return views, total, nil
}

func (s *Store) GetStation(ctx context.Context, id int64) (StationView, error) {
	rows, err := s.Pool.Query(ctx, stationSelect+` WHERE s.id = $1`, id)
	if err != nil {
		return StationView{}, fmt.Errorf("lecture station %d: %w", id, err)
	}
	defer rows.Close()

	views, err := scanStationViews(rows)
	if err != nil {
		return StationView{}, err
	}
	if len(views) == 0 {
		return StationView{}, ErrNotFound
	}
	return views[0], nil
}

func scanStationViews(rows pgx.Rows) ([]StationView, error) {
	var out []StationView
	for rows.Next() {
		var v StationView
		var (
			snapTime       *time.Time
			bikesAvailable *int
			docksAvailable *int
			bikesDisabled  *int
			docksDisabled  *int
			isInstalled    *bool
			isRenting      *bool
			isReturning    *bool
			lastReported   *time.Time
		)
		if err := rows.Scan(
			&v.ID, &v.CityID, &v.StationID, &v.Name, &v.Lat, &v.Lon, &v.Address,
			&v.Capacity, &v.CreatedAt, &v.UpdatedAt, &v.CityName,
			&snapTime, &bikesAvailable, &docksAvailable, &bikesDisabled,
			&docksDisabled, &isInstalled, &isRenting, &isReturning, &lastReported,
		); err != nil {
			return nil, fmt.Errorf("scan station: %w", err)
		}
		if snapTime != nil {
			st := Snapshot{
				StationPK:      v.ID,
				Time:           *snapTime,
				BikesAvailable: deref(bikesAvailable),
				DocksAvailable: deref(docksAvailable),
				BikesDisabled:  deref(bikesDisabled),
				DocksDisabled:  deref(docksDisabled),
				IsInstalled:    deref(isInstalled),
				IsRenting:      deref(isRenting),
				IsReturning:    deref(isReturning),
			}
			if lastReported != nil {
				st.LastReported = *lastReported
			}
			v.Status = &st
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours stations: %w", err)
	}
	return out, nil
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
