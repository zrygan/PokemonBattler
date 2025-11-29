package main

import (
	"flag"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	monsters "github.com/zrygan/pokemonbattler/poke/mons"
)

// Global sequence number for spectator chat messages
var (
	spectatorSeqNum int
	seqMutex        sync.Mutex
)

func getNextSpectatorSeqNum() int {
	seqMutex.Lock()
	defer seqMutex.Unlock()
	spectatorSeqNum++
	return spectatorSeqNum
}

func main() {
	// Parse command-line flags
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging of network events")
	flag.Parse()

	// Set global verbose mode
	netio.Verbose = *verboseFlag

	// Create peer descriptor (Login() will print welcome message)
	self := peer.MakePDFromLogin("spectatorW")

	fmt.Println("Spectator Mode Activated")
	fmt.Println()
	defer self.Conn.Close()

	// Main spectator loop - keep looking for battles
	for {
		// Discover hosts (will loop until a host is found)
		host := discoverHost(self)

		// Send spectator request
		fmt.Printf("\nLOG :: Requesting to spectate %s's battle\n", host.Name)
		spectatorReq := messages.MakeSpectatorRequest()
		reqBytes := spectatorReq.SerializeMessage()

		// Send spectator request multiple times to ensure delivery
		for i := 0; i < 3; i++ {
			self.Conn.WriteToUDP(reqBytes, host.Addr)
			time.Sleep(100 * time.Millisecond) // Small delay between attempts
		} // Verbose logging for SPECTATOR_REQUEST
		netio.VerboseEventLog(
			"PokeProtocol: Sent SPECTATOR_REQUEST to host",
			&netio.LogOptions{
				Name: host.Name,
				IP:   host.Addr.IP.String(),
				Port: strconv.Itoa(host.Addr.Port),
			},
		)

		// Wait for battle to start and observe
		observeBattle(self, host)

		// Battle ended, return to main menu
		fmt.Println("\n=== RETURNING TO MAIN MENU ===")
		fmt.Println("Looking for another battle...")

		// Wait a moment for hosts to finish their cleanup and be ready for new battles
		time.Sleep(3 * time.Second)
		fmt.Println()
	}
}

func discoverHost(self peer.PeerDescriptor) *peer.PeerDescriptor {
	fmt.Println("Searching for active battles...")
	fmt.Println("Listening for 5 seconds...")
	fmt.Println()

	// Broadcast to multiple ports to discover hosts (50000-50010)
	// This allows discovery of hosts that auto-incremented to different ports
	discoveryMsg := messages.MakeJoiningMMB()
	msgBytes := discoveryMsg.SerializeMessage()

	for port := 50000; port <= 50010; port++ {
		broadcastAddr := &net.UDPAddr{
			IP:   net.IPv4bcast, // 255.255.255.255
			Port: port,
		}
		self.Conn.WriteToUDP(msgBytes, broadcastAddr)
	}

	// Listen for responses
	self.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer self.Conn.SetReadDeadline(time.Time{})

	// make a map of discovered hosts (same format as joiner)
	discoveredHosts := make(map[string]string)

	buf := make([]byte, 1024)
	for {
		n, _, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		msg := messages.DeserializeMessage(buf[:n])
		if msg.MessageType == messages.MMB_HOSTING {
			hostName := fmt.Sprint((*msg.MessageParams)["name"])
			hostPort := fmt.Sprint((*msg.MessageParams)["port"])
			hostIP := fmt.Sprint((*msg.MessageParams)["ip"])

			netio.VerboseEventLog(
				"PokeProtocol: Spectator Peer discovered available Host Peer '"+hostName+"'",
				&netio.LogOptions{
					Name: hostName,
					Port: hostPort,
					IP:   hostIP,
				},
			)

			hostDets := fmt.Sprintf("%s %s", hostIP, hostPort)
			discoveredHosts[hostName] = hostDets
		}
	}

	fmt.Println("Discovered Hosts")
	for name, details := range discoveredHosts {
		fmt.Printf("\t%s @ %s\n", name, details)
	}
	fmt.Println()

	// Select host by name (same logic as joiner)
	var hostDets []string = nil
	hostName := ""
	for hostDets == nil {
		// If no hosts were found, show appropriate message
		if len(discoveredHosts) == 0 {
			fmt.Println("No hosts found!")
			fmt.Println("Make sure a host is running and in 'waiting for match' state.")
		}

		hostName = netio.PRLine("Select a host to spectate... (or type /R to search again)")

		// Check for restart command (case-insensitive)
		if strings.ToUpper(hostName) == "/R" {
			return discoverHost(self) // Recursively restart discovery
		}

		// If no hosts available, continue loop to allow /R again
		if len(discoveredHosts) == 0 {
			fmt.Println("No hosts available. Try /R to search again.")
			continue
		}

		// Try exact match first, then case-insensitive
		hostVal, ok := discoveredHosts[hostName]
		if !ok {
			// Try case-insensitive search
			for key, val := range discoveredHosts {
				if strings.EqualFold(key, hostName) {
					hostVal = val
					ok = true
					break
				}
			}
		}
		if ok {
			hostDets = strings.Split(hostVal, " ")
		} else {
			fmt.Println("Host name not found. Try again.")
		}
	}

	pd := peer.MakeRemotePD(hostName, hostDets[0], hostDets[1])
	return &pd
}

// sendSpectatorChat sends a chat message or sticker from spectator to host
func sendSpectatorChat(self peer.PeerDescriptor, host peer.PeerDescriptor, messageText string) {
	seqNum := getNextSpectatorSeqNum()

	// Check if it's a sticker command or esticker command
	contentType := "TEXT"
	stickerData := ""
	displayText := messageText

	// Check for esticker command first
	if isEsticker, filePath := game.IsEstickerCommand(messageText); isEsticker {
		base64Data, err := game.LoadEsticker(filePath)
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
		if stickerText, exists := game.Stickers[strings.ToLower(messageText)]; exists {
			contentType = "STICKER"
			stickerData = messageText // For ASCII stickers, use the command as ID
			displayText = stickerText
			messageText = "" // Clear message text for stickers
		}
	}

	msg := messages.MakeChatMessage(
		self.Name,
		contentType,
		messageText,
		stickerData,
		seqNum,
	)

	msgBytes := msg.SerializeMessage()

	// Send to host (who will relay to joiner and other spectators)
	self.Conn.WriteToUDP(msgBytes, host.Addr)

	if contentType == "STICKER" {
		fmt.Printf("You sent sticker: %s\n", displayText)
	} else {
		fmt.Printf("You: %s\n", messageText)
	}
}

func observeBattle(self peer.PeerDescriptor, host *peer.PeerDescriptor) {
	fmt.Println("\n=== SPECTATING BATTLE ===")
	fmt.Println("You are now observing the battle. Press Ctrl+C to exit.")
	fmt.Println("Type 'chat <message>', stickers like '/gg', or 'esticker <filepath>' to send messages!")
	fmt.Println()

	// Start input listener for non-blocking input (like joiner)
	inputChan := netio.StartInputListener()

	buf := make([]byte, 65535)

	var hostPokemon, joinerPokemon string
	var hostHP, joinerHP int
	var hostMaxHP, joinerMaxHP int
	battleStarted := false

	// Message deduplication to prevent duplicate logging in broadcast mode
	processedMessages := make(map[string]bool)

	for {
		select {
		case input := <-inputChan:
			// Handle user input
			if len(input) == 0 {
				continue
			}

			// Check if it's a chat command
			messageText := ""
			if len(input) > 5 && input[:5] == "chat " {
				messageText = input[5:]
			} else if strings.HasPrefix(input, "esticker ") {
				messageText = input // Treat as esticker command
			} else if strings.HasPrefix(input, "/") {
				messageText = input // Treat as sticker
			} else {
				messageText = input // Treat as regular message
			}

			// Send chat message to host
			sendSpectatorChat(self, *host, messageText)

		default:
			// Check for incoming network messages with short timeout
			self.Conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
			n, _, err := self.Conn.ReadFromUDP(buf)
			if err != nil {
				self.Conn.SetReadDeadline(time.Time{}) // Clear deadline
				time.Sleep(10 * time.Millisecond)      // Small delay to prevent busy waiting
				continue
			}
			self.Conn.SetReadDeadline(time.Time{}) // Clear deadline

			msg := messages.DeserializeMessage(buf[:n])

			// Create unique message ID for deduplication
			msgID := ""
			if msg.MessageParams != nil {
				params := *msg.MessageParams
				if seqNum, ok := params["sequence_number"].(int); ok {
					if sender, ok := params["sender_name"].(string); ok {
						msgID = fmt.Sprintf("%s-%s-%d", msg.MessageType, sender, seqNum)
					} else {
						msgID = fmt.Sprintf("%s-%d", msg.MessageType, seqNum)
					}
				} else {
					// For messages without sequence numbers, use message type and content
					msgID = fmt.Sprintf("%s-%v", msg.MessageType, params)
				}
			} else {
				msgID = msg.MessageType
			}

			// Skip if we've already processed this message
			if processedMessages[msgID] {
				continue
			}
			processedMessages[msgID] = true

			switch msg.MessageType {
			case messages.BattleSetup:
				// Verbose logging for received BATTLE_SETUP
				netio.VerboseEventLog(
					"PokeProtocol: Received BATTLE_SETUP from host",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
					},
				)

				params := *msg.MessageParams
				pokemonName := params["pokemon_name"].(string)

				if !battleStarted {
					if hostPokemon == "" {
						hostPokemon = pokemonName
						if mon, ok := monsters.MONSTERS[pokemonName]; ok {
							hostHP = mon.HP
							hostMaxHP = mon.HP
						}
					} else {
						joinerPokemon = pokemonName
						if mon, ok := monsters.MONSTERS[pokemonName]; ok {
							joinerHP = mon.HP
							joinerMaxHP = mon.HP
						}
						battleStarted = true
						fmt.Printf("\nBATTLE: %s vs %s\n", hostPokemon, joinerPokemon)
						fmt.Printf("   %s: %d/%d HP\n", hostPokemon, hostHP, hostMaxHP)
						fmt.Printf("   %s: %d/%d HP\n\n", joinerPokemon, joinerHP, joinerMaxHP)
					}
				}

			case messages.AttackAnnounce:
				// Verbose logging for received ATTACK_ANNOUNCE
				netio.VerboseEventLog(
					"PokeProtocol: Received ATTACK_ANNOUNCE",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
					},
				)

				params := *msg.MessageParams
				moveName := params["move_name"].(string)
				fmt.Printf("Attack announced: %s\n", moveName)

			case messages.CalculationReport:
				// Verbose logging for received CALCULATION_REPORT
				netio.VerboseEventLog(
					"PokeProtocol: Received CALCULATION_REPORT",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
					},
				)

				params := *msg.MessageParams
				attacker := params["attacker"].(string)
				moveName := params["move_used"].(string)
				damage := params["damage_dealt"].(int)
				defenderHP := params["defender_hp_remaining"].(int)
				statusMsg := params["status_message"].(string)

				fmt.Printf("\n%s used %s!\n", attacker, moveName)
				fmt.Printf("   Damage: %d\n", damage)

				// Update HP tracking
				if attacker == hostPokemon {
					joinerHP = defenderHP
				} else {
					hostHP = defenderHP
				}

				fmt.Printf("   Status: %s\n", statusMsg)
				fmt.Printf("\n   Current HP:\n")
				fmt.Printf("   %s: %d/%d\n", hostPokemon, hostHP, hostMaxHP)
				fmt.Printf("   %s: %d/%d\n\n", joinerPokemon, joinerHP, joinerMaxHP)

			case messages.GameOver:
				// Verbose logging for received GAME_OVER
				netio.VerboseEventLog(
					"PokeProtocol: Received GAME_OVER",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
					},
				)

				params := *msg.MessageParams
				winner := params["winner"].(string)
				loser := params["loser"].(string)

				fmt.Printf("\n=== BATTLE END ===\n")
				fmt.Printf("Winner: %s\n", winner)
				fmt.Printf("Loser: %s\n", loser)
				fmt.Println("\nBattle has ended. Returning to main menu...")

				// Keep listening for any final messages
				time.Sleep(3 * time.Second)
				return

			case messages.ChatMessage:
				// Verbose logging for received CHAT_MESSAGE
				netio.VerboseEventLog(
					"PokeProtocol: Received CHAT_MESSAGE",
					&netio.LogOptions{
						MessageParams: msg.MessageParams,
					},
				)

				params := *msg.MessageParams
				sender, _ := params["sender_name"].(string)
				contentType, _ := params["content_type"].(string)

				if contentType == "TEXT" {
					if text, ok := params["message_text"].(string); ok && text != "" {
						fmt.Printf("[%s]: %s\n", sender, text)
					}
				} else if contentType == "STICKER" {
					if stickerData, ok := params["sticker_data"].(string); ok && stickerData != "" {
						fmt.Printf("Debug: Spectator received sticker data, length: %d bytes, starts with '/': %v\n", len(stickerData), strings.HasPrefix(stickerData, "/"))
						// Check if it's an ASCII art sticker (starts with /)
						if strings.HasPrefix(stickerData, "/") {
							if stickerText, exists := game.Stickers[strings.ToLower(stickerData)]; exists {
								fmt.Printf("[%s] sent sticker: %s\n", sender, stickerText)
							} else {
								fmt.Printf("[%s] sent an unknown sticker\n", sender)
							}
						} else {
							// Handle Base64 encoded sticker (esticker)
							fmt.Printf("Debug: Spectator processing Base64 esticker from %s\n", sender)
							filename, err := game.SaveEsticker(stickerData, sender)
							if err != nil {
								fmt.Printf("[%s] sent an invalid esticker: %v\n", sender, err)
							} else {
								fmt.Printf("[%s] sent an esticker (saved as: %s)\n", sender, filename)
							}
						}
					} else {
						fmt.Printf("Debug: Spectator received STICKER message but no sticker_data found or empty\n")
					}
				}
			}
		}
	}
}
