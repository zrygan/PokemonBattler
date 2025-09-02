package messages

func MakeHandshakeRequest() Message {
	return Message{
		MessageType:   HandshakeRequest,
		MessageParams: nil, // Request has no Params
	}
}
