package messages

// MakeDefenseAnnounce creates a defense announcement message.
// This message is sent by the defending player to acknowledge the opponent's attack.
func MakeDefenseAnnounce(sequenceNumber int) Message {
	params := map[string]any{
		"sequence_number": sequenceNumber,
	}

	return Message{
		MessageType:   DefenseAnnounce,
		MessageParams: &params,
	}
}
