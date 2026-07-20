package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/katchalamele/velodispo/internal/config"
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

	log.Println("prêt")
	<-ctx.Done()
	log.Println("arrêt")
}
