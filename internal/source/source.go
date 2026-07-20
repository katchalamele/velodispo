// Package source définit le contrat commun à toutes les sources de données.
//
// Les flux amont sont hétérogènes (GBFS 1.x, GBFS 2.x, API propriétaires). Toute
// cette variabilité est isolée derrière l'interface Source : le cœur de
// l'application ne manipule que du domain.Station / domain.Status. Ajouter une
// ville revient à fournir une nouvelle implémentation (ou une nouvelle config
// pour un adaptateur existant) sans jamais toucher au cœur.
package source

import (
	"context"

	"github.com/katchalamele/velodispo/internal/domain"
)

// Source expose la disponibilité d'un service de vélos en libre-service pour une
// ville. Toutes les méthodes qui déclenchent un appel externe prennent un
// context.Context afin de porter timeout et annulation.
type Source interface {
	// City retourne le nom de la ville couverte par la source.
	City() string
	// Stations retourne les informations statiques des stations.
	Stations(ctx context.Context) ([]domain.Station, error)
	// Statuses retourne la disponibilité temps réel des stations.
	Statuses(ctx context.Context) ([]domain.Status, error)
}
