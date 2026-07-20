package gbfs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Les types de ce fichier décrivent le format GBFS *brut*, tel qu'il arrive sur
// le réseau. Ils restent internes à l'adaptateur : rien de tout cela ne fuit
// vers internal/domain. Le décodage est volontairement tolérant aux différences
// entre GBFS 1.x et 2.x (c'est le cœur de l'aspect « version-aware »).

// gbfsFile est l'enveloppe commune à tous les flux GBFS.
type gbfsFile struct {
	LastUpdated int64           `json:"last_updated"`
	TTL         int             `json:"ttl"`
	Version     string          `json:"version"`
	Data        json.RawMessage `json:"data"`
}

// discoveryData correspond au corps de gbfs.json : une map indexée par code
// langue (1.x et 2.x partagent cette structure ; la 3.0, non gérée ici,
// aplatit les feeds). On la décode donc comme map[lang]feedSet.
type feedSet struct {
	Feeds []feed `json:"feeds"`
}

type feed struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// stationInformationData correspond au corps de station_information.json.
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

// stationStatusData correspond au corps de station_status.json.
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

// flexBool décode un booléen GBFS qu'il soit exprimé en true/false (GBFS 2.x,
// conforme au schéma) ou en 0/1 (certains flux GBFS 1.0 historiques). C'est le
// point précis où la variabilité de version est absorbée.
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	switch {
	case bytes.Equal(data, []byte("true")), bytes.Equal(data, []byte("1")):
		*b = true
		return nil
	case bytes.Equal(data, []byte("false")), bytes.Equal(data, []byte("0")):
		*b = false
		return nil
	case bytes.Equal(data, []byte("null")):
		*b = false
		return nil
	default:
		return fmt.Errorf("flexBool: valeur booléenne inattendue %q", data)
	}
}
