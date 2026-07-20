package gbfs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
)

func fixtureServer(t *testing.T, dir string) (*httptest.Server, *int64) {
	t.Helper()

	var discoveryHits int64
	mux := http.NewServeMux()
	var srv *httptest.Server

	serveFile := func(name string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			b, err := os.ReadFile(filepath.Join("testdata", dir, name))
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(b)
		}
	}

	mux.HandleFunc("/station_information.json", serveFile("station_information.json"))
	mux.HandleFunc("/station_status.json", serveFile("station_status.json"))

	mux.HandleFunc("/gbfs.json", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&discoveryHits, 1)
		raw, err := os.ReadFile(filepath.Join("testdata", dir, "gbfs.json"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rewriteDiscovery(t, raw, srv.URL))
	})

	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, &discoveryHits
}

// rewriteDiscovery fait pointer les URLs des flux vers le serveur de test.
func rewriteDiscovery(t *testing.T, raw []byte, base string) []byte {
	t.Helper()
	var file struct {
		LastUpdated int64  `json:"last_updated"`
		TTL         int    `json:"ttl"`
		Version     string `json:"version"`
		Data        map[string]struct {
			Feeds []feed `json:"feeds"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("rewriteDiscovery: %v", err)
	}
	for lang, set := range file.Data {
		for i := range set.Feeds {
			set.Feeds[i].URL = base + "/" + path.Base(set.Feeds[i].URL)
		}
		file.Data[lang] = set
	}
	out, err := json.Marshal(file)
	if err != nil {
		t.Fatalf("rewriteDiscovery marshal: %v", err)
	}
	return out
}

func indexStations(stations []domain.Station) map[string]domain.Station {
	m := make(map[string]domain.Station, len(stations))
	for _, s := range stations {
		m[s.ID] = s
	}
	return m
}

func indexStatuses(statuses []domain.Status) map[string]domain.Status {
	m := make(map[string]domain.Status, len(statuses))
	for _, s := range statuses {
		m[s.StationID] = s
	}
	return m
}

func TestClientStations(t *testing.T) {
	srv, _ := fixtureServer(t, "nantes")
	c := New(Config{City: "Nantes", DiscoveryURL: srv.URL + "/gbfs.json", PreferredLang: "fr"}, srv.Client())

	stations, err := c.Stations(context.Background())
	if err != nil {
		t.Fatalf("Stations: %v", err)
	}
	if got, want := len(stations), 3; got != want {
		t.Fatalf("nombre de stations = %d, attendu %d", got, want)
	}
	if got := c.City(); got != "Nantes" {
		t.Errorf("City() = %q, attendu Nantes", got)
	}

	pref := indexStations(stations)["1"]
	if pref.Name != "PRÉFECTURE" {
		t.Errorf("station 1 name = %q, attendu PRÉFECTURE", pref.Name)
	}
	if pref.Lat != 47.21984 || pref.Lon != -1.554891 {
		t.Errorf("station 1 coords = (%v, %v)", pref.Lat, pref.Lon)
	}
	if pref.Capacity != 33 {
		t.Errorf("station 1 capacity = %d, attendu 33", pref.Capacity)
	}
	if pref.City != "Nantes" {
		t.Errorf("station 1 city = %q, attendu Nantes", pref.City)
	}
}

func TestClientStatuses(t *testing.T) {
	srv, _ := fixtureServer(t, "nantes")
	c := New(Config{City: "Nantes", DiscoveryURL: srv.URL + "/gbfs.json", PreferredLang: "fr"}, srv.Client())

	statuses, err := c.Statuses(context.Background())
	if err != nil {
		t.Fatalf("Statuses: %v", err)
	}
	byID := indexStatuses(statuses)

	if s := byID["1"]; !s.IsRenting || !s.IsInstalled || !s.IsReturning {
		t.Errorf("station 1 flags = %+v, attendu tous vrais", s)
	}
	if s := byID["1"]; s.BikesAvailable != 26 || s.DocksAvailable != 7 {
		t.Errorf("station 1 dispo = %d vélos / %d bornes", s.BikesAvailable, s.DocksAvailable)
	}
	if s := byID["1"]; !s.LastReported.Equal(time.Unix(1784510482, 0).UTC()) {
		t.Errorf("station 1 last_reported = %v", s.LastReported)
	}
	if s := byID["12"]; s.IsRenting {
		t.Errorf("station 12 IsRenting = true, attendu false")
	}
}

func TestClientVersionAwareV1(t *testing.T) {
	srv, _ := fixtureServer(t, "v1")
	c := New(Config{City: "Paris", DiscoveryURL: srv.URL + "/gbfs.json"}, srv.Client())

	statuses, err := c.Statuses(context.Background())
	if err != nil {
		t.Fatalf("Statuses v1: %v", err)
	}
	s42 := indexStatuses(statuses)["42"]
	if !s42.IsInstalled || !s42.IsRenting {
		t.Errorf("station 42 : is_installed/is_renting=1 devaient donner true, got %+v", s42)
	}
	if s42.IsReturning {
		t.Errorf("station 42 : is_returning=0 devait donner false")
	}

	stations, err := c.Stations(context.Background())
	if err != nil {
		t.Fatalf("Stations v1: %v", err)
	}
	s99, ok := indexStations(stations)["99"]
	if !ok {
		t.Fatal("station 99 absente")
	}
	if s99.Capacity != 0 {
		t.Errorf("station 99 sans capacity devait donner 0, got %d", s99.Capacity)
	}
	if s99.City != "Paris" {
		t.Errorf("station 99 city = %q, attendu Paris", s99.City)
	}
}

func TestDiscoveryCached(t *testing.T) {
	srv, hits := fixtureServer(t, "nantes")
	c := New(Config{City: "Nantes", DiscoveryURL: srv.URL + "/gbfs.json", PreferredLang: "fr"}, srv.Client())

	if _, err := c.Stations(context.Background()); err != nil {
		t.Fatalf("Stations: %v", err)
	}
	if _, err := c.Statuses(context.Background()); err != nil {
		t.Fatalf("Statuses: %v", err)
	}
	if got := atomic.LoadInt64(hits); got != 1 {
		t.Errorf("gbfs.json récupéré %d fois, attendu 1 (mise en cache)", got)
	}
}

func TestMissingFeed(t *testing.T) {
	mux := http.NewServeMux()
	var srv *httptest.Server
	mux.HandleFunc("/gbfs.json", func(w http.ResponseWriter, r *http.Request) {
		disco := map[string]any{
			"version": "2.3",
			"data": map[string]any{
				"fr": map[string]any{
					"feeds": []map[string]string{
						{"name": "station_information", "url": srv.URL + "/station_information.json"},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(disco)
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := New(Config{City: "X", DiscoveryURL: srv.URL + "/gbfs.json", PreferredLang: "fr"}, srv.Client())
	if _, err := c.Statuses(context.Background()); err == nil {
		t.Fatal("Statuses aurait dû échouer : flux station_status absent")
	}
}

func TestContextCancelled(t *testing.T) {
	srv, _ := fixtureServer(t, "nantes")
	c := New(Config{City: "Nantes", DiscoveryURL: srv.URL + "/gbfs.json", PreferredLang: "fr"}, srv.Client())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.Stations(ctx); err == nil {
		t.Fatal("Stations aurait dû échouer avec un context annulé")
	}
}

func TestFlexBool(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    bool
		wantErr bool
	}{
		{"bool true (2.x)", `true`, true, false},
		{"bool false (2.x)", `false`, false, false},
		{"int 1 (1.0)", `1`, true, false},
		{"int 0 (1.0)", `0`, false, false},
		{"null", `null`, false, false},
		{"entier inattendu", `2`, false, true},
		{"chaîne", `"true"`, false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b flexBool
			err := json.Unmarshal([]byte(tc.in), &b)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("attendu une erreur pour %q", tc.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("erreur inattendue pour %q: %v", tc.in, err)
			}
			if bool(b) != tc.want {
				t.Errorf("flexBool(%q) = %v, attendu %v", tc.in, bool(b), tc.want)
			}
		})
	}
}

func TestChooseLang(t *testing.T) {
	set := map[string]feedSet{"en": {}, "fr": {}}
	tests := []struct {
		name      string
		preferred string
		want      string
	}{
		{"préférence honorée", "fr", "fr"},
		{"préférence absente -> fallback déterministe", "de", "en"},
		{"pas de préférence -> fallback déterministe", "", "en"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := chooseLang(set, tc.preferred)
			if !ok || got != tc.want {
				t.Errorf("chooseLang(%q) = %q,%v ; attendu %q", tc.preferred, got, ok, tc.want)
			}
		})
	}
}
