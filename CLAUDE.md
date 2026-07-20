# Vélo Aggregator — Contexte projet

## Ce que c'est
Service Go qui agrège la disponibilité temps réel de vélos en libre-service de
plusieurs villes (Nantes/Bicloo, Paris/Vélib', extensible), normalise des flux
hétérogènes dans un modèle unifié, historise la disponibilité, et expose une API
REST + une carte Leaflet.
Projet portfolio : privilégier le Go idiomatique et l'outillage de l'écosystème
plutôt que les raccourcis. Ne pas se contenter d'un binaire stdlib.

## Stack (décidée — ne pas re-débattre en cours de route)
- Langage : Go 1.22+
- Framework web : Echo
- BDD : PostgreSQL + extension TimescaleDB (hypertable pour les snapshots de dispo)
- Accès BDD : GORM pour le CRUD des entités ; SQL brut via pgx pour les requêtes
  time-series / analytiques (savoir quand NE PAS utiliser l'ORM)
- Migrations : golang-migrate
- Config : variables d'environnement (envconfig)
- Orchestration : docker-compose (app + postgres/timescale + adminer)
- Front : Leaflet + OpenStreetMap, vanilla JS

## Principe d'architecture central
Les sources sont hétérogènes : Nantes publie du GBFS 2.3, Paris du GBFS 1.0, et
JCDecaux expose en plus une API propriétaire. Toute cette variabilité est isolée
derrière une interface Source.
- internal/domain  : modèle unifié (Station, Status) — aucun champ spécifique à une source
- internal/source  : interface Source { City(); Stations(ctx); Statuses(ctx) }
- internal/source/gbfs     : adaptateur GBFS générique, conscient des versions (1.x et 2.x)
- internal/source/jcdecaux : adaptateur propriétaire (optionnel)
Ajouter une ville = ajouter une entrée de config (si GBFS) ou un nouvel adaptateur.
On ne touche JAMAIS au cœur pour ajouter une source.

## Conventions
- Tout appel externe prend un context.Context avec timeout.
- Concurrence via errgroup / worker pool borné ; pas de goroutine non bornée.
- Tests table-driven ; capturer de vrais échantillons de flux comme fixtures testdata.
- Le modèle de domaine reste libre de tout nommage propre à un fournisseur.

## Commandes
- docker compose up        # démarre tout
- make test
- make migrate
