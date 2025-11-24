package messages

// MakeGameOver creates a game over message.
// This message is sent when a Pokemon faints to declare the battle winner.
func MakeGameOver(winner string, loser string, sequenceNumber int) Message {
	params := map[string]any{
		"winner":          winner,
		"loser":           loser,
		"sequence_number": sequenceNumber,
	}

	return Message{
		MessageType:   GameOver,
		MessageParams: &params,
	}
}
