package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/katchalamele/velodispo/internal/domain"
)

type Snapshot struct {
	StationPK      int64
	Time           time.Time
	BikesAvailable int
	DocksAvailable int
	BikesDisabled  int
	DocksDisabled  int
	IsInstalled    bool
	IsRenting      bool
	IsReturning    bool
	LastReported   time.Time
}

func SnapshotFrom(stationPK int64, at time.Time, s domain.Status) Snapshot {
	return Snapshot{
		StationPK:      stationPK,
		Time:           at,
		BikesAvailable: s.BikesAvailable,
		DocksAvailable: s.DocksAvailable,
		BikesDisabled:  s.BikesDisabled,
		DocksDisabled:  s.DocksDisabled,
		IsInstalled:    s.IsInstalled,
		IsRenting:      s.IsRenting,
		IsReturning:    s.IsReturning,
		LastReported:   s.LastReported,
	}
}

var snapshotColumns = []string{
	"time", "station_pk", "bikes_available", "docks_available",
	"bikes_disabled", "docks_disabled", "is_installed", "is_renting",
	"is_returning", "last_reported",
}

func (s *Store) InsertSnapshots(ctx context.Context, snaps []Snapshot) (int64, error) {
	if len(snaps) == 0 {
		return 0, nil
	}

	rows := make([][]any, len(snaps))
	for i, sn := range snaps {
		var lastReported *time.Time
		if !sn.LastReported.IsZero() {
			lastReported = &sn.LastReported
		}
		rows[i] = []any{
			sn.Time, sn.StationPK, sn.BikesAvailable, sn.DocksAvailable,
			sn.BikesDisabled, sn.DocksDisabled, sn.IsInstalled, sn.IsRenting,
			sn.IsReturning, lastReported,
		}
	}

	n, err := s.Pool.CopyFrom(ctx, pgx.Identifier{"station_status"}, snapshotColumns, pgx.CopyFromRows(rows))
	if err != nil {
		return 0, fmt.Errorf("insertion snapshots: %w", err)
	}
	return n, nil
}

func (s *Store) History(ctx context.Context, stationPK int64, from, to time.Time) ([]Snapshot, error) {
	const q = `
		SELECT time, station_pk, bikes_available, docks_available,
		       bikes_disabled, docks_disabled, is_installed, is_renting,
		       is_returning, last_reported
		FROM station_status
		WHERE station_pk = $1 AND time >= $2 AND time <= $3
		ORDER BY time`

	rows, err := s.Pool.Query(ctx, q, stationPK, from, to)
	if err != nil {
		return nil, fmt.Errorf("lecture historique (station %d): %w", stationPK, err)
	}
	defer rows.Close()

	var out []Snapshot
	for rows.Next() {
		var sn Snapshot
		var lastReported *time.Time
		if err := rows.Scan(
			&sn.Time, &sn.StationPK, &sn.BikesAvailable, &sn.DocksAvailable,
			&sn.BikesDisabled, &sn.DocksDisabled, &sn.IsInstalled, &sn.IsRenting,
			&sn.IsReturning, &lastReported,
		); err != nil {
			return nil, fmt.Errorf("scan historique: %w", err)
		}
		if lastReported != nil {
			sn.LastReported = *lastReported
		}
		out = append(out, sn)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("parcours historique: %w", err)
	}
	return out, nil
}
