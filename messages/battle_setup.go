package messages

import (
	"github.com/zrygan/pokemonbattler/game/player"
)

// MakeBattleSetup creates a battle setup message with game configuration.
func MakeBattleSetup(
	p player.Player,
	cmode string, // ensure, only "P" or "B"
	pokeName string,
	atk int8,
	def int8,
) Message {
	params := map[string]any{
		"communication_mode":   cmode,
		"pokemon_name":         pokeName,
		"special_attack_uses":  int(atk),
		"special_defense_uses": int(def),
	}
	return Message{
		MessageType:   BattleSetup,
		MessageParams: &params,
	}
}
