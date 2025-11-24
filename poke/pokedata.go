package poke

import (
	"encoding/csv"
	"os"
	"strconv"
	"strings"
)

// LoadPokemonFromCSV loads Pokemon data from the CSV file.
// Returns a map of Pokemon name to Pokemon struct.
func LoadPokemonFromCSV(filepath string) (map[string]Pokemon, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	pokemons := make(map[string]Pokemon)

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]

		// Parse stats from CSV columns
		// attack (19), defense (25), hp (28), sp_attack (33), sp_defense (34),
		// speed (35), type1 (36), type2 (37), name (30)

		attack, _ := strconv.Atoi(strings.TrimSpace(record[19]))
		defense, _ := strconv.Atoi(strings.TrimSpace(record[25]))
		hp, _ := strconv.Atoi(strings.TrimSpace(record[28]))
		spAttack, _ := strconv.Atoi(strings.TrimSpace(record[33]))
		spDefense, _ := strconv.Atoi(strings.TrimSpace(record[34]))
		speed, _ := strconv.Atoi(strings.TrimSpace(record[35]))

		name := strings.TrimSpace(record[30])
		type1 := strings.ToLower(strings.TrimSpace(record[36]))
		type2 := strings.ToLower(strings.TrimSpace(record[37]))

		// Create basic moves for this Pokemon based on its type
		moves := createDefaultMoves(type1, type2)

		pokemon := Pokemon{
			Name:           name,
			HP:             hp,
			MaxHP:          hp,
			Attack:         attack,
			Defense:        defense,
			SpecialAttack:  spAttack,
			SpecialDefense: spDefense,
			Speed:          speed,
			Type1:          type1,
			Type2:          type2,
			Moves:          moves,
		}

		pokemons[name] = pokemon
	}

	return pokemons, nil
}

// createDefaultMoves creates a set of default moves for a Pokemon based on its types.
func createDefaultMoves(type1, type2 string) []Move {
	moves := []Move{
		{Name: "Tackle", BasePower: 40, Type: "normal", DamageCategory: Physical},
	}

	// Add a STAB (Same Type Attack Bonus) move for primary type
	if type1 != "" {
		moves = append(moves, Move{
			Name:           strings.Title(type1) + " Attack",
			BasePower:      60,
			Type:           type1,
			DamageCategory: getTypicalDamageCategory(type1),
		})
	}

	// Add a move for secondary type if it exists
	if type2 != "" {
		moves = append(moves, Move{
			Name:           strings.Title(type2) + " Attack",
			BasePower:      60,
			Type:           type2,
			DamageCategory: getTypicalDamageCategory(type2),
		})
	}

	// Add a special move
	moves = append(moves, Move{
		Name:           "Special Blast",
		BasePower:      70,
		Type:           type1,
		DamageCategory: Special,
	})

	return moves
}

// getTypicalDamageCategory returns the typical damage category for a type.
func getTypicalDamageCategory(pokeType string) string {
	// Special types (typically use special attack/defense)
	specialTypes := map[string]bool{
		"fire": true, "water": true, "grass": true, "electric": true,
		"ice": true, "psychic": true, "dragon": true, "dark": true, "fairy": true,
	}

	if specialTypes[pokeType] {
		return Special
	}
	return Physical
}
