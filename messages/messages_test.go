package messages

import (
	"strings"
	"testing"
)

func TestSerializeMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      Message
		expected string
	}{
		{
			name: "HandshakeRequest",
			msg: Message{
				MessageType:   HandshakeRequest,
				MessageParams: nil,
			},
			expected: "message_type: HANDSHAKE_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := string(tt.msg.SerializeMessage())

			// for Messages with a Parameter but none is given
			if tt.msg.MessageParams == nil {
				if out != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, out)
				}
				return
			}

			if !strings.Contains(out, "message_type: HANDSHAKE_REQUEST") {
				t.Errorf("missing message_type, got %q", out)
			}
		})
	}
}
