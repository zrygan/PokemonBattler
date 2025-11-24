package messages

// MakeResolutionRequest creates a resolution request message.
// This message is sent when a calculation discrepancy is detected.
func MakeResolutionRequest(
	attacker string,
	moveUsed string,
	damageDealt int,
	defenderHPRemaining int,
	sequenceNumber int,
) Message {
	params := map[string]any{
		"attacker":              attacker,
		"move_used":             moveUsed,
		"damage_dealt":          damageDealt,
		"defender_hp_remaining": defenderHPRemaining,
		"sequence_number":       sequenceNumber,
	}

	return Message{
		MessageType:   ResolutionRequest,
		MessageParams: &params,
	}
}
