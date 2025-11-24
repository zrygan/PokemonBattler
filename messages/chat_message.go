package messages

// MakeChatMessage creates a chat message.
// This message can contain either text or a sticker (Base64 encoded).
// contentType should be "TEXT" or "STICKER".
func MakeChatMessage(
	senderName string,
	contentType string,
	messageText string,
	stickerData string,
	sequenceNumber int,
) Message {
	params := map[string]any{
		"sender_name":     senderName,
		"content_type":    contentType,
		"sequence_number": sequenceNumber,
	}

	if contentType == "TEXT" && messageText != "" {
		params["message_text"] = messageText
	} else if contentType == "STICKER" && stickerData != "" {
		params["sticker_data"] = stickerData
	}

	return Message{
		MessageType:   ChatMessage,
		MessageParams: &params,
	}
}
