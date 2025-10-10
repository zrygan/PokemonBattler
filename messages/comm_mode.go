package messages

// GS_MakeCMode creates a communication mode set up.
// The communication mode here are one of the game.CommunicationModeEnum.
// This is only used by a HOST user, and only HOST users can set the mode.
func GS_MakeCMode(mode string) Message {
	params := map[string]any{
		"cmode": mode,
	}

	return Message{
		MessageType:   GS_COMMMODE,
		MessageParams: &params,
	}
}
