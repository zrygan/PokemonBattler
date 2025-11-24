package messages

// MakeHandshakeRejected creates a handshake rejection message.
// This message is sent by the host when declining a joiner's connection request.
func MakeHandshakeRejected() Message {
	return Message{
		MessageType:   HandshakeRejected,
		MessageParams: &map[string]any{},
	}
}
