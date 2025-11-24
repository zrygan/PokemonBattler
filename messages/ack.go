package messages

// MakeAck creates an acknowledgement message.
// This message is sent to confirm receipt of a message with a sequence number.
func MakeAck(ackNumber int) Message {
	params := map[string]any{
		"ack_number": ackNumber,
	}

	return Message{
		MessageType:   ACK,
		MessageParams: &params,
	}
}
