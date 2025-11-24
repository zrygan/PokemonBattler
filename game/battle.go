package game

import (
	"math"
	"math/rand"

	"github.com/zrygan/pokemonbattler/poke"
)

// BattleState represents the current state of the battle.
type BattleState int

const (
	StateSetup BattleState = iota
	StateWaitingForMove
	StateProcessingTurn
	StateGameOver
)

// CalculateDamage calculates the damage dealt by an attack.
// Uses the protocol-specified damage formula with type effectiveness.
func CalculateDamage(
	attacker *poke.Pokemon,
	defender *poke.Pokemon,
	move poke.Move,
	attackerUsesBoost bool, // Whether attacker uses a special attack boost
	defenderUsesBoost bool, // Whether defender uses a special defense boost
	rng *rand.Rand, // Seeded random number generator
) int {
	// Determine which stats to use based on move's damage category
	var attackerStat float64
	var defenderStat float64

	if move.DamageCategory == poke.Physical {
		attackerStat = float64(attacker.Attack)
		defenderStat = float64(defender.Defense)
	} else { // Special
		attackerStat = float64(attacker.SpecialAttack)
		defenderStat = float64(defender.SpecialDefense)

		// Apply stat boosts if used
		if attackerUsesBoost && move.DamageCategory == poke.Special {
			attackerStat *= 1.5 // 50% boost
		}
		if defenderUsesBoost && move.DamageCategory == poke.Special {
			defenderStat *= 1.5 // 50% boost
		}
	}

	// Calculate type effectiveness
	type1Effectiveness := poke.GetTypeEffectiveness(move.Type, defender.Type1)
	type2Effectiveness := 1.0
	if defender.Type2 != "" {
		type2Effectiveness = poke.GetTypeEffectiveness(move.Type, defender.Type2)
	}
	typeEffectiveness := type1Effectiveness * type2Effectiveness

	// Base power (default to 1.0 if not set)
	basePower := move.BasePower
	if basePower == 0 {
		basePower = 1.0
	}

	// Calculate damage using the protocol formula
	// Damage = ((AttackerStat / DefenderStat) * BasePower * TypeEffectiveness) + RandomFactor
	damageFloat := ((attackerStat / defenderStat) * basePower * typeEffectiveness)

	// Add random factor (0-15% variation)
	randomFactor := 0.85 + (rng.Float64() * 0.15)
	damageFloat *= randomFactor

	damage := int(math.Round(damageFloat))

	// Ensure at least 1 damage if the attack hits
	if damage < 1 && typeEffectiveness > 0 {
		damage = 1
	}

	return damage
}

// ApplyDamage applies damage to a Pokemon and returns the new HP.
func ApplyDamage(pokemon *poke.Pokemon, damage int) int {
	pokemon.HP -= damage
	if pokemon.HP < 0 {
		pokemon.HP = 0
	}
	return pokemon.HP
}

// IsFainted checks if a Pokemon has fainted.
func IsFainted(pokemon *poke.Pokemon) bool {
	return pokemon.HP <= 0
}

// GetStatusMessage generates a status message for a turn.
func GetStatusMessage(
	attacker *poke.Pokemon,
	defender *poke.Pokemon,
	move poke.Move,
	damage int,
	typeEffectiveness float64,
) string {
	msg := attacker.Name + " used " + move.Name + "! "

	if typeEffectiveness == 0 {
		msg += "It had no effect..."
	} else if typeEffectiveness > 2.0 {
		msg += "It was super effective!"
	} else if typeEffectiveness >= 1.5 {
		msg += "It was effective!"
	} else if typeEffectiveness < 1.0 && typeEffectiveness > 0 {
		msg += "It was not very effective..."
	}

	msg += " Dealt " + string(rune(damage)) + " damage."

	if IsFainted(defender) {
		msg += " " + defender.Name + " fainted!"
	}

	return msg
}
