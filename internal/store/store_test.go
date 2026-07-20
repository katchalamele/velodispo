package store

import (
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

func TestStationModelRoundTrip(t *testing.T) {
	src := domain.Station{
		ID:       "42",
		City:     "Nantes",
		Name:     "PRÉFECTURE",
		Lat:      47.21984,
		Lon:      -1.554891,
		Address:  "Place du Port Communeau",
		Capacity: 33,
	}

	m := stationModel(7, src)
	if m.CityID != 7 || m.StationID != "42" {
		t.Fatalf("stationModel: city_id/station_id = %d/%q", m.CityID, m.StationID)
	}

	got := m.toDomain("Nantes")
	src.City = "Nantes"
	if got != src {
		t.Errorf("round-trip = %+v, attendu %+v", got, src)
	}
}

func TestSnapshotFrom(t *testing.T) {
	at := time.Unix(1784561205, 0).UTC()
	status := domain.Status{
		StationID:      "42",
		BikesAvailable: 26,
		DocksAvailable: 7,
		IsInstalled:    true,
		IsRenting:      true,
		IsReturning:    false,
		LastReported:   time.Unix(1784510482, 0).UTC(),
	}

	sn := SnapshotFrom(3, at, status)
	if sn.StationPK != 3 || !sn.Time.Equal(at) {
		t.Fatalf("SnapshotFrom pk/time = %d/%v", sn.StationPK, sn.Time)
	}
	if sn.BikesAvailable != 26 || sn.DocksAvailable != 7 || !sn.IsRenting || sn.IsReturning {
		t.Errorf("SnapshotFrom champs incohérents: %+v", sn)
	}
}

func TestMigrateURL(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"postgres://u:p@db:5432/x?sslmode=disable", "pgx5://u:p@db:5432/x?sslmode=disable"},
		{"postgresql://u:p@db:5432/x", "pgx5://u:p@db:5432/x"},
		{"pgx5://deja", "pgx5://deja"},
	}
	for _, tc := range tests {
		if got := migrateURL(tc.in); got != tc.want {
			t.Errorf("migrateURL(%q) = %q, attendu %q", tc.in, got, tc.want)
		}
	}
}
