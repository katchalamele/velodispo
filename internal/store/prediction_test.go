//go:build integration

package store

import (
	"context"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

func TestPredictionProfileBucketing(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	city, err := st.UpsertCity(ctx, "Nantes")
	if err != nil {
		t.Fatalf("UpsertCity: %v", err)
	}
	if err := st.UpsertStations(ctx, city.ID, []domain.Station{{ID: "1", Name: "A"}}); err != nil {
		t.Fatalf("UpsertStations: %v", err)
	}
	models, _ := st.StationsByCity(ctx, city.ID)
	pk := models[0].ID

	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		t.Fatalf("tz: %v", err)
	}
	// Deux lundis, même tranche 08h00-08h30 (slot 16) -> moyenne attendue (10+20)/2 = 15.
	lundiA := time.Date(2026, 7, 6, 8, 5, 0, 0, loc)
	lundiB := time.Date(2026, 7, 13, 8, 20, 0, 0, loc)
	// Un mardi, même heure -> ne doit PAS entrer dans le bucket du lundi.
	mardi := time.Date(2026, 7, 7, 8, 15, 0, 0, loc)

	snaps := []Snapshot{
		SnapshotFrom(pk, lundiA, domain.Status{BikesAvailable: 10, DocksAvailable: 5}),
		SnapshotFrom(pk, lundiB, domain.Status{BikesAvailable: 20, DocksAvailable: 15}),
		SnapshotFrom(pk, mardi, domain.Status{BikesAvailable: 99, DocksAvailable: 99}),
	}
	if _, err := st.InsertSnapshots(ctx, snaps); err != nil {
		t.Fatalf("InsertSnapshots: %v", err)
	}

	profile, err := st.PredictionProfile(ctx, pk, 1) // isodow 1 = lundi
	if err != nil {
		t.Fatalf("PredictionProfile: %v", err)
	}

	var slot16 *SlotAvg
	for i := range profile {
		if profile[i].Slot == 16 {
			slot16 = &profile[i]
		}
	}
	if slot16 == nil {
		t.Fatal("slot 16 (08h00-08h30) absent du profil lundi")
	}
	if slot16.Samples != 2 {
		t.Errorf("samples = %d, attendu 2 (le mardi exclu)", slot16.Samples)
	}
	if slot16.AvgBikes != 15 || slot16.AvgDocks != 10 {
		t.Errorf("moyennes = %.1f/%.1f, attendu 15/10", slot16.AvgBikes, slot16.AvgDocks)
	}
}

func TestMapStationsReturnsAll(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	city, _ := st.UpsertCity(ctx, "Nantes")
	if err := st.UpsertStations(ctx, city.ID, []domain.Station{
		{ID: "1", Name: "A"}, {ID: "2", Name: "B"}, {ID: "3", Name: "C"},
	}); err != nil {
		t.Fatalf("UpsertStations: %v", err)
	}

	views, err := st.MapStations(ctx)
	if err != nil {
		t.Fatalf("MapStations: %v", err)
	}
	if len(views) != 3 {
		t.Errorf("stations = %d, attendu 3", len(views))
	}
}
