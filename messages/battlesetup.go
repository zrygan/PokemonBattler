package messages

import "github.com/zrygan/pokemonbattler/game"

// MakeBattleSetup creates a battle setup message with game configuration.
// Includes communication mode, pokemon selection, and stat boost allocation.
func MakeBattleSetup(
	commMode game.CommunicationModeEnum,
	pokemonName string,
	userStatBoost game.StatBoosts,
) Message {
	params := map[string]any{
		"communication_mode": commMode,
		"pokemon":            pokemonName,
		"stat_boosts":        userStatBoost,
	}

	return Message{
		MessageType:   BattleSetup,
		MessageParams: &params,
	}
}
