package gbfs

import (
	"encoding/json"
	"fmt"
)

// feedIndex associe un nom de flux GBFS (station_information, station_status, ...)
// à son URL, résolue à partir du fichier de découverte.
type feedIndex map[string]string

// parseDiscovery lit le corps d'un gbfs.json et construit l'index des flux pour
// la langue demandée. Structure attendue (commune GBFS 1.x / 2.x) :
//
//	{ "data": { "<lang>": { "feeds": [ { "name": ..., "url": ... } ] } } }
//
// preferredLang est privilégiée si présente ; sinon la première langue publiée
// est retenue (ordre déterministe : on garde la langue la plus courte puis par
// ordre alphabétique, pour rester stable d'un appel à l'autre).
func parseDiscovery(raw json.RawMessage, preferredLang string) (feedIndex, error) {
	var byLang map[string]feedSet
	if err := json.Unmarshal(raw, &byLang); err != nil {
		return nil, fmt.Errorf("découverte gbfs: data illisible: %w", err)
	}
	if len(byLang) == 0 {
		return nil, fmt.Errorf("découverte gbfs: aucune langue dans data")
	}

	lang, ok := chooseLang(byLang, preferredLang)
	if !ok {
		return nil, fmt.Errorf("découverte gbfs: langue %q absente et aucune alternative", preferredLang)
	}

	idx := make(feedIndex, len(byLang[lang].Feeds))
	for _, f := range byLang[lang].Feeds {
		if f.Name == "" || f.URL == "" {
			continue
		}
		idx[f.Name] = f.URL
	}
	if len(idx) == 0 {
		return nil, fmt.Errorf("découverte gbfs: aucun flux exploitable pour la langue %q", lang)
	}
	return idx, nil
}

// chooseLang sélectionne la langue à utiliser : preferredLang si disponible,
// sinon un choix déterministe parmi les langues publiées.
func chooseLang(byLang map[string]feedSet, preferredLang string) (string, bool) {
	if preferredLang != "" {
		if _, ok := byLang[preferredLang]; ok {
			return preferredLang, true
		}
	}
	var best string
	for lang := range byLang {
		if best == "" || lang < best {
			best = lang
		}
	}
	return best, best != ""
}

// url retourne l'URL d'un flux nommé, ou une erreur explicite s'il est absent de
// la découverte.
func (idx feedIndex) url(name string) (string, error) {
	u, ok := idx[name]
	if !ok {
		return "", fmt.Errorf("flux %q absent de la découverte gbfs", name)
	}
	return u, nil
}
