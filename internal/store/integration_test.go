//go:build integration

package store

import (
	"context"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/config"
	"github.com/katchalamele/velodispo/internal/domain"
)

func setup(t *testing.T) *Store {
	t.Helper()
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if err := Migrate(cfg.DB.DSN()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	st, err := Open(ctx, cfg.DB.DSN())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	if _, err := st.Pool.Exec(ctx, "TRUNCATE station_status, stations, cities RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return st
}

func TestPersistenceRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	city, err := st.UpsertCity(ctx, "Nantes")
	if err != nil {
		t.Fatalf("UpsertCity: %v", err)
	}
	if city.ID == 0 {
		t.Fatal("city.ID vide après upsert")
	}

	again, err := st.UpsertCity(ctx, "Nantes")
	if err != nil {
		t.Fatalf("UpsertCity (2e): %v", err)
	}
	if again.ID != city.ID {
		t.Errorf("upsert non idempotent: %d != %d", again.ID, city.ID)
	}

	stations := []domain.Station{
		{ID: "1", Name: "PRÉFECTURE", Lat: 47.21984, Lon: -1.554891, Address: "Port Communeau", Capacity: 33},
		{ID: "6", Name: "COMMERCE", Lat: 47.214321, Lon: -1.560125, Capacity: 25},
	}
	if err := st.UpsertStations(ctx, city.ID, stations); err != nil {
		t.Fatalf("UpsertStations: %v", err)
	}
	if err := st.UpsertStations(ctx, city.ID, stations); err != nil {
		t.Fatalf("UpsertStations (2e): %v", err)
	}

	persisted, err := st.StationsByCity(ctx, city.ID)
	if err != nil {
		t.Fatalf("StationsByCity: %v", err)
	}
	if len(persisted) != 2 {
		t.Fatalf("stations persistées = %d, attendu 2 (idempotence)", len(persisted))
	}

	pkByStationID := map[string]int64{}
	for _, s := range persisted {
		pkByStationID[s.StationID] = s.ID
	}

	now := time.Now().UTC().Truncate(time.Second)
	snaps := []Snapshot{
		SnapshotFrom(pkByStationID["1"], now, domain.Status{BikesAvailable: 26, DocksAvailable: 7, IsInstalled: true, IsRenting: true, IsReturning: true, LastReported: now.Add(-time.Minute)}),
		SnapshotFrom(pkByStationID["6"], now, domain.Status{BikesAvailable: 10, DocksAvailable: 13, IsInstalled: true, IsRenting: true, IsReturning: true}),
	}
	n, err := st.InsertSnapshots(ctx, snaps)
	if err != nil {
		t.Fatalf("InsertSnapshots: %v", err)
	}
	if n != 2 {
		t.Errorf("snapshots insérés = %d, attendu 2", n)
	}

	hist, err := st.History(ctx, pkByStationID["1"], now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(hist) != 1 {
		t.Fatalf("historique station 1 = %d entrées, attendu 1", len(hist))
	}
	if hist[0].BikesAvailable != 26 || !hist[0].Time.Equal(now) {
		t.Errorf("snapshot relu incohérent: %+v", hist[0])
	}
}
