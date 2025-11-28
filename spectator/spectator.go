package main

import (
	"flag"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/zrygan/pokemonbattler/game"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	monsters "github.com/zrygan/pokemonbattler/poke/mons"
)

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

	// Discover hosts
	host := discoverHost(self)
	if host == nil {
		fmt.Println("No hosts found. Exiting...")
		return
	}

	// Send spectator request
	fmt.Printf("\nLOG :: Requesting to spectate %s's battle\n", host.Name)
	spectatorReq := messages.MakeSpectatorRequest()
	self.Conn.WriteToUDP(spectatorReq.SerializeMessage(), host.Addr)

	// Wait for battle to start and observe
	observeBattle(self, host)
}

func discoverHost(self peer.PeerDescriptor) *peer.PeerDescriptor {
	fmt.Println("Searching for active battles...")
	fmt.Println("Listening for 3 seconds...")
	fmt.Println()

	// Broadcast to multiple ports to discover hosts (50000-50010)
	// This allows discovery of hosts that auto-incremented to different ports
	discoveryMsg := messages.MakeJoiningMMB()
	msgBytes := discoveryMsg.SerializeMessage()

	// Try both broadcast and localhost discovery for better reliability
	addresses := []string{
		"255.255.255.255", // Network broadcast
		"127.0.0.1",       // Localhost
		"192.168.68.106",  // Current network IP
	}

	for _, addr := range addresses {
		for port := 50000; port <= 50010; port++ {
			broadcastAddr := &net.UDPAddr{
				IP:   net.ParseIP(addr),
				Port: port,
			}
			self.Conn.WriteToUDP(msgBytes, broadcastAddr)
		}
	}

	// Listen for responses
	self.Conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	defer self.Conn.SetReadDeadline(time.Time{})

	buf := make([]byte, 1024)
	discoveredHosts := make(map[string]peer.PeerDescriptor)

	for {
		n, _, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		msg := messages.DeserializeMessage(buf[:n])
		if msg.MessageType == messages.MMB_HOSTING {
			hostName := (*msg.MessageParams)["name"].(string)
			hostIP := fmt.Sprint((*msg.MessageParams)["ip"])
			hostPort := (*msg.MessageParams)["port"].(int)

			// Create proper address from message params
			hostAddr := &net.UDPAddr{
				IP:   net.ParseIP(hostIP),
				Port: hostPort,
			}

			host := peer.PeerDescriptor{
				Name: hostName,
				Addr: hostAddr,
			}
			discoveredHosts[hostName] = host

			netio.VerboseEventLog(
				"PokeProtocol: Spectator Peer discovered available Host Peer '"+hostName+"'",
				&netio.LogOptions{
					Name: hostName,
					Port: fmt.Sprint(hostPort),
					IP:   hostIP,
				},
			)
		}
	}

	fmt.Println("Discovered Hosts")
	for name, host := range discoveredHosts {
		fmt.Printf("\t%s @ %s:%d\n", name, host.Addr.IP, host.Addr.Port)
	}
	fmt.Println()

	if len(discoveredHosts) == 0 {
		return nil
	}

	// Select host by name
	for {
		hostName := netio.PRLine("Select a host to spectate... (or type /R to search again)")

		// Check for restart command (case-insensitive)
		if strings.ToUpper(hostName) == "/R" {
			return discoverHost(self) // Recursively restart discovery
		}

		// Try exact match first, then case-insensitive
		host, ok := discoveredHosts[hostName]
		if !ok {
			// Try case-insensitive search
			for key, val := range discoveredHosts {
				if strings.EqualFold(key, hostName) {
					host = val
					ok = true
					break
				}
			}
		}

		if ok {
			return &host
		}

		fmt.Println("Host name not found. Try again.")
	}
}

// sendSpectatorChat sends a chat message or sticker from spectator to host
func sendSpectatorChat(self peer.PeerDescriptor, host peer.PeerDescriptor, messageText string) {
	seqNum := 0 // Simple sequence for chat

	// Check if it's a sticker command
	contentType := "TEXT"
	stickerID := ""
	displayText := messageText

	if strings.HasPrefix(messageText, "/") {
		if stickerText, exists := game.Stickers[strings.ToLower(messageText)]; exists {
			contentType = "STICKER"
			stickerID = messageText
			displayText = stickerText
			messageText = "" // Clear message text for stickers
		}
	}

	msg := messages.MakeChatMessage(
		self.Name,
		contentType,
		messageText,
		stickerID,
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
	fmt.Println("Type 'chat <message>' or stickers like '/gg' and press Enter to send chat messages!")
	fmt.Println()

	// Start goroutine to handle chat input from spectator
	go func() {
		for {
			input := netio.PRLine("")
			if len(input) == 0 {
				continue
			}

			// Check if it's a chat command
			messageText := ""
			if len(input) > 5 && input[:5] == "chat " {
				messageText = input[5:]
			} else if strings.HasPrefix(input, "/") {
				messageText = input // Treat as sticker
			} else {
				messageText = input // Treat as regular message
			}

			// Send chat message to host
			sendSpectatorChat(self, *host, messageText)
		}
	}()

	buf := make([]byte, 65535)

	var hostPokemon, joinerPokemon string
	var hostHP, joinerHP int
	var hostMaxHP, joinerMaxHP int
	battleStarted := false

	for {
		n, _, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := messages.DeserializeMessage(buf[:n])

		switch msg.MessageType {
		case messages.BattleSetup:
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
			params := *msg.MessageParams
			moveName := params["move_name"].(string)
			fmt.Printf("Attack announced: %s\n", moveName)

		case messages.CalculationReport:
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
			params := *msg.MessageParams
			winner := params["winner"].(string)
			loser := params["loser"].(string)

			fmt.Printf("\n=== BATTLE END ===\n")
			fmt.Printf("Winner: %s\n", winner)
			fmt.Printf("Loser: %s\n", loser)
			fmt.Println("\nBattle has ended. Press Ctrl+C to exit.")

			// Keep listening for any final messages
			time.Sleep(3 * time.Second)
			return

		case messages.ChatMessage:
			params := *msg.MessageParams
			sender, _ := params["sender_name"].(string)
			contentType, _ := params["content_type"].(string)

			if contentType == "TEXT" {
				if text, ok := params["message_text"].(string); ok && text != "" {
					fmt.Printf("[%s]: %s\n", sender, text)
				}
			} else if contentType == "STICKER" {
				if stickerID, ok := params["sticker_data"].(string); ok && stickerID != "" {
					// Display sticker with its visual representation
					if stickerText, exists := game.Stickers[strings.ToLower(stickerID)]; exists {
						fmt.Printf("[%s] sent sticker: %s\n", sender, stickerText)
					} else {
						fmt.Printf("[%s] sent a sticker\n", sender)
					}
				}
			}
		}
	}
}
