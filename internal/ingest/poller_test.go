package ingest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
	"github.com/katchalamele/velodispo/internal/source"
	"github.com/katchalamele/velodispo/internal/store"
)

type fakeSource struct {
	city       string
	stations   []domain.Station
	statuses   []domain.Status
	statusErr  error
	stationErr error
}

func (f *fakeSource) City() string { return f.city }

func (f *fakeSource) Stations(context.Context) ([]domain.Station, error) {
	return f.stations, f.stationErr
}

func (f *fakeSource) Statuses(context.Context) ([]domain.Status, error) {
	return f.statuses, f.statusErr
}

type fakeIngester struct {
	mu       sync.Mutex
	cities   map[string]int64
	nextID   int64
	stations map[int64][]store.Station
	inserted []store.Snapshot
}

func newFakeIngester() *fakeIngester {
	return &fakeIngester{cities: map[string]int64{}, stations: map[int64][]store.Station{}}
}

func (f *fakeIngester) UpsertCity(_ context.Context, name string) (store.City, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.cities[name]
	if !ok {
		f.nextID++
		id = f.nextID * 100
		f.cities[name] = id
	}
	return store.City{ID: id, Name: name}, nil
}

func (f *fakeIngester) UpsertStations(_ context.Context, cityID int64, stations []domain.Station) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	models := make([]store.Station, 0, len(stations))
	for i, s := range stations {
		models = append(models, store.Station{
			ID:        cityID + int64(i) + 1,
			CityID:    cityID,
			StationID: s.ID,
			Name:      s.Name,
		})
	}
	f.stations[cityID] = models
	return nil
}

func (f *fakeIngester) StationsByCity(_ context.Context, cityID int64) ([]store.Station, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.stations[cityID], nil
}

func (f *fakeIngester) InsertSnapshots(_ context.Context, snaps []store.Snapshot) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inserted = append(f.inserted, snaps...)
	return int64(len(snaps)), nil
}

func TestPollOnceMapsStationIDToPK(t *testing.T) {
	ing := newFakeIngester()
	src := &fakeSource{
		city:     "Nantes",
		stations: []domain.Station{{ID: "1", Name: "A"}, {ID: "6", Name: "B"}},
		statuses: []domain.Status{
			{StationID: "1", BikesAvailable: 26},
			{StationID: "6", BikesAvailable: 10},
			{StationID: "999", BikesAvailable: 5},
		},
	}
	p := New([]source.Source{src}, ing, time.Minute, 5*time.Second, 4)
	p.pollOnce(context.Background())

	if len(ing.inserted) != 2 {
		t.Fatalf("snapshots insérés = %d, attendu 2 (statut orphelin ignoré)", len(ing.inserted))
	}

	cityID := ing.cities["Nantes"]
	pkByStationID := map[string]int64{}
	for _, m := range ing.stations[cityID] {
		pkByStationID[m.StationID] = m.ID
	}
	byPK := map[int64]int{}
	for _, s := range ing.inserted {
		byPK[s.StationPK] = s.BikesAvailable
	}
	if byPK[pkByStationID["1"]] != 26 || byPK[pkByStationID["6"]] != 10 {
		t.Errorf("mapping station_id -> PK incorrect: %+v", byPK)
	}
}

func TestPollOnceIsolatesFailingSource(t *testing.T) {
	ing := newFakeIngester()
	healthy := &fakeSource{
		city:     "Nantes",
		stations: []domain.Station{{ID: "1", Name: "A"}},
		statuses: []domain.Status{{StationID: "1", BikesAvailable: 26}},
	}
	broken := &fakeSource{city: "Paris", statusErr: errors.New("réseau mort")}

	p := New([]source.Source{broken, healthy}, ing, time.Minute, 5*time.Second, 4)
	p.pollOnce(context.Background())

	if len(ing.inserted) != 1 {
		t.Fatalf("snapshots = %d, attendu 1 (la source saine écrit malgré l'échec de l'autre)", len(ing.inserted))
	}
	if ing.inserted[0].BikesAvailable != 26 {
		t.Errorf("snapshot de la source saine incorrect: %+v", ing.inserted[0])
	}
}
