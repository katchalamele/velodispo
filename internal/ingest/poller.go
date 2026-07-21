package ingest

import (
	"context"
	"log"
	"time"

	"github.com/katchalamele/velodispo/internal/domain"
	"github.com/katchalamele/velodispo/internal/source"
	"github.com/katchalamele/velodispo/internal/store"
	"golang.org/x/sync/errgroup"
)

type Ingester interface {
	UpsertCity(ctx context.Context, name string) (store.City, error)
	UpsertStations(ctx context.Context, cityID int64, stations []domain.Station) error
	StationsByCity(ctx context.Context, cityID int64) ([]store.Station, error)
	InsertSnapshots(ctx context.Context, snaps []store.Snapshot) (int64, error)
}

type Poller struct {
	sources    []source.Source
	store      Ingester
	interval   time.Duration
	srcTimeout time.Duration
	maxConc    int
	now        func() time.Time
}

func New(sources []source.Source, st Ingester, interval, srcTimeout time.Duration, maxConc int) *Poller {
	if maxConc < 1 {
		maxConc = 1
	}
	return &Poller{
		sources:    sources,
		store:      st,
		interval:   interval,
		srcTimeout: srcTimeout,
		maxConc:    maxConc,
		now:        time.Now,
	}
}

func (p *Poller) Run(ctx context.Context) {
	p.pollOnce(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollOnce(ctx)
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(p.maxConc)

	for _, src := range p.sources {
		src := src
		g.Go(func() error {
			if err := p.ingestSource(ctx, src); err != nil {
				log.Printf("ingestion %s: %v", src.City(), err)
			}
			return nil
		})
	}
	_ = g.Wait()
}

func (p *Poller) ingestSource(ctx context.Context, src source.Source) error {
	ctx, cancel := context.WithTimeout(ctx, p.srcTimeout)
	defer cancel()

	city, err := p.store.UpsertCity(ctx, src.City())
	if err != nil {
		return err
	}

	stations, err := src.Stations(ctx)
	if err != nil {
		return err
	}
	if err := p.store.UpsertStations(ctx, city.ID, stations); err != nil {
		return err
	}

	statuses, err := src.Statuses(ctx)
	if err != nil {
		return err
	}

	models, err := p.store.StationsByCity(ctx, city.ID)
	if err != nil {
		return err
	}
	pkByStationID := make(map[string]int64, len(models))
	for _, m := range models {
		pkByStationID[m.StationID] = m.ID
	}

	at := p.now().UTC()
	snaps := make([]store.Snapshot, 0, len(statuses))
	for _, st := range statuses {
		pk, ok := pkByStationID[st.StationID]
		if !ok {
			continue
		}
		snaps = append(snaps, store.SnapshotFrom(pk, at, st))
	}

	n, err := p.store.InsertSnapshots(ctx, snaps)
	if err != nil {
		return err
	}
	log.Printf("ingestion %s: %d stations, %d snapshots", src.City(), len(stations), n)
	return nil
}
