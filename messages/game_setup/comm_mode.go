package gamesetup

import (
	"github.com/zrygan/pokemonbattler/messages"
)

// GS_MakeCommMode creates a communication mode set up.
// The communication mode here are one of the game.CommunicationModeEnum.
// This is only used by a HOST user, and only HOST users can set the mode.
func GS_MakeCommMode(mode string) messages.Message {
	params := map[string]any{
		"commMode": mode,
	}

	return messages.Message{
		MessageType:   messages.GS_COMMMODE,
		MessageParams: &params,
	}
}
