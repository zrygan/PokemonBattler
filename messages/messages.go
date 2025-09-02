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

func (m *Message) SerializeMessage() []byte {
	var sm []byte
	sm = append(sm, fmt.Appendf(nil, "message_type: %s", m.MessageType)...)

	if m.MessageParams != nil {
		for k, v := range *m.MessageParams {
			sm = append(sm, fmt.Appendf(nil, "%s: %s", k, v)...)
		}
	}

	return sm
}

func DeserializeMessage(bs []byte) *Message {
	msg := new(Message)
	s := string(bs)
	// first split it by space
	sArr := strings.Split(s, " ")

	// the type will alwyas be the 2nd element in sArr
	msg.MessageType = sArr[1]

	switch msg.MessageType {
	case HandshakeRequest:
		msg.MessageParams = nil
	}

	return msg
}
