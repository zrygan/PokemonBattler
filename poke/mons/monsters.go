package monsters

import (
	"log"
	"path/filepath"

	"github.com/zrygan/pokemonbattler/poke"
)

// MONSTERS contains all loaded Pokemon data from the CSV file.
var MONSTERS map[string]poke.Pokemon

// init loads the Pokemon data when the package is imported.
func init() {
	// Try to load from the data directory
	csvPath := filepath.Join("data", "pokemon.csv")

	var err error
	MONSTERS, err = poke.LoadPokemonFromCSV(csvPath)
	if err != nil {
		log.Printf("Warning: Failed to load Pokemon data from %s: %v", csvPath, err)
		// Initialize with empty map as fallback
		MONSTERS = make(map[string]poke.Pokemon)
	} else {
		log.Printf("Successfully loaded %d Pokemon from CSV", len(MONSTERS))
	}
}
