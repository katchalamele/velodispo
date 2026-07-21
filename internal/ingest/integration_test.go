//go:build integration

package ingest

import (
	"context"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/config"
	"github.com/katchalamele/velodispo/internal/domain"
	"github.com/katchalamele/velodispo/internal/source"
	"github.com/katchalamele/velodispo/internal/store"
)

func TestPollOncePersists(t *testing.T) {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	if err := store.Migrate(cfg.DB.DSN()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	st, err := store.Open(ctx, cfg.DB.DSN())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	if _, err := st.Pool.Exec(ctx, "TRUNCATE station_status, stations, cities RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	src := &fakeSource{
		city:     "Nantes",
		stations: []domain.Station{{ID: "1", Name: "PRÉFECTURE", Lat: 47.2, Lon: -1.5, Capacity: 33}},
		statuses: []domain.Status{{StationID: "1", BikesAvailable: 26, DocksAvailable: 7, IsInstalled: true, IsRenting: true, IsReturning: true}},
	}

	p := New([]source.Source{src}, st, time.Minute, 10*time.Second, 4)
	p.pollOnce(ctx)

	views, total, err := st.ListStations(ctx, store.StationFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListStations: %v", err)
	}
	if total != 1 || len(views) != 1 {
		t.Fatalf("stations = %d, attendu 1", total)
	}
	if views[0].Status == nil || views[0].Status.BikesAvailable != 26 {
		t.Fatalf("dernier statut absent ou incorrect: %+v", views[0].Status)
	}

	pk := views[0].ID
	hist, err := st.History(ctx, pk, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if len(hist) != 1 {
		t.Errorf("historique = %d entrées, attendu 1", len(hist))
	}
}
