package main

import (
	"flag"
	"fmt"
	"net"
	"time"

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

	fmt.Println("Welcome to PokeBattler - Spectator Mode")
	fmt.Println("(c) Zhean Ganituen /zrygan/, 2025")
	fmt.Println()

	// Create peer descriptor
	self := peer.MakePDFromLogin("spectatorW")
	defer self.Conn.Close()

	// Discover hosts
	host := discoverHost(self)
	if host == nil {
		fmt.Println("No hosts found. Exiting...")
		return
	}

	// Send spectator request
	fmt.Printf("\n🔴 LOG :: Requesting to spectate %s's battle\n", host.Name)
	spectatorReq := messages.MakeSpectatorRequest()
	self.Conn.WriteToUDP(spectatorReq.SerializeMessage(), host.Addr)

	// Wait for battle to start and observe
	observeBattle(self, host)
}

func discoverHost(self peer.PeerDescriptor) *peer.PeerDescriptor {
	fmt.Println("🔍 Searching for active battles...")
	fmt.Println("Listening for 5 seconds...")
	fmt.Println()

	// Broadcast to multiple ports to discover hosts (50000-50010)
	// This allows discovery of hosts that auto-incremented to different ports
	discoveryMsg := messages.MakeJoiningMMB()
	msgBytes := discoveryMsg.SerializeMessage()
	
	for port := 50000; port <= 50010; port++ {
		broadcastAddr := &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		}
		self.Conn.WriteToUDP(msgBytes, broadcastAddr)
	}

	// Listen for responses
	self.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer self.Conn.SetReadDeadline(time.Time{})

	buf := make([]byte, 1024)
	var hosts []peer.PeerDescriptor

	for {
		n, addr, err := self.Conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		msg := messages.DeserializeMessage(buf[:n])
		if msg.MessageType == messages.MMB_HOSTING {
			host := peer.PeerDescriptor{
				Name: (*msg.MessageParams)["name"].(string),
				Addr: addr,
			}
			hosts = append(hosts, host)
			fmt.Printf("📡 Found battle: %s @ %s:%d\n", host.Name, addr.IP, addr.Port)
		}
	}

	if len(hosts) == 0 {
		return nil
	}

	// Select host
	fmt.Println("\nAvailable Battles:")
	for i, h := range hosts {
		fmt.Printf("%d. %s @ %s:%d\n", i+1, h.Name, h.Addr.IP, h.Addr.Port)
	}

	for {
		input := netio.PRLine("\nSelect battle to spectate (enter number): ")
		var idx int
		_, err := fmt.Sscanf(input, "%d", &idx)
		if err == nil && idx > 0 && idx <= len(hosts) {
			return &hosts[idx-1]
		}
		fmt.Println("Invalid selection. Try again.")
	}
}

func observeBattle(self peer.PeerDescriptor, host *peer.PeerDescriptor) {
	fmt.Println("\n🎮 === SPECTATING BATTLE ===")
	fmt.Println("You are now observing the battle. Press Ctrl+C to exit.")
	fmt.Println()

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
					fmt.Printf("\n⚔️  BATTLE: %s vs %s\n", hostPokemon, joinerPokemon)
					fmt.Printf("   %s: %d/%d HP\n", hostPokemon, hostHP, hostMaxHP)
					fmt.Printf("   %s: %d/%d HP\n\n", joinerPokemon, joinerHP, joinerMaxHP)
				}
			}

		case messages.AttackAnnounce:
			params := *msg.MessageParams
			moveName := params["move_name"].(string)
			fmt.Printf("⚡ Attack announced: %s\n", moveName)

		case messages.CalculationReport:
			params := *msg.MessageParams
			attacker := params["attacker"].(string)
			moveName := params["move_used"].(string)
			damage := params["damage_dealt"].(int)
			defenderHP := params["defender_hp_remaining"].(int)
			statusMsg := params["status_message"].(string)

			fmt.Printf("\n📊 %s used %s!\n", attacker, moveName)
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

			fmt.Printf("\n🏆 === BATTLE END ===\n")
			fmt.Printf("Winner: %s\n", winner)
			fmt.Printf("Loser: %s\n", loser)
			fmt.Println("\nBattle has ended. Press Ctrl+C to exit.")

			// Keep listening for any final messages
			time.Sleep(3 * time.Second)
			return

		case messages.ChatMessage:
			params := *msg.MessageParams
			sender := params["sender_name"].(string)
			contentType := params["content_type"].(string)

			if contentType == "TEXT" {
				text := params["message_text"].(string)
				fmt.Printf("💬 [%s]: %s\n", sender, text)
			} else if contentType == "STICKER" {
				fmt.Printf("🎨 [%s]: <sent a sticker>\n", sender)
			}
		}
	}
}
