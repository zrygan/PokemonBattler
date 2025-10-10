package poke

// Pokemon represents a player's pokemon with stats and boosts.
type Pokemon struct {
	Name      string        // Pokemon name
	HP        int           // Hit points
	Moves     []MovesStruct // List of moves the pokemon can use
	SpAttack  int           // Special attack modification
	SpDefense int           // Special defense modification
}

// MovesStruct represents a single move a pokemon can use in battle.
type MovesStruct struct {
	Name   string // Name of the pokemon move
	Damage string // Damage the move does
}
