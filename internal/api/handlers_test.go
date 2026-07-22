package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/katchalamele/velodispo/internal/store"
)

type fakeReader struct {
	views      []store.StationView
	total      int
	lastFilter store.StationFilter
	getErr     error
	getView    store.StationView
	history    []store.Snapshot
	lastFrom   time.Time
	lastTo     time.Time
	profile    []store.SlotAvg
	lastIsodow int
	mapViews   []store.StationView
}

func (f *fakeReader) PredictionProfile(_ context.Context, _ int64, isodow int) ([]store.SlotAvg, error) {
	f.lastIsodow = isodow
	return f.profile, nil
}

func (f *fakeReader) MapStations(_ context.Context) ([]store.StationView, error) {
	return f.mapViews, nil
}

func (f *fakeReader) ListStations(_ context.Context, flt store.StationFilter) ([]store.StationView, int, error) {
	f.lastFilter = flt
	return f.views, f.total, nil
}

func (f *fakeReader) GetStation(_ context.Context, _ int64) (store.StationView, error) {
	return f.getView, f.getErr
}

func (f *fakeReader) History(_ context.Context, _ int64, from, to time.Time) ([]store.Snapshot, error) {
	f.lastFrom, f.lastTo = from, to
	return f.history, nil
}

func do(t *testing.T, reader StationReader, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	e := New(reader)
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestListStationsEmpty(t *testing.T) {
	f := &fakeReader{views: nil, total: 0}
	rec := do(t, f, http.MethodGet, "/stations")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}

	var resp ListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Data == nil {
		t.Error("data devrait être [] et non null")
	}
	if resp.Pagination.Limit != defaultLimit || resp.Pagination.Total != 0 {
		t.Errorf("pagination inattendue: %+v", resp.Pagination)
	}
}

func TestListStationsFilterAndPagination(t *testing.T) {
	f := &fakeReader{total: 3}
	rec := do(t, f, http.MethodGet, "/stations?city=Nantes&limit=9999&offset=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	if f.lastFilter.City != "Nantes" {
		t.Errorf("filtre ville = %q", f.lastFilter.City)
	}
	if f.lastFilter.Limit != maxLimit {
		t.Errorf("limit = %d, attendu borné à %d", f.lastFilter.Limit, maxLimit)
	}
	if f.lastFilter.Offset != 10 {
		t.Errorf("offset = %d, attendu 10", f.lastFilter.Offset)
	}
}

func TestListStationsBadLimit(t *testing.T) {
	rec := do(t, &fakeReader{}, http.MethodGet, "/stations?limit=abc")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, attendu 400", rec.Code)
	}
}

func TestGetStationFound(t *testing.T) {
	now := time.Unix(1784561205, 0).UTC()
	f := &fakeReader{getView: store.StationView{
		Station:  store.Station{ID: 1, StationID: "1", Name: "PRÉFECTURE", Capacity: 33},
		CityName: "Nantes",
		Status:   &store.Snapshot{StationPK: 1, Time: now, BikesAvailable: 26, DocksAvailable: 7, IsRenting: true},
	}}
	rec := do(t, f, http.MethodGet, "/stations/1")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}

	var resp StationResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.City != "Nantes" || resp.StationID != "1" {
		t.Errorf("réponse inattendue: %+v", resp)
	}
	if resp.Status == nil || resp.Status.BikesAvailable != 26 {
		t.Errorf("statut absent ou incorrect: %+v", resp.Status)
	}
}

func TestGetStationNotFound(t *testing.T) {
	f := &fakeReader{getErr: store.ErrNotFound}
	rec := do(t, f, http.MethodGet, "/stations/999")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, attendu 404", rec.Code)
	}
}

func TestGetStationBadID(t *testing.T) {
	rec := do(t, &fakeReader{}, http.MethodGet, "/stations/abc")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, attendu 400", rec.Code)
	}
}

func TestHistoryDefaultInterval(t *testing.T) {
	f := &fakeReader{}
	rec := do(t, f, http.MethodGet, "/stations/1/history")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	span := f.lastTo.Sub(f.lastFrom)
	if span < 23*time.Hour || span > 25*time.Hour {
		t.Errorf("plage par défaut = %v, attendu ~24h", span)
	}
}

func TestHistoryBadFrom(t *testing.T) {
	rec := do(t, &fakeReader{}, http.MethodGet, "/stations/1/history?from=notadate")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, attendu 400", rec.Code)
	}
}

func TestHistoryNotFound(t *testing.T) {
	f := &fakeReader{getErr: store.ErrNotFound}
	rec := do(t, f, http.MethodGet, "/stations/999/history")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, attendu 404", rec.Code)
	}
}

func TestPredictionDenseProfile(t *testing.T) {
	f := &fakeReader{profile: []store.SlotAvg{{Slot: 10, AvgBikes: 12.4, AvgDocks: 7.6, Samples: 3}}}
	rec := do(t, f, http.MethodGet, "/stations/1/prediction")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	var resp PredictionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp.Profile) != 48 {
		t.Fatalf("profil = %d slots, attendu 48 (dense)", len(resp.Profile))
	}
	if resp.Profile[10].Samples != 3 || resp.Profile[10].AvgBikes != 12.4 {
		t.Errorf("slot 10 non rempli: %+v", resp.Profile[10])
	}
	if resp.Profile[0].Samples != 0 {
		t.Errorf("slot vide devrait avoir 0 samples: %+v", resp.Profile[0])
	}
	if resp.Weekday < 1 || resp.Weekday > 7 {
		t.Errorf("weekday hors bornes: %d", resp.Weekday)
	}
}

func TestPredictionBadAt(t *testing.T) {
	rec := do(t, &fakeReader{}, http.MethodGet, "/stations/1/prediction?at=notadate")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, attendu 400", rec.Code)
	}
}

func TestPredictionNotFound(t *testing.T) {
	f := &fakeReader{getErr: store.ErrNotFound}
	rec := do(t, f, http.MethodGet, "/stations/999/prediction")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, attendu 404", rec.Code)
	}
}

func TestMap(t *testing.T) {
	f := &fakeReader{mapViews: []store.StationView{
		{Station: store.Station{ID: 1, Name: "A", Lat: 47.2, Lon: -1.5, Capacity: 20}, CityName: "Nantes",
			Status: &store.Snapshot{BikesAvailable: 5, DocksAvailable: 15}},
		{Station: store.Station{ID: 2, Name: "B", Lat: 48.8, Lon: 2.3}, CityName: "Paris"},
	}}
	rec := do(t, f, http.MethodGet, "/map")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200", rec.Code)
	}
	var resp []MapStationResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("stations = %d, attendu 2", len(resp))
	}
	if resp[0].BikesAvailable == nil || *resp[0].BikesAvailable != 5 {
		t.Errorf("station 1 bikes attendu 5: %+v", resp[0].BikesAvailable)
	}
	if resp[1].BikesAvailable != nil {
		t.Errorf("station 2 sans statut -> bikes nil, got %+v", resp[1].BikesAvailable)
	}
}
