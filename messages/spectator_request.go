package messages

// MakeSpectatorRequest creates a spectator request message.
// This message is sent by a peer to join an existing battle as an observer.
func MakeSpectatorRequest() Message {
	return Message{
		MessageType:   SpectatorRequest,
		MessageParams: &map[string]any{},
	}
}
