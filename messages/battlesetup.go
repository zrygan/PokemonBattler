package messages

import "github.com/zrygan/pokemonbattler/game"

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
