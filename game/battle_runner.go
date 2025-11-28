package game

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
	"github.com/zrygan/pokemonbattler/reliability"
)

// Predefined stickers mapping
var Stickers = map[string]string{
	"/smile":      ":)",
	"/laugh":      "LOL",
	"/cool":       "B)",
	"/angry":      ">:(",
	"/sad":        ":(",
	"/love":       "<3",
	"/fire":       "(~)",
	"/star":       "*",
	"/thumbsup":   "(Y)",
	"/thumbsdown": "(N)",
	"/hi":         "o/",
	"/bye":        "\\o",
	"/gg":         "  ___  ___ \n / __|/ __|\n| (_ | (_ |\n \\___|\\___|",
	"/nice":       "Nice!",
	"/wow":        "WOW!",
	"/ouch":       "Ouch!",
	"/lucky":      "Lucky!",
	"/unlucky":    "Unlucky!",
	"/attack":     ">>--->>",
	"/defend":     "[SHIELD]",
	"/heal":       "+HP+",
	"/critical":   "***CRIT***",
	"/miss":       "X MISS X",
	"/hit":        "[HIT!]",
}

// RunBattle starts and manages the complete battle loop.
func RunBattle(
	selfPlayer *player.Player,
	opponentPlayer *player.Player,
	seed int,
	commMode string,
	isHost bool,
	spectators []peer.PeerDescriptor,
) {
	// Initialize game
	game := NewGame(seed, commMode)
	if isHost {
		game.Host = selfPlayer
		game.Joiner = opponentPlayer
	} else {
		game.Joiner = selfPlayer
		game.Host = opponentPlayer
	}

	// Add spectators
	for _, spec := range spectators {
		game.AddSpectator(spec)
	}

	if len(spectators) > 0 {
		fmt.Printf("\n%d spectator(s) watching this battle\n", len(spectators))
	}

	// Create reliable connection
	reliableConn := reliability.NewReliableConnection(selfPlayer.Peer.Conn)

	// Create battle context
	battleCtx := &BattleContext{
		Game:         game,
		SelfPlayer:   selfPlayer,
		OpponentAddr: opponentPlayer.Peer.Addr,
		ReliableConn: reliableConn,
		IsHost:       isHost,
	}

	// Set initial state
	game.State = StateWaitingForMove

	fmt.Println("\n=== BATTLE START ===")

	// Show personality flavor text if profile exists
	if selfPlayer.Profile != nil {
		poke.ShowPreBattleMessage(selfPlayer.Profile)
	}

	fmt.Printf("Your Pokemon: %s (HP: %d/%d)\n",
		selfPlayer.PokemonStruct.Name,
		selfPlayer.PokemonStruct.HP,
		selfPlayer.PokemonStruct.MaxHP)
	fmt.Println("\nAvailable Moves:")
	for i, move := range selfPlayer.PokemonStruct.Moves {
		fmt.Printf("%d. %s (Power: %.0f, Type: %s, Category: %s)\n",
			i+1, move.Name, move.BasePower, move.Type, move.DamageCategory)
	}
	fmt.Printf("Special Attack Boosts: %d\n", selfPlayer.SpecialAttackUsesLeft)
	fmt.Printf("Special Defense Boosts: %d\n", selfPlayer.SpecialDefenseUsesLeft)
	fmt.Println("\nTip: Type 'chat <message>' to send a chat message!")
	fmt.Println("Stickers: /smile /laugh /cool /angry /sad /love /fire /star /thumbsup /hi /gg /nice /wow /ouch /lucky /attack /defend /heal /critical /miss /hit")
	fmt.Println("You can chat anytime during the battle, even during opponent's turn!")
	fmt.Println()

	// Start input listener for non-blocking input
	inputChan := netio.StartInputListener()

	// Battle loop
	turnNumber := 1
	winner := ""
	loser := ""

	for game.State != StateGameOver {
		fmt.Printf("\n--- Turn %d ---\n", turnNumber)

		isMyTurn := (isHost && game.CurrentTurn == "host") ||
			(!isHost && game.CurrentTurn == "joiner")

		if isMyTurn {
			fmt.Println("Your turn!")
			fmt.Println("Select a move (enter number) or 'chat <message>': ")

			// Get move selection using non-blocking input
			moveIndex := -1
			for moveIndex < 0 || moveIndex >= len(selfPlayer.PokemonStruct.Moves) {
				select {
				case input := <-inputChan:
					// Check if it's a chat command
					if len(input) > 5 && input[:5] == "chat " {
						chatText := input[5:]
						sendChatMessage(selfPlayer, opponentPlayer, game, chatText)
						continue
					}

					// Check if it's a sticker
					if strings.HasPrefix(input, "/") {
						sendChatMessage(selfPlayer, opponentPlayer, game, input)
						continue
					}

					idx, err := strconv.Atoi(input)
					if err == nil && idx > 0 && idx <= len(selfPlayer.PokemonStruct.Moves) {
						moveIndex = idx - 1
					} else {
						fmt.Println("Invalid selection. Please try again.")
					}
				default:
					// Check for incoming network messages (chat from opponent)
					buf := make([]byte, 65535)
					selfPlayer.Peer.Conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
					n, _, err := selfPlayer.Peer.Conn.ReadFromUDP(buf)
					if err == nil {
						msg := messages.DeserializeMessage(buf[:n])
						if msg.MessageType == messages.ChatMessage {
							processIncomingChat(msg, isHost, battleCtx, buf[:n])
						}
					}
					selfPlayer.Peer.Conn.SetReadDeadline(time.Time{}) // Clear deadline
					time.Sleep(10 * time.Millisecond)                 // Small delay to prevent busy waiting
				}
			}
			selectedMove := selfPlayer.PokemonStruct.Moves[moveIndex]

			// Ask if they want to use a boost
			useBoost := false
			if selectedMove.DamageCategory == poke.Special && selfPlayer.SpecialAttackUsesLeft > 0 {
				fmt.Println("Use a Special Attack boost? (y/n): ")
				boostSelected := false
				for !boostSelected {
					select {
					case boostInput := <-inputChan:
						if boostInput == "y" || boostInput == "Y" {
							useBoost = true
							selfPlayer.SpecialAttackUsesLeft--
						}
						boostSelected = true
					default:
						// Check for incoming network messages
						buf := make([]byte, 65535)
						selfPlayer.Peer.Conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
						n, _, err := selfPlayer.Peer.Conn.ReadFromUDP(buf)
						if err == nil {
							msg := messages.DeserializeMessage(buf[:n])
							if msg.MessageType == messages.ChatMessage {
								processIncomingChat(msg, isHost, battleCtx, buf[:n])
							}
						}
						selfPlayer.Peer.Conn.SetReadDeadline(time.Time{})
						time.Sleep(10 * time.Millisecond)
					}
				}
			}

			// Log the attack
			logEntry := fmt.Sprintf("%s used %s", selfPlayer.PokemonStruct.Name, selectedMove.Name)
			game.BattleLog = append(game.BattleLog, logEntry)

			// Process the turn
			err := battleCtx.ProcessTurn(selectedMove, useBoost)
			if err != nil {
				if err.Error() == "opponent_fainted" {
					// Opponent's Pokemon fainted - we won!
					game.State = StateGameOver
					fmt.Println("\nOpponent's Pokemon fainted! You win!")
					winner = selfPlayer.Peer.Name
					loser = opponentPlayer.Peer.Name
					game.BattleLog = append(game.BattleLog, fmt.Sprintf("%s's Pokemon fainted!", opponentPlayer.Peer.Name))
					game.BattleLog = append(game.BattleLog, fmt.Sprintf("Winner: %s", winner))

					// Send GAME_OVER message to opponent
					seqNum := reliableConn.GetNextSequenceNumber()
					gameOverMsg := messages.MakeGameOver(
						winner,
						loser,
						seqNum,
					)
					gameOverBytes := gameOverMsg.SerializeMessage()
					selfPlayer.Peer.Conn.WriteToUDP(gameOverBytes, opponentPlayer.Peer.Addr)

					// Send game over message according to communication mode
					if isHost {
						// In P2P mode, explicitly send to spectators
						// In broadcast mode, spectators already received the message
						if game.CommunicationMode == "P" {
							game.BroadcastToSpectators(gameOverBytes)
						}
					}
					break
				}
				fmt.Printf("Error during turn: %v\n", err)
				break
			}
		} else {
			fmt.Println("Opponent's turn... waiting...")
			fmt.Println("(You can still type messages and they'll be sent)")

			// Create done channel to signal when turn is complete
			turnDone := make(chan error, 1)

			// Start opponent's turn processing in goroutine
			go func() {
				err := battleCtx.ProcessTurn(poke.Move{}, false)
				turnDone <- err
			}()

			// Handle input while waiting for opponent's turn
			for {
				select {
				case err := <-turnDone:
					// Turn is complete
					if err != nil {
						if err.Error() == "opponent_fainted" {
							// This shouldn't happen on defender's turn, but handle it
							game.State = StateGameOver
							fmt.Println("\nOpponent's Pokemon fainted! You win!")
							winner = selfPlayer.Peer.Name
							loser = opponentPlayer.Peer.Name
							game.BattleLog = append(game.BattleLog, fmt.Sprintf("%s's Pokemon fainted!", opponentPlayer.Peer.Name))
							game.BattleLog = append(game.BattleLog, fmt.Sprintf("Winner: %s", winner))

							// Send GAME_OVER message to opponent
							seqNum := reliableConn.GetNextSequenceNumber()
							gameOverMsg := messages.MakeGameOver(
								winner,
								loser,
								seqNum,
							)
							gameOverBytes := gameOverMsg.SerializeMessage()
							selfPlayer.Peer.Conn.WriteToUDP(gameOverBytes, opponentPlayer.Peer.Addr)

							// Send game over message according to communication mode
							if isHost {
								// In P2P mode, explicitly send to spectators
								// In broadcast mode, spectators already received the message
								if game.CommunicationMode == "P" {
									game.BroadcastToSpectators(gameOverBytes)
								}
							}
							goto exitBattle
						}
						fmt.Printf("Error during opponent's turn: %v\n", err)
						goto exitBattle
					}
					goto turnComplete

				case input := <-inputChan:
					// User typed something during opponent's turn
					if len(input) == 0 {
						continue
					}

					// Check if it's a chat command
					if len(input) > 5 && input[:5] == "chat " {
						chatText := input[5:]
						sendChatMessage(selfPlayer, opponentPlayer, game, chatText)
					} else if strings.HasPrefix(input, "/") {
						// Treat as sticker
						sendChatMessage(selfPlayer, opponentPlayer, game, input)
					} else {
						// Treat as regular message
						sendChatMessage(selfPlayer, opponentPlayer, game, input)
					}
				}
			}
		turnComplete:
		}

		// Display current status
		fmt.Printf("\nYour Pokemon: %s (HP: %d/%d)\n",
			selfPlayer.PokemonStruct.Name,
			selfPlayer.PokemonStruct.HP,
			selfPlayer.PokemonStruct.MaxHP)

		// Show low HP warning if HP is below 30%
		hpPercent := float64(selfPlayer.PokemonStruct.HP) / float64(selfPlayer.PokemonStruct.MaxHP)
		if hpPercent < 0.3 && hpPercent > 0 && selfPlayer.Profile != nil {
			poke.ShowLowHealthMessage(selfPlayer.Profile)
		}

		// Check if self fainted
		if IsFainted(&selfPlayer.PokemonStruct) {
			game.State = StateGameOver
			fmt.Println("\nYour Pokemon fainted! You lose!")
			winner = opponentPlayer.Peer.Name
			loser = selfPlayer.Peer.Name

			game.BattleLog = append(game.BattleLog, fmt.Sprintf("%s's Pokemon fainted!", selfPlayer.Peer.Name))
			game.BattleLog = append(game.BattleLog, fmt.Sprintf("Winner: %s", winner))

			// Send GAME_OVER message
			seqNum := reliableConn.GetNextSequenceNumber()
			gameOverMsg := messages.MakeGameOver(
				winner,
				loser,
				seqNum,
			)
			gameOverBytes := gameOverMsg.SerializeMessage()
			selfPlayer.Peer.Conn.WriteToUDP(gameOverBytes, opponentPlayer.Peer.Addr)

			// Broadcast to spectators
			if isHost {
				game.BroadcastToSpectators(gameOverBytes)
			}
			break
		}

		// Check if opponent fainted (update opponent HP tracking)
		if IsFainted(&opponentPlayer.PokemonStruct) {
			game.State = StateGameOver
			fmt.Println("\nOpponent's Pokemon fainted! You win!")
			winner = selfPlayer.Peer.Name
			loser = opponentPlayer.Peer.Name

			game.BattleLog = append(game.BattleLog, fmt.Sprintf("%s's Pokemon fainted!", opponentPlayer.Peer.Name))
			game.BattleLog = append(game.BattleLog, fmt.Sprintf("Winner: %s", winner))

			// Send GAME_OVER message
			seqNum := reliableConn.GetNextSequenceNumber()
			gameOverMsg := messages.MakeGameOver(
				winner,
				loser,
				seqNum,
			)
			gameOverBytes := gameOverMsg.SerializeMessage()
			selfPlayer.Peer.Conn.WriteToUDP(gameOverBytes, opponentPlayer.Peer.Addr)

			// Broadcast to spectators
			if isHost {
				game.BroadcastToSpectators(gameOverBytes)
			}
			break
		}

		turnNumber++
	}

exitBattle:
	// Display Battle Log
	fmt.Println("\n=== BATTLE END ===")
	fmt.Println("\nBATTLE LOG:")
	for i, entry := range game.BattleLog {
		fmt.Printf("%d. %s\n", i+1, entry)
	}
	fmt.Println()

	// Update Pokemon profiles after battle
	if selfPlayer.Profile != nil {
		won := winner == selfPlayer.Peer.Name

		// Use the original trainer name to ensure profile continuity
		teamManager := poke.NewTeamManager(selfPlayer.TrainerName)

		// Update profile with battle results
		err := teamManager.UpdateProfileAfterBattle(selfPlayer.Profile, won)
		if err != nil {
			fmt.Printf("Warning: Could not save profile: %v\n", err)
		}
	}
}

// ListenForMessages is a helper goroutine that can listen for async messages like chat.
func ListenForMessages(
	selfPlayer *player.Player,
	opponentPlayer *player.Player,
	reliableConn *reliability.ReliableConnection,
	chatHandler func(msg *messages.Message),
) {
	buf := make([]byte, 65535)
	for {
		n, addr, err := selfPlayer.Peer.Conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := messages.DeserializeMessage(buf[:n])

		switch msg.MessageType {
		case messages.ChatMessage:
			// Handle chat message
			if chatHandler != nil {
				chatHandler(msg)
			}
			// Send ACK
			if seqNum, ok := (*msg.MessageParams)["sequence_number"].(int); ok {
				reliableConn.SendAck(seqNum, addr)
			}

		case messages.ACK:
			// Handle ACK
			if ackNum, ok := (*msg.MessageParams)["ack_number"].(int); ok {
				reliableConn.ReceiveAck(ackNum)
			}
		}
	}
}

// sendChatMessage sends a chat message or sticker to the opponent and spectators
func sendChatMessage(selfPlayer *player.Player, opponentPlayer *player.Player, game *Game, messageText string) {
	seqNum := 0 // Simple sequence for chat

	// Check if it's a sticker command
	contentType := "TEXT"
	stickerID := ""
	displayText := messageText

	if strings.HasPrefix(messageText, "/") {
		if stickerText, exists := Stickers[strings.ToLower(messageText)]; exists {
			contentType = "STICKER"
			stickerID = messageText
			displayText = stickerText
			messageText = "" // Clear message text for stickers
		}
	}

	msg := messages.MakeChatMessage(
		selfPlayer.Peer.Name,
		contentType,
		messageText,
		stickerID,
		seqNum,
	)

	msgBytes := msg.SerializeMessage()

	// Send to opponent
	selfPlayer.Peer.Conn.WriteToUDP(msgBytes, opponentPlayer.Peer.Addr)

	// Send chat message according to communication mode
	switch game.CommunicationMode {
	case "P": // P2P mode - explicitly broadcast to spectators
		game.BroadcastToSpectators(msgBytes)
	case "B": // Broadcast mode - spectators receive as part of network broadcast
		game.BroadcastToSpectators(msgBytes)
	default: // Default to broadcasting
		game.BroadcastToSpectators(msgBytes)
	}

	if contentType == "STICKER" {
		fmt.Printf("You sent sticker: %s\n", displayText)
	} else {
		fmt.Printf("You: %s\n", messageText)
	}
}

// processIncomingChat handles and displays incoming chat messages
func processIncomingChat(msg *messages.Message, isHost bool, battleCtx *BattleContext, msgBytes []byte) {
	params := *msg.MessageParams
	senderName, _ := params["sender_name"].(string)
	contentType, _ := params["content_type"].(string)

	if contentType == "TEXT" {
		if messageText, ok := params["message_text"].(string); ok && messageText != "" {
			fmt.Printf("\n[%s]: %s\n", senderName, messageText)
		}
	} else if contentType == "STICKER" {
		if stickerID, ok := params["sticker_data"].(string); ok && stickerID != "" {
			if stickerText, exists := Stickers[strings.ToLower(stickerID)]; exists {
				fmt.Printf("\n[%s] sent sticker: %s\n", senderName, stickerText)
			} else {
				fmt.Printf("\n[%s] sent a sticker\n", senderName)
			}
		}
	}

	// Host relays chat to spectators
	if isHost {
		battleCtx.broadcastToSpectators(msgBytes)
	}
}
