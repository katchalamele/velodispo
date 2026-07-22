# Vélo Aggregator

![CI](https://github.com/katchalamele/velodispo/actions/workflows/ci.yml/badge.svg)
![Go](https://img.shields.io/badge/Go-1.25-00ADD8)

Service Go qui agrège la disponibilité **temps réel** des vélos en libre-service
de plusieurs villes (Nantes/Bicloo, Paris/Vélib', extensible), normalise des flux
hétérogènes dans un modèle unifié, historise la disponibilité dans une hypertable
TimescaleDB, et expose une API REST + une carte MapLibre GL avec **prédiction**.

<!-- Capture à ajouter : lancer la stack, ouvrir http://localhost:8081/, cliquer une station,
     enregistrer docs/map.png puis décommenter la ligne ci-dessous.
<p align="center"><img src="docs/map.png" alt="Carte des stations et prédiction" width="800"></p>
-->

## Points d'ingénierie

- **Sources hétérogènes isolées derrière une interface.** Nantes publie du GBFS
  **2.3**, Paris du GBFS **1.0** : champs et types différents (booléens `true/false`
  vs `0/1`, `station_id` chaîne vs nombre). Toute cette variabilité est absorbée par
  un adaptateur GBFS **version-aware** (types `flexBool`/`flexString`) ; le cœur ne
  manipule qu'un modèle de domaine neutre. Ajouter une ville GBFS = une entrée de
  config, jamais toucher au cœur.
- **Time-series.** Les snapshots de disponibilité vont dans une **hypertable
  TimescaleDB**. GORM pour le CRUD des entités (villes/stations) ; **pgx brut** pour
  l'écriture massive (`CopyFrom`) et les requêtes analytiques (savoir quand ne pas
  utiliser l'ORM).
- **Ingestion concurrente.** Un poller en arrière-plan (`time.Ticker` + `context`)
  interroge toutes les sources en parallèle via un **worker pool borné** (errgroup).
  Une source morte est isolée : elle n'affecte ni les autres ni les tics suivants.
- **Prédiction.** Moyenne glissante par `(station, jour de semaine, tranche de
  30 min)` en heure locale Europe/Paris, pour vélos et bornes.
- **Un seul binaire autoportant.** Front (HTML/CSS/JS + MapLibre GL), migrations SQL et
  base des fuseaux horaires sont **embarqués** (`//go:embed`, `time/tzdata`). L'image
  de prod (~40 Mo, distroless) ne contient que l'exécutable.

## Stack

Go 1.25 · Echo · PostgreSQL + TimescaleDB · GORM + pgx · golang-migrate ·
MapLibre GL + OpenFreeMap · Docker Compose · GitHub Actions.

## Architecture

```
cmd/velodispo         point d'entrée unique (poller + API + front)
internal/
  domain              modèle unifié (Station, Status) — neutre
  source              interface Source { City / Stations / Statuses }
    gbfs              adaptateur GBFS version-aware (1.x et 2.x)
    registry          villes connues (config-as-code)
  ingest              poller concurrent (errgroup borné, isolation par source)
  store               Postgres/Timescale : GORM (CRUD) + pgx (time-series) + migrations
  api                 handlers Echo, DTO, OpenAPI/Swagger
  web                 front vanilla JS + MapLibre GL (embarqué)
  config              variables d'environnement (envconfig)
```

## Démarrage (dev)

Aucun toolchain à installer : tout passe par Docker.

```bash
docker compose up          # db (Timescale) + adminer + app
```

- Carte : http://localhost:8081/
- Swagger : http://localhost:8081/swagger/index.html
- Adminer : http://localhost:8080

L'app applique les migrations au démarrage, puis le poller remplit l'historique
tout seul (Nantes + Paris) toutes les 60 s.

## API

| Endpoint | Description |
|---|---|
| `GET /health` | Sonde de vie |
| `GET /stations?city=&limit=&offset=` | Stations paginées + dernière dispo |
| `GET /stations/{id}` | Détail d'une station + dernier statut |
| `GET /stations/{id}/history?from=&to=` | Historique (RFC3339, défaut 24 h) |
| `GET /stations/{id}/prediction?at=` | Profil 48 tranches + prédiction du moment |
| `GET /map` | Toutes les stations (position + dispo) pour la carte |

## Configuration

Variables d'environnement (voir `.env.example`) :

| Variable | Défaut | Rôle |
|---|---|---|
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `db` / `5432` / `velodispo` / … | Connexion Postgres |
| `HTTP_ADDR` | `:8081` | Écoute HTTP |
| `SOURCES` | `nantes,paris` | Villes ingérées |
| `POLL_INTERVAL` | `60s` | Fréquence d'interrogation |

## Tests

```bash
make test     # unitaires (hors-ligne)
make itest    # intégration contre une vraie TimescaleDB
make lint     # golangci-lint
```

## Déploiement (homelab, Docker Compose)

L'image est publiée sur GHCR par la CI (`ghcr.io/katchalamele/velodispo`).
Sur le serveur :

```bash
cp .env.example .env      # renseigner DB_PASSWORD, etc.
docker compose -f compose.prod.yml pull
docker compose -f compose.prod.yml up -d
```

`compose.prod.yml` tire l'image (pas de source ni de toolchain) et lance
l'app + TimescaleDB. Placer un reverse proxy (Traefik/nginx) devant le port 8081
pour le TLS si exposé.

## Roadmap / bonus

- Adaptateur JCDecaux propriétaire (structure non-GBFS).
- Recherche géospatiale (PostGIS / Haversine) : stations proches d'un point.
- Prédiction enrichie (lissage, intervalles de confiance).
