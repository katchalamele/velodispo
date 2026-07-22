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

const (
	feedStationInformation = "station_information"
	feedStationStatus      = "station_status"
)

type Client struct {
	cfg  Config
	http *http.Client

	discoOnce sync.Once
	discoIdx  feedIndex
	discoErr  error
}

var _ source.Source = (*Client)(nil)

func New(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{cfg: cfg, http: httpClient}
}

func (c *Client) City() string { return c.cfg.City }

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

func (c *Client) fetchJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("construction requête: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("appel http: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
		return fmt.Errorf("statut http %d: %s", resp.StatusCode, snippet)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("décodage json: %w", err)
	}
	return nil
}
