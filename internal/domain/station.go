// Package domain contient le modèle unifié de l'agrégateur. Il ne doit contenir
// aucun champ ni nommage propre à une source (GBFS, JCDecaux, ...) : toute cette
// variabilité est absorbée par les adaptateurs de internal/source.
package domain

// Station décrit une station de vélos en libre-service, indépendamment de la
// source qui l'a fournie. Les informations statiques (position, capacité) sont
// séparées de la disponibilité temps réel, portée par Status.
type Station struct {
	// ID est l'identifiant de la station tel que fourni par la source. Il n'est
	// unique qu'au sein d'une même ville.
	ID string
	// City est la ville d'appartenance (clé de désambiguïsation entre sources).
	City string
	// Name est le libellé lisible de la station.
	Name string
	// Lat et Lon sont les coordonnées WGS84 de la station.
	Lat float64
	Lon float64
	// Address est l'adresse postale, optionnelle selon les sources.
	Address string
	// Capacity est le nombre total de points d'attache. Zéro si non renseigné.
	Capacity int
}
