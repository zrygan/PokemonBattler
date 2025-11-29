package game

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "image/gif" // Register GIF format
)

const (
	MaxStickerSize = 10 * 1024 * 1024 // 10MB
	RequiredWidth  = 320
	RequiredHeight = 320
	StickerSaveDir = "received_stickers"
)

// Esticker represents an encoded sticker with validation
type Esticker struct {
	Data     []byte
	Format   string
	Width    int
	Height   int
	FileSize int
}

// LoadEsticker loads an image file, validates dimensions and size, then encodes it to Base64
func LoadEsticker(filePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("sticker file not found: %s", filePath)
	}

	// Read file
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read sticker file: %v", err)
	}

	// Check file size
	if len(fileData) > MaxStickerSize {
		return "", fmt.Errorf("sticker file too large: %d bytes (max %d bytes)", len(fileData), MaxStickerSize)
	}

	// Decode image to check dimensions
	img, format, err := image.Decode(bytes.NewReader(fileData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %v", err)
	}

	// Check dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != RequiredWidth || height != RequiredHeight {
		return "", fmt.Errorf("invalid sticker dimensions: %dx%d (required: exactly %dx%d pixels)\nTip: Resize your image to exactly 320x320 pixels before sending as an esticker", width, height, RequiredWidth, RequiredHeight)
	}

	// Encode to Base64
	encoded := base64.StdEncoding.EncodeToString(fileData)

	fmt.Printf("Loaded esticker: %s (%s, %dx%d, %d bytes)\n",
		filepath.Base(filePath), format, width, height, len(fileData))

	return encoded, nil
}

// SaveEsticker decodes Base64 sticker data and saves it as an image file
func SaveEsticker(base64Data string, senderName string) (string, error) {
	// Decode Base64
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode Base64 sticker data: %v", err)
	}

	// Check decoded size
	if len(data) > MaxStickerSize {
		return "", fmt.Errorf("decoded sticker too large: %d bytes (max %d bytes)", len(data), MaxStickerSize)
	}

	// Decode image to validate and get format
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to decode sticker image: %v", err)
	}

	// Check dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != RequiredWidth || height != RequiredHeight {
		return "", fmt.Errorf("received esticker has invalid dimensions: %dx%d (required: exactly %dx%d pixels)", width, height, RequiredWidth, RequiredHeight)
	}

	// Create save directory if it doesn't exist
	if err := os.MkdirAll(StickerSaveDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create sticker directory: %v", err)
	}

	// Generate filename with timestamp to avoid conflicts
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("esticker_%s_%s.%s", senderName, timestamp, format)
	filePath := filepath.Join(StickerSaveDir, filename)

	// Write file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create sticker file: %v", err)
	}
	defer file.Close()

	// Save in original format
	switch format {
	case "png":
		err = png.Encode(file, img)
	case "jpeg", "jpg":
		err = jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	default:
		// Default to PNG for unsupported formats
		err = png.Encode(file, img)
		filename = strings.Replace(filename, "."+format, ".png", 1)
	}

	if err != nil {
		return "", fmt.Errorf("failed to save sticker image: %v", err)
	}

	return filename, nil
}

// IsEstickerCommand checks if a command is an esticker command
func IsEstickerCommand(input string) (bool, string) {
	if strings.HasPrefix(input, "esticker ") && len(input) > 9 {
		filePath := strings.TrimSpace(input[9:])
		return true, filePath
	}
	return false, ""
}

// ProcessEstickerCommand processes an esticker command and returns the Base64 data
func ProcessEstickerCommand(input string) (string, error) {
	isEsticker, filePath := IsEstickerCommand(input)
	if !isEsticker {
		return "", fmt.Errorf("not an esticker command")
	}

	return LoadEsticker(filePath)
}
