package messages

import (
	"fmt"
	"strings"
)

type Message struct {
	MessageType   string
	MessageParams *map[string]string
}

const (
	HandshakeRequest  = "HANDSHAKE_REQUEST"
	HandshakeResponse = "HANDSHAKE_RESPONSE"
)

// SerializeMessage converts a Message to bytes
func (m *Message) SerializeMessage() []byte {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("message_type: %s\n", m.MessageType))
	if m.MessageParams != nil {
		for k, v := range *m.MessageParams {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
	}

	return []byte(sb.String())
}

func DeserializeMessage(bs []byte) *Message {
	msg := &Message{
		MessageParams: &map[string]string{},
	}

	lines := strings.Split(string(bs), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ": ", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		if key == "message_type" {
			msg.MessageType = value
		} else {
			(*msg.MessageParams)[key] = value
		}
	}

	return msg
}
