package gbfs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type gbfsFile struct {
	LastUpdated int64           `json:"last_updated"`
	TTL         int             `json:"ttl"`
	Version     string          `json:"version"`
	Data        json.RawMessage `json:"data"`
}

type feedSet struct {
	Feeds []feed `json:"feeds"`
}

type feed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type stationInformationData struct {
	Stations []stationInformation `json:"stations"`
}

type stationInformation struct {
	StationID string  `json:"station_id"`
	Name      string  `json:"name"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Address   string  `json:"address"`
	Capacity  int     `json:"capacity"`
}

type stationStatusData struct {
	Stations []stationStatus `json:"stations"`
}

type stationStatus struct {
	StationID         string   `json:"station_id"`
	NumBikesAvailable int      `json:"num_bikes_available"`
	NumBikesDisabled  int      `json:"num_bikes_disabled"`
	NumDocksAvailable int      `json:"num_docks_available"`
	NumDocksDisabled  int      `json:"num_docks_disabled"`
	IsInstalled       flexBool `json:"is_installed"`
	IsRenting         flexBool `json:"is_renting"`
	IsReturning       flexBool `json:"is_returning"`
	LastReported      int64    `json:"last_reported"`
}

// flexBool tolère true/false (GBFS 2.x) et 0/1 (GBFS 1.0).
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	switch string(bytes.TrimSpace(data)) {
	case "true", "1":
		*b = true
	case "false", "0", "null":
		*b = false
	default:
		return fmt.Errorf("flexBool: valeur inattendue %q", data)
	}
	return nil
}
