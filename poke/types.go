package poke

// Pokemon represents a player's pokemon with stats and boosts.
type Pokemon struct {
	Name           string // Pokemon name
	HP             int    // Hit points (current HP)
	MaxHP          int    // Maximum hit points
	Attack         int    // Physical attack stat
	Defense        int    // Physical defense stat
	SpecialAttack  int    // Special attack stat
	SpecialDefense int    // Special defense stat
	Speed          int    // Speed stat (determines turn order)
	Type1          string // Primary type (e.g., "fire", "water", "grass")
	Type2          string // Secondary type (empty string if single type)
	Moves          []Move // List of moves the pokemon can use
}

// Move represents a single move a pokemon can use in battle.
type Move struct {
	Name           string  // Name of the pokemon move
	BasePower      float64 // Base power of the move (default 1.0)
	Type           string  // Type of the move (e.g., "fire", "water")
	DamageCategory string  // "physical" or "special"
}

// DamageCategory constants
const (
	Physical = "physical"
	Special  = "special"
)

// Type effectiveness multipliers
var TypeEffectiveness = map[string]map[string]float64{
	"normal": {
		"rock": 0.5, "ghost": 0.0, "steel": 0.5,
	},
	"fire": {
		"fire": 0.5, "water": 0.5, "grass": 2.0, "ice": 2.0, "bug": 2.0, "rock": 0.5,
		"dragon": 0.5, "steel": 2.0,
	},
	"water": {
		"fire": 2.0, "water": 0.5, "grass": 0.5, "ground": 2.0, "rock": 2.0, "dragon": 0.5,
	},
	"electric": {
		"water": 2.0, "electric": 0.5, "grass": 0.5, "ground": 0.0, "flying": 2.0, "dragon": 0.5,
	},
	"grass": {
		"fire": 0.5, "water": 2.0, "grass": 0.5, "poison": 0.5, "ground": 2.0, "flying": 0.5,
		"bug": 0.5, "rock": 2.0, "dragon": 0.5, "steel": 0.5,
	},
	"ice": {
		"fire": 0.5, "water": 0.5, "grass": 2.0, "ice": 0.5, "ground": 2.0, "flying": 2.0,
		"dragon": 2.0, "steel": 0.5,
	},
	"fighting": {
		"normal": 2.0, "ice": 2.0, "poison": 0.5, "flying": 0.5, "psychic": 0.5, "bug": 0.5,
		"rock": 2.0, "ghost": 0.0, "dark": 2.0, "steel": 2.0, "fairy": 0.5,
	},
	"poison": {
		"grass": 2.0, "poison": 0.5, "ground": 0.5, "rock": 0.5, "ghost": 0.5, "steel": 0.0, "fairy": 2.0,
	},
	"ground": {
		"fire": 2.0, "electric": 2.0, "grass": 0.5, "poison": 2.0, "flying": 0.0, "bug": 0.5,
		"rock": 2.0, "steel": 2.0,
	},
	"flying": {
		"electric": 0.5, "grass": 2.0, "fighting": 2.0, "bug": 2.0, "rock": 0.5, "steel": 0.5,
	},
	"psychic": {
		"fighting": 2.0, "poison": 2.0, "psychic": 0.5, "dark": 0.0, "steel": 0.5,
	},
	"bug": {
		"fire": 0.5, "grass": 2.0, "fighting": 0.5, "poison": 0.5, "flying": 0.5, "psychic": 2.0,
		"ghost": 0.5, "dark": 2.0, "steel": 0.5, "fairy": 0.5,
	},
	"rock": {
		"fire": 2.0, "ice": 2.0, "fighting": 0.5, "ground": 0.5, "flying": 2.0, "bug": 2.0, "steel": 0.5,
	},
	"ghost": {
		"normal": 0.0, "psychic": 2.0, "ghost": 2.0, "dark": 0.5,
	},
	"dragon": {
		"dragon": 2.0, "steel": 0.5, "fairy": 0.0,
	},
	"dark": {
		"fighting": 0.5, "psychic": 2.0, "ghost": 2.0, "dark": 0.5, "fairy": 0.5,
	},
	"steel": {
		"fire": 0.5, "water": 0.5, "electric": 0.5, "ice": 2.0, "rock": 2.0, "steel": 0.5, "fairy": 2.0,
	},
	"fairy": {
		"fire": 0.5, "fighting": 2.0, "poison": 0.5, "dragon": 2.0, "dark": 2.0, "steel": 0.5,
	},
}

// GetTypeEffectiveness returns the type effectiveness multiplier for an attack type against a defending type.
func GetTypeEffectiveness(attackType string, defenseType string) float64 {
	if matchups, ok := TypeEffectiveness[attackType]; ok {
		if multiplier, ok := matchups[defenseType]; ok {
			return multiplier
		}
	}
	return 1.0 // Neutral effectiveness
}
