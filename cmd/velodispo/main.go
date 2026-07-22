// @title       Vélo Aggregator API
// @version     1.0
// @description Disponibilité temps réel de vélos en libre-service, multi-villes.
// @BasePath    /
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/katchalamele/velodispo/internal/api"
	"github.com/katchalamele/velodispo/internal/config"
	"github.com/katchalamele/velodispo/internal/ingest"
	"github.com/katchalamele/velodispo/internal/source/registry"
	"github.com/katchalamele/velodispo/internal/store"
)

func main() {
	migrateOnly := flag.Bool("migrate", false, "applique les migrations puis quitte")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := store.Migrate(cfg.DB.DSN()); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations appliquées")

	if *migrateOnly {
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st, err := store.Open(ctx, cfg.DB.DSN())
	if err != nil {
		log.Fatalf("connexion base: %v", err)
	}
	defer st.Close()

	httpClient := &http.Client{Timeout: cfg.Poll.HTTPTimeout}
	sources, err := registry.Build(cfg.Poll.Sources, httpClient)
	if err != nil {
		log.Fatalf("sources: %v", err)
	}
	poller := ingest.New(sources, st, cfg.Poll.Interval, cfg.Poll.HTTPTimeout, cfg.Poll.MaxConcurrency)
	go poller.Run(ctx)
	log.Printf("poller démarré (%s, sources: %v)", cfg.Poll.Interval, cfg.Poll.Sources)

	e := api.New(st)
	go func() {
		if err := e.Start(cfg.HTTPAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serveur http: %v", err)
		}
	}()
	log.Printf("API à l'écoute sur %s", cfg.HTTPAddr)

	<-ctx.Done()
	log.Println("arrêt en cours")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("arrêt serveur: %v", err)
	}
}
