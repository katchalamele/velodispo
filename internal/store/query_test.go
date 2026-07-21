//go:build integration

package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

func TestListAndGetStations(t *testing.T) {
	ctx := context.Background()
	st := setup(t)

	city, err := st.UpsertCity(ctx, "Nantes")
	if err != nil {
		t.Fatalf("UpsertCity: %v", err)
	}
	stations := []domain.Station{
		{ID: "1", Name: "PRÉFECTURE", Lat: 47.21984, Lon: -1.554891, Capacity: 33},
		{ID: "6", Name: "COMMERCE", Lat: 47.214321, Lon: -1.560125, Capacity: 25},
	}
	if err := st.UpsertStations(ctx, city.ID, stations); err != nil {
		t.Fatalf("UpsertStations: %v", err)
	}

	persisted, err := st.StationsByCity(ctx, city.ID)
	if err != nil {
		t.Fatalf("StationsByCity: %v", err)
	}
	pk := map[string]int64{}
	for _, s := range persisted {
		pk[s.StationID] = s.ID
	}

	old := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	recent := time.Now().UTC().Truncate(time.Second)
	snaps := []Snapshot{
		SnapshotFrom(pk["1"], old, domain.Status{BikesAvailable: 5, DocksAvailable: 28, IsInstalled: true, IsRenting: true, IsReturning: true}),
		SnapshotFrom(pk["1"], recent, domain.Status{BikesAvailable: 26, DocksAvailable: 7, IsInstalled: true, IsRenting: true, IsReturning: true}),
	}
	if _, err := st.InsertSnapshots(ctx, snaps); err != nil {
		t.Fatalf("InsertSnapshots: %v", err)
	}

	views, total, err := st.ListStations(ctx, StationFilter{Limit: 10})
	if err != nil {
		t.Fatalf("ListStations: %v", err)
	}
	if total != 2 || len(views) != 2 {
		t.Fatalf("total/len = %d/%d, attendu 2/2", total, len(views))
	}

	var s1 *StationView
	for i := range views {
		if views[i].StationID == "1" {
			s1 = &views[i]
		}
	}
	if s1 == nil {
		t.Fatal("station 1 absente de la liste")
	}
	if s1.Status == nil {
		t.Fatal("dernier statut manquant pour la station 1")
	}
	if s1.Status.BikesAvailable != 26 || !s1.Status.Time.Equal(recent) {
		t.Errorf("dernier statut incorrect (doit être le plus récent): %+v", s1.Status)
	}
	if s1.CityName != "Nantes" {
		t.Errorf("city_name = %q, attendu Nantes", s1.CityName)
	}

	var s6 *StationView
	for i := range views {
		if views[i].StationID == "6" {
			s6 = &views[i]
		}
	}
	if s6 == nil || s6.Status != nil {
		t.Errorf("station 6 sans snapshot devrait avoir Status nil: %+v", s6)
	}

	filtered, total, err := st.ListStations(ctx, StationFilter{City: "Paris", Limit: 10})
	if err != nil {
		t.Fatalf("ListStations (filtre): %v", err)
	}
	if total != 0 || len(filtered) != 0 {
		t.Errorf("filtre ville inconnue = %d résultats, attendu 0", len(filtered))
	}

	got, err := st.GetStation(ctx, pk["1"])
	if err != nil {
		t.Fatalf("GetStation: %v", err)
	}
	if got.Name != "PRÉFECTURE" || got.Status == nil || got.Status.BikesAvailable != 26 {
		t.Errorf("GetStation incohérent: %+v", got)
	}

	if _, err := st.GetStation(ctx, 999999); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetStation(inexistant) err = %v, attendu ErrNotFound", err)
	}
}
