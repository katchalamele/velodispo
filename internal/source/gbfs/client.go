package gbfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/katchalamele/velodispo/internal/domain"
	"github.com/katchalamele/velodispo/internal/source"
)

// Noms des flux GBFS suivis par l'adaptateur.
const (
	feedStationInformation = "station_information"
	feedStationStatus      = "station_status"
)

// Client est un adaptateur GBFS générique et version-aware (1.x et 2.x). Il part
// du fichier de découverte défini en Config, suit la découverte vers les
// sous-flux, et normalise le tout vers internal/domain. Il implémente
// source.Source.
type Client struct {
	cfg  Config
	http *http.Client

	// La découverte est résolue paresseusement puis mise en cache : le gbfs.json
	// n'est lu qu'une fois, pas à chaque appel de Stations/Statuses.
	discoOnce sync.Once
	discoIdx  feedIndex
	discoErr  error
}

// Vérification à la compilation que Client satisfait bien l'interface.
var _ source.Source = (*Client)(nil)

// New construit un Client. Si httpClient est nil, http.DefaultClient est utilisé.
// En production, fournir un *http.Client avec un timeout ; les appels portent de
// toute façon un context.Context.
func New(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{cfg: cfg, http: httpClient}
}

// City retourne le nom de la ville couverte.
func (c *Client) City() string { return c.cfg.City }

// Stations résout la découverte puis retourne les informations statiques des
// stations, normalisées.
func (c *Client) Stations(ctx context.Context) ([]domain.Station, error) {
	idx, err := c.discovery(ctx)
	if err != nil {
		return nil, err
	}
	u, err := idx.url(feedStationInformation)
	if err != nil {
		return nil, err
	}

	var file gbfsFile
	if err := c.fetchJSON(ctx, u, &file); err != nil {
		return nil, fmt.Errorf("%s: %w", feedStationInformation, err)
	}
	var data stationInformationData
	if err := json.Unmarshal(file.Data, &data); err != nil {
		return nil, fmt.Errorf("%s: data illisible: %w", feedStationInformation, err)
	}

	stations := make([]domain.Station, 0, len(data.Stations))
	for _, s := range data.Stations {
		stations = append(stations, toStation(c.cfg.City, s))
	}
	return stations, nil
}

// Statuses résout la découverte puis retourne la disponibilité temps réel des
// stations, normalisée.
func (c *Client) Statuses(ctx context.Context) ([]domain.Status, error) {
	idx, err := c.discovery(ctx)
	if err != nil {
		return nil, err
	}
	u, err := idx.url(feedStationStatus)
	if err != nil {
		return nil, err
	}

	var file gbfsFile
	if err := c.fetchJSON(ctx, u, &file); err != nil {
		return nil, fmt.Errorf("%s: %w", feedStationStatus, err)
	}
	var data stationStatusData
	if err := json.Unmarshal(file.Data, &data); err != nil {
		return nil, fmt.Errorf("%s: data illisible: %w", feedStationStatus, err)
	}

	statuses := make([]domain.Status, 0, len(data.Stations))
	for _, s := range data.Stations {
		statuses = append(statuses, toStatus(s))
	}
	return statuses, nil
}

// discovery résout (une seule fois) l'index des flux à partir du gbfs.json.
//
// Note : sync.Once fige aussi la première erreur. C'est acceptable ici — un
// gbfs.json injoignable est une mauvaise config plutôt qu'un aléa réseau ; les
// étapes ultérieures (poller) recréeront un Client si besoin. On garde ainsi la
// simplicité sans re-télécharger la découverte à chaque tick.
func (c *Client) discovery(ctx context.Context) (feedIndex, error) {
	c.discoOnce.Do(func() {
		var file gbfsFile
		if err := c.fetchJSON(ctx, c.cfg.DiscoveryURL, &file); err != nil {
			c.discoErr = fmt.Errorf("découverte gbfs (%s): %w", c.cfg.DiscoveryURL, err)
			return
		}
		c.discoIdx, c.discoErr = parseDiscovery(file.Data, c.cfg.PreferredLang)
	})
	return c.discoIdx, c.discoErr
}

// fetchJSON exécute un GET porté par ctx et décode le corps JSON dans out.
func (c *Client) fetchJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("construction requête: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("appel http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// On draine un extrait du corps pour un message d'erreur exploitable.
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("statut http %d: %s", resp.StatusCode, snippet)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("décodage json: %w", err)
	}
	return nil
}
