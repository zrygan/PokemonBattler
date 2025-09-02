package messages

import (
	"fmt"
	"strconv"
	"strings"
)

type Message struct {
	MessageType   string
	MessageParams *map[string]any
}

const (
	HandshakeRequest  = "HANDSHAKE_REQUEST"
	HandshakeResponse = "HANDSHAKE_RESPONSE"
	BattleSetup       = "BATTLE_SETUP"
)

// SerializeMessage converts a Message to bytes
func (m *Message) SerializeMessage() []byte {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("message_type: %s\n", m.MessageType))
	if m.MessageParams != nil {
		for k, v := range *m.MessageParams {
			sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
		}
	}

	return []byte(sb.String())
}

func DeserializeMessage(bs []byte) *Message {

	msg := &Message{
		MessageParams: &map[string]any{},
	}

	lines := strings.SplitSeq(string(bs), "\n")
	for line := range lines {
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
		}

		// check if the value is a numeric type
		if numValue, err := strconv.Atoi(value); err == nil {
			(*msg.MessageParams)[key] = numValue
		} else {
			(*msg.MessageParams)[key] = value
		}

	}

	return msg
}
