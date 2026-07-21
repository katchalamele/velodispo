package registry

import "testing"

func TestBuildKnownSources(t *testing.T) {
	sources, err := Build([]string{"nantes", "paris"}, nil)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(sources) != 2 {
		t.Fatalf("sources = %d, attendu 2", len(sources))
	}
	if sources[0].City() != "Nantes" || sources[1].City() != "Paris" {
		t.Errorf("villes = %q/%q, attendu Nantes/Paris", sources[0].City(), sources[1].City())
	}
}

func TestBuildUnknownSource(t *testing.T) {
	if _, err := Build([]string{"tokyo"}, nil); err == nil {
		t.Fatal("Build devait échouer sur une source inconnue")
	}
}
