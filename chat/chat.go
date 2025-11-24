// Package chat provides functionality for peer-to-peer chat with text and sticker support.
package chat

import (
	"encoding/base64"
	"fmt"
	"os"
)

const (
	ContentTypeText    = "TEXT"
	ContentTypeSticker = "STICKER"
	MaxStickerSize     = 10 * 1024 * 1024 // 10MB
)

// ChatHandler manages chat functionality.
type ChatHandler struct {
	SenderName string
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(senderName string) *ChatHandler {
	return &ChatHandler{
		SenderName: senderName,
	}
}

// EncodeStickerFromFile reads a sticker file and encodes it to Base64.
func EncodeStickerFromFile(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	if len(data) > MaxStickerSize {
		return "", fmt.Errorf("sticker file too large: %d bytes (max %d)", len(data), MaxStickerSize)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// DecodeStickerToFile decodes a Base64 sticker string and saves it to a file.
func DecodeStickerToFile(base64Data string, filepath string) error {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0644)
}

// FormatChatMessage formats a chat message for display.
func FormatChatMessage(senderName string, contentType string, messageText string) string {
	if contentType == ContentTypeText {
		return fmt.Sprintf("[%s]: %s", senderName, messageText)
	} else if contentType == ContentTypeSticker {
		return fmt.Sprintf("[%s] sent a sticker", senderName)
	}
	return fmt.Sprintf("[%s] sent a message", senderName)
}
