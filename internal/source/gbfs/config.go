package gbfs

// Config décrit une ville exposée au format GBFS. Ajouter une ville GBFS = créer
// une nouvelle Config ; on ne modifie jamais l'adaptateur ni le cœur.
type Config struct {
	// City est le nom de la ville (repris tel quel dans domain.Station.City).
	City string
	// DiscoveryURL est l'URL du fichier de découverte gbfs.json. Elle doit être
	// récupérée sur la page officielle de la source : on ne suit ensuite QUE la
	// découverte, sans jamais deviner les URLs des sous-flux.
	DiscoveryURL string
	// PreferredLang est le code langue à privilégier dans la découverte
	// (data.<lang>.feeds). Si vide ou absent du flux, la première langue publiée
	// est utilisée.
	PreferredLang string
}
