package gbfs

import (
	"encoding/json"
	"fmt"
)

type feedIndex map[string]string

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

func (idx feedIndex) url(name string) (string, error) {
	u, ok := idx[name]
	if !ok {
		return "", fmt.Errorf("flux %q absent de la découverte gbfs", name)
	}
	return u, nil
}
