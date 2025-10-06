package messages

func MakeHandshakeRequest(username string) Message {
	param := map[string]any{
		"name": username,
	}
	return Message{
		MessageType:   HandshakeRequest,
		MessageParams: &param, // Request has no Params
	}
}
