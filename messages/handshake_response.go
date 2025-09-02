package messages

import (
	"math/rand"
)

func MakeHandshakeResponse() Message {
	params := map[string]any{
		"seed": rand.Intn(999),
	}

	return Message{
		MessageType:   HandshakeResponse,
		MessageParams: &params,
	}
}
