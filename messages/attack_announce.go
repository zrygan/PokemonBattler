package messages

// MakeAttackAnnounce creates an attack announcement message.
// This message is sent by the attacking player to announce their move choice.
func MakeAttackAnnounce(moveName string, sequenceNumber int) Message {
	params := map[string]any{
		"move_name":       moveName,
		"sequence_number": sequenceNumber,
	}

	return Message{
		MessageType:   AttackAnnounce,
		MessageParams: &params,
	}
}
