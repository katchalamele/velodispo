package domain

import "time"

// Status décrit la disponibilité d'une station à un instant donné. C'est cet état
// qui sera historisé (hypertable Timescale) dans les étapes ultérieures.
type Status struct {
	// StationID référence Station.ID au sein d'une même ville.
	StationID string
	// BikesAvailable est le nombre de vélos disponibles à la location.
	BikesAvailable int
	// DocksAvailable est le nombre de bornes libres pour reposer un vélo.
	DocksAvailable int
	// BikesDisabled et DocksDisabled comptent les vélos / bornes hors service.
	BikesDisabled int
	DocksDisabled int
	// IsInstalled indique que la station est déployée et opérationnelle.
	IsInstalled bool
	// IsRenting indique que la station autorise la prise de vélos.
	IsRenting bool
	// IsReturning indique que la station accepte le retour de vélos.
	IsReturning bool
	// LastReported est l'horodatage du dernier relevé remonté par la station.
	LastReported time.Time
}
