package messages

import (
	"math/rand"
	"strconv"
)

func MakeHandshakeResponse() Message {
	params := map[string]string{
		"seed": strconv.Itoa(rand.Intn(999)),
	}

	return Message{
		MessageType:   HandshakeResponse,
		MessageParams: &params,
	}
}
