package registry

import (
	"fmt"
	"net/http"

	"github.com/katchalamele/velodispo/internal/source"
	"github.com/katchalamele/velodispo/internal/source/gbfs"
)

var knownGBFS = map[string]gbfs.Config{
	"nantes": {
		City:          "Nantes",
		DiscoveryURL:  "https://api.cyclocity.fr/contracts/nantes/gbfs/gbfs.json",
		PreferredLang: "fr",
	},
	"paris": {
		City:          "Paris",
		DiscoveryURL:  "https://velib-metropole-opendata.smovengo.cloud/opendata/Velib_Metropole/gbfs.json",
		PreferredLang: "en",
	},
}

func Build(enabled []string, httpClient *http.Client) ([]source.Source, error) {
	sources := make([]source.Source, 0, len(enabled))
	for _, key := range enabled {
		cfg, ok := knownGBFS[key]
		if !ok {
			return nil, fmt.Errorf("source inconnue: %q", key)
		}
		sources = append(sources, gbfs.New(cfg, httpClient))
	}
	return sources, nil
}
