package messages

// MakeCalculationConfirm creates a calculation confirmation message.
// This message is sent by a player to confirm that their opponent's calculation matches their own.
func MakeCalculationConfirm(sequenceNumber int) Message {
	params := map[string]any{
		"sequence_number": sequenceNumber,
	}

	return Message{
		MessageType:   CalculationConfirm,
		MessageParams: &params,
	}
}
