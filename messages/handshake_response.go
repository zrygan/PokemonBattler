package messages

import (
	"math/rand"
)

// MakeHandshakeResponse creates a handshake response message with a random seed.
// The seed is used to synchronize random number generation between host and joiner.
func MakeHandshakeResponse() Message {
	params := map[string]any{
		"seed": rand.Intn(999),
	}

	return Message{
		MessageType:   HandshakeResponse,
		MessageParams: &params,
	}
}
