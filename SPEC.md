# SPEC — Vélo Aggregator

Plan de construction incrémental. Chaque étape = un passage en **plan mode**
Claude Code : tu fais valider le plan, puis tu exécutes. Vise 1→4 pour avoir
quelque chose de vivant, 5 est le différenciateur, 6 rend le tout présentable.

## Sources de données réelles
Récupérer l'URL exacte du fichier de découverte `gbfs.json` sur les pages officielles
(elles changent, ne pas coder en dur une URL supposée) :

- **Nantes / Bicloo (Naolib)** — GBFS **2.3**, opéré par JCDecaux.
  Page open data : https://data.nantesmetropole.fr/explore/dataset/244400404_disponibilite-temps-reel-velos-libre-service-naolib-gbfs/
- **Paris / Vélib' Métropole** — GBFS **1.0**, opéré par Smovengo, accès libre sans clé.
  Page open data : https://www.velib-metropole.fr/donnees-open-data-gbfs-du-service-velib-metropole
- (Optionnel) **API propriétaire JCDecaux** — nécessite une clé.
  https://developer.jcdecaux.com/

Note : Nantes en 2.3 et Paris en 1.0 = deux versions du standard aux champs
différents. C'est CE point qui justifie l'adaptateur GBFS "version-aware" et qui
donne au projet sa vraie valeur d'ingénierie.

## Décision en attente à confirmer
- Couche BDD : par défaut **GORM** (ORM reconnu, montée en charge rapide). Si tu
  préfères mettre en avant **Ent** (schema-as-code, plus sophistiqué), le signaler
  AVANT l'étape 2, car ça change la modélisation. Le reste de la stack ne bouge pas.

## Étapes

### 1. Adaptateur GBFS générique (une ville)
Client GBFS qui part du `gbfs.json`, suit la découverte vers `station_information`
et `station_status`, et normalise dans internal/domain. Le rendre version-aware
(gérer 1.x ET 2.x). Tests table-driven sur un flux Nantes capturé en fixture.
Livrable : `Source` implémentée pour Nantes, testée hors-ligne.

### 2. Modèle + persistance Postgres/Timescale
Schéma : tables relationnelles (villes, stations) + hypertable Timescale pour les
snapshots de disponibilité. Migrations golang-migrate. GORM pour le CRUD entités.
Livrable : `docker compose up` monte app + Postgres/Timescale, migrations appliquées.

### 3. API REST (Echo)
`GET /stations`, `GET /stations/{id}`, `GET /stations/{id}/history`.
Pagination + filtres. Doc OpenAPI/Swagger.
Livrable : du JSON propre qui sort.

### 4. Poller concurrent en arrière-plan
Goroutine + time.Ticker + context ; errgroup / worker pool borné pour interroger
toutes les sources en parallèle et écrire les snapshots. Gestion des erreurs par
source (une source morte ne tue pas les autres).
Livrable : l'historique se remplit tout seul — la concurrence Go entre en scène.

### 5. Front carte + prédiction  ← LE DIFFÉRENCIATEUR
Carte Leaflet/OSM avec marqueurs live. Au clic sur une station : courbe historique
+ prédiction simple = moyenne glissante par (station, jour de semaine, tranche
horaire). Endpoint dédié pour la prédiction.
Livrable : démontrable seul par un recruteur, sans 2e utilisateur.

### 6. Finition & déploiement
Dockerfile multi-stage, docker-compose complet, pipeline GitLab CI (lint + test +
build), déploiement, README soigné avec GIF de la carte qui se remplit.

## Bonus (si le temps le permet)
- 2e ville active (Paris) pour prouver le multi-sources en conditions réelles.
- Adaptateur JCDecaux propriétaire pour prouver que l'archi encaisse une structure
  totalement différente, pas juste deux variantes de GBFS.
- Recherche géospatiale : stations proches d'un point (PostGIS ou calcul Haversine).
