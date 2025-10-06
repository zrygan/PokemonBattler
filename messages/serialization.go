package messages

import (
	"fmt"
	"strconv"
	"strings"
)

// SerializeMessage converts a Message to a byte slice for network transmission.
// The format is a simple key-value text protocol with newline separation.
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

// DeserializeMessage converts a byte slice back to a Message struct.
// It parses the key-value text protocol and automatically converts numeric strings to integers.
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
