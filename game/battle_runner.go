package game

import (
	"fmt"
	"net"
	"path/filepath"
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
	fmt.Println("\nTip: Type 'chat <message>', use stickers like '/gg', or send image files with 'esticker <filepath>'!")
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
			fmt.Println("Select a move (enter number), 'chat <message>', stickers (/gg), or 'esticker <filepath>': ")

			// Get move selection using non-blocking input
			moveIndex := -1
			for moveIndex < 0 || moveIndex >= len(selfPlayer.PokemonStruct.Moves) {
				select {
				case input := <-inputChan:
					// Check if it's a chat command
					if len(input) > 5 && input[:5] == "chat " {
						chatText := input[5:]
						sendChatMessage(battleCtx, chatText)
						continue
					}

					// Check if it's an esticker command
					if strings.HasPrefix(input, "esticker ") {
						sendChatMessage(battleCtx, input)
						continue
					}

					// Check if it's a sticker
					if strings.HasPrefix(input, "/") {
						sendChatMessage(battleCtx, input)
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
					n, addr, err := selfPlayer.Peer.Conn.ReadFromUDP(buf)
					if err == nil {
						msg := messages.DeserializeMessage(buf[:n])
						if msg.MessageType == messages.ChatMessage {
							processIncomingChat(msg, isHost, battleCtx, buf[:n], addr)
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
						n, addr, err := selfPlayer.Peer.Conn.ReadFromUDP(buf)
						if err == nil {
							msg := messages.DeserializeMessage(buf[:n])
							if msg.MessageType == messages.ChatMessage {
								processIncomingChat(msg, isHost, battleCtx, buf[:n], addr)
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

					// Verbose logging for GAME_OVER
					netio.VerboseEventLog(
						"PokeProtocol: Sent GAME_OVER message to opponent",
						&netio.LogOptions{
							MessageParams: gameOverMsg.MessageParams,
						},
					)

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

							// Verbose logging for GAME_OVER
							netio.VerboseEventLog(
								"PokeProtocol: Sent GAME_OVER message to opponent",
								&netio.LogOptions{
									MessageParams: gameOverMsg.MessageParams,
								},
							)

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
						sendChatMessage(battleCtx, chatText)
					} else if strings.HasPrefix(input, "esticker ") {
						// Treat as esticker command
						sendChatMessage(battleCtx, input)
					} else if strings.HasPrefix(input, "/") {
						// Treat as sticker
						sendChatMessage(battleCtx, input)
					} else {
						// Treat as regular message
						sendChatMessage(battleCtx, input)
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

			// Verbose logging for GAME_OVER
			netio.VerboseEventLog(
				"PokeProtocol: Sent GAME_OVER message to opponent",
				&netio.LogOptions{
					MessageParams: gameOverMsg.MessageParams,
				},
			)

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
	// Clear spectators list to prevent stale connections
	game.Spectators = make([]peer.PeerDescriptor, 0)

	// Display Battle Log
	fmt.Println("\n=== BATTLE END ===")
	fmt.Println("\nBATTLE LOG:")
	for i, entry := range game.BattleLog {
		fmt.Printf("%d. %s\n", i+1, entry)
	}
	fmt.Println() // Update Pokemon profiles after battle
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
func sendChatMessage(battleCtx *BattleContext, messageText string) {
	seqNum := battleCtx.ReliableConn.GetNextSequenceNumber()

	// Check if it's a sticker command or esticker command
	contentType := "TEXT"
	stickerData := ""
	displayText := messageText

	// Check for esticker command first
	if isEsticker, filePath := IsEstickerCommand(messageText); isEsticker {
		base64Data, err := LoadEsticker(filePath)
		if err != nil {
			fmt.Printf("Error loading esticker: %v\n", err)
			// Send a fallback text message so others know someone tried to send an esticker
			messageText = fmt.Sprintf("Failed to send esticker: %s", filepath.Base(filePath))
			contentType = "TEXT"
			displayText = messageText
		} else {
			contentType = "STICKER"
			stickerData = base64Data
			displayText = fmt.Sprintf("[Encoded Sticker: %s]", filepath.Base(filePath))
			messageText = "" // Clear message text for stickers
		}
	} else if strings.HasPrefix(messageText, "/") {
		// Check for regular ASCII art stickers
		if stickerText, exists := Stickers[strings.ToLower(messageText)]; exists {
			contentType = "STICKER"
			stickerData = messageText // For ASCII stickers, use the command as ID
			displayText = stickerText
			messageText = "" // Clear message text for stickers
		}
	}

	msg := messages.MakeChatMessage(
		battleCtx.SelfPlayer.Peer.Name,
		contentType,
		messageText,
		stickerData,
		seqNum,
	)

	msgBytes := msg.SerializeMessage()

	// Verbose logging for CHAT_MESSAGE
	netio.VerboseEventLog(
		"PokeProtocol: Sent CHAT_MESSAGE",
		&netio.LogOptions{
			MessageParams: msg.MessageParams,
		},
	)

	// Send chat message according to communication mode
	switch battleCtx.Game.CommunicationMode {
	case "P": // P2P mode - direct to opponent, explicit spectator broadcast
		battleCtx.SelfPlayer.Peer.Conn.WriteToUDP(msgBytes, battleCtx.OpponentAddr)
		battleCtx.Game.BroadcastToSpectators(msgBytes)
	case "B": // Broadcast mode - send to opponent AND spectators simultaneously
		battleCtx.SelfPlayer.Peer.Conn.WriteToUDP(msgBytes, battleCtx.OpponentAddr)
		battleCtx.Game.BroadcastToSpectators(msgBytes)
	default: // Default to P2P behavior
		battleCtx.SelfPlayer.Peer.Conn.WriteToUDP(msgBytes, battleCtx.OpponentAddr)
		battleCtx.Game.BroadcastToSpectators(msgBytes)
	}

	if contentType == "STICKER" {
		fmt.Printf("You sent sticker: %s\n", displayText)
	} else {
		fmt.Printf("You: %s\n", messageText)
	}
}

// processIncomingChat handles and displays incoming chat messages
func processIncomingChat(msg *messages.Message, isHost bool, battleCtx *BattleContext, msgBytes []byte, senderAddr *net.UDPAddr) {
	// Verbose logging for received CHAT_MESSAGE
	netio.VerboseEventLog(
		"PokeProtocol: Received CHAT_MESSAGE during battle",
		&netio.LogOptions{
			MessageParams: msg.MessageParams,
			MS:            senderAddr.String(),
		},
	)

	params := *msg.MessageParams
	senderName, _ := params["sender_name"].(string)
	contentType, _ := params["content_type"].(string)

	if contentType == "TEXT" {
		if messageText, ok := params["message_text"].(string); ok && messageText != "" {
			fmt.Printf("\n[%s]: %s\n", senderName, messageText)
		}
	} else if contentType == "STICKER" {
		if stickerData, ok := params["sticker_data"].(string); ok && stickerData != "" {
			// Check if it's an ASCII art sticker (starts with /)
			if strings.HasPrefix(stickerData, "/") {
				if stickerText, exists := Stickers[strings.ToLower(stickerData)]; exists {
					fmt.Printf("\n[%s] sent sticker: %s\n", senderName, stickerText)
				} else {
					fmt.Printf("\n[%s] sent an unknown sticker\n", senderName)
				}
			} else {
				// Handle Base64 encoded sticker (esticker)
				filename, err := SaveEsticker(stickerData, senderName)
				if err != nil {
					fmt.Printf("\n[%s] sent an invalid esticker: %v\n", senderName, err)
				} else {
					fmt.Printf("\n[%s] sent an esticker (saved as: %s)\n", senderName, filename)
				}
			}
		}
	}

	// Host relays chat messages according to communication mode
	if isHost {
		// Check if message is from joiner (not from us or spectators)
		isFromOpponent := senderAddr.IP.Equal(battleCtx.OpponentAddr.IP) && senderAddr.Port == battleCtx.OpponentAddr.Port

		if isFromOpponent {
			// Message from joiner - always relay to spectators
			battleCtx.broadcastToSpectators(msgBytes)
		} else {
			// Message from spectator - relay according to communication mode
			switch battleCtx.Game.CommunicationMode {
			case "P": // P2P mode - spectator messages stay with spectators only
				battleCtx.broadcastToSpectatorsExcept(msgBytes, senderAddr)
			case "B": // Broadcast mode - relay to joiner AND other spectators
				battleCtx.SelfPlayer.Peer.Conn.WriteToUDP(msgBytes, battleCtx.OpponentAddr)
				battleCtx.broadcastToSpectatorsExcept(msgBytes, senderAddr)
			default: // Default to P2P behavior
				battleCtx.broadcastToSpectatorsExcept(msgBytes, senderAddr)
			}
		}
	}
}
