package messages

// MakeCalculationReport creates a calculation report message.
// This message is sent by both players to report the results of their independent damage calculation.
func MakeCalculationReport(
	attacker string,
	moveUsed string,
	remainingHealth int,
	damageDealt int,
	defenderHPRemaining int,
	statusMessage string,
	sequenceNumber int,
) Message {
	params := map[string]any{
		"attacker":              attacker,
		"move_used":             moveUsed,
		"remaining_health":      remainingHealth,
		"damage_dealt":          damageDealt,
		"defender_hp_remaining": defenderHPRemaining,
		"status_message":        statusMessage,
		"sequence_number":       sequenceNumber,
	}

	return Message{
		MessageType:   CalculationReport,
		MessageParams: &params,
	}
}
