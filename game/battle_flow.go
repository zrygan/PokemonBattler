package game

import (
	"fmt"
	"net"
	"strings"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/messages"
	"github.com/zrygan/pokemonbattler/netio"
	"github.com/zrygan/pokemonbattler/peer"
	"github.com/zrygan/pokemonbattler/poke"
	"github.com/zrygan/pokemonbattler/reliability"
)

// BattleContext contains all information needed to run a battle.
type BattleContext struct {
	Game         *Game
	SelfPlayer   *player.Player
	OpponentAddr *net.UDPAddr
	ReliableConn *reliability.ReliableConnection
	IsHost       bool
}

// broadcastToSpectators sends a message to all spectators
func (bc *BattleContext) broadcastToSpectators(msg []byte) {
	if len(bc.Game.Spectators) > 0 {
		bc.Game.BroadcastToSpectators(msg)
	}
}

// broadcastToSpectatorsExcept sends a message to all spectators except the specified address
func (bc *BattleContext) broadcastToSpectatorsExcept(msg []byte, exceptAddr *net.UDPAddr) {
	if bc.Game.Host == nil || bc.Game.Host.Peer.Conn == nil {
		return // No host connection available
	}

	for _, spectator := range bc.Game.Spectators {
		// Skip the spectator who sent the message
		if spectator.Addr.IP.Equal(exceptAddr.IP) && spectator.Addr.Port == exceptAddr.Port {
			continue
		}
		bc.Game.Host.Peer.Conn.WriteToUDP(msg, spectator.Addr)
	}
}

// ProcessTurn handles the attack, defense, and calculation phases of a turn.
func (bc *BattleContext) ProcessTurn(selectedMove poke.Move, useAttackBoost bool) error {
	isMyTurn := (bc.IsHost && bc.Game.CurrentTurn == "host") ||
		(!bc.IsHost && bc.Game.CurrentTurn == "joiner")

	if isMyTurn {
		// Attacker's turn
		// Send ATTACK_ANNOUNCE
		seqNum := bc.ReliableConn.GetNextSequenceNumber()
		attackMsg := messages.MakeAttackAnnounce(selectedMove.Name, seqNum)
		attackMsgBytes := attackMsg.SerializeMessage()
		bc.SelfPlayer.Peer.Conn.WriteToUDP(attackMsgBytes, bc.OpponentAddr)
		bc.broadcastToSpectators(attackMsgBytes)

		// Verbose logging for ATTACK_ANNOUNCE
		netio.VerboseEventLog(
			"PokeProtocol: Sent ATTACK_ANNOUNCE to opponent",
			&netio.LogOptions{
				MessageParams: attackMsg.MessageParams,
			},
		)

		// Wait for DEFENSE_ANNOUNCE
		defenseMsg, err := bc.waitForMessage(messages.DefenseAnnounce)
		if err != nil {
			return err
		}

		// Verbose logging for received DEFENSE_ANNOUNCE
		netio.VerboseEventLog(
			"PokeProtocol: Received DEFENSE_ANNOUNCE from opponent",
			&netio.LogOptions{
				MessageParams: defenseMsg.MessageParams,
			},
		)

		// Calculate damage (attacker's calculation is authoritative)
		opponentPokemon := getOpponentPokemon(bc)
		damage := bc.calculateDamage(
			&bc.SelfPlayer.PokemonStruct,
			opponentPokemon,
			selectedMove,
			useAttackBoost,
			false,
		)

		// Calculate projected HP
		projectedHP := opponentPokemon.HP - damage
		if projectedHP < 0 {
			projectedHP = 0
		}

		// Update opponent's HP in our tracking
		opponentPokemon.HP = projectedHP

		// Send CALCULATION_REPORT with the damage
		seqNum = bc.ReliableConn.GetNextSequenceNumber()
		calcMsg := bc.makeCalculationReport(
			bc.SelfPlayer.PokemonStruct.Name,
			selectedMove.Name,
			bc.SelfPlayer.PokemonStruct.HP,
			damage,
			projectedHP,
			seqNum,
		)
		calcMsgBytes := calcMsg.SerializeMessage()
		bc.SelfPlayer.Peer.Conn.WriteToUDP(calcMsgBytes, bc.OpponentAddr)
		bc.broadcastToSpectators(calcMsgBytes)

		// Verbose logging for CALCULATION_REPORT
		netio.VerboseEventLog(
			"PokeProtocol: Sent CALCULATION_REPORT to opponent",
			&netio.LogOptions{
				MessageParams: calcMsg.MessageParams,
			},
		)

		// Log the attack
		logEntry := fmt.Sprintf("%s used %s and dealt %d damage to opponent (HP: %d)",
			bc.SelfPlayer.PokemonStruct.Name, selectedMove.Name, damage, projectedHP)
		bc.Game.BattleLog = append(bc.Game.BattleLog, logEntry)

		// Wait for CALCULATION_CONFIRM from defender
		confirmMsg, err := bc.waitForMessage(messages.CalculationConfirm)
		if err != nil {
			return err
		}

		// Verbose logging for received CALCULATION_CONFIRM
		netio.VerboseEventLog(
			"PokeProtocol: Received CALCULATION_CONFIRM from opponent",
			&netio.LogOptions{
				MessageParams: confirmMsg.MessageParams,
			},
		)

		// Switch turns
		bc.switchTurn()

	} else {
		// Defender's turn
		// Wait for ATTACK_ANNOUNCE
		attackMsg, err := bc.waitForMessage(messages.AttackAnnounce)
		if err != nil {
			return err
		}

		// Verbose logging for received ATTACK_ANNOUNCE
		netio.VerboseEventLog(
			"PokeProtocol: Received ATTACK_ANNOUNCE from opponent",
			&netio.LogOptions{
				MessageParams: attackMsg.MessageParams,
			},
		)

		// Send DEFENSE_ANNOUNCE
		seqNum := bc.ReliableConn.GetNextSequenceNumber()
		defenseMsg := messages.MakeDefenseAnnounce(seqNum)
		bc.SelfPlayer.Peer.Conn.WriteToUDP(defenseMsg.SerializeMessage(), bc.OpponentAddr)

		// Verbose logging for sent DEFENSE_ANNOUNCE
		netio.VerboseEventLog(
			"PokeProtocol: Sent DEFENSE_ANNOUNCE to opponent",
			&netio.LogOptions{
				MessageParams: defenseMsg.MessageParams,
			},
		)

		// Wait for attacker's CALCULATION_REPORT
		calcMsg, err := bc.waitForMessage(messages.CalculationReport)
		if err != nil {
			return err
		}

		// Verbose logging for received CALCULATION_REPORT
		netio.VerboseEventLog(
			"PokeProtocol: Received CALCULATION_REPORT from opponent",
			&netio.LogOptions{
				MessageParams: calcMsg.MessageParams,
			},
		)

		// Extract damage from calculation report
		damage := (*calcMsg.MessageParams)["damage_dealt"].(int)
		defenderHPRemaining := (*calcMsg.MessageParams)["defender_hp_remaining"].(int)

		// Apply damage to self
		bc.SelfPlayer.PokemonStruct.HP = defenderHPRemaining

		// Send CALCULATION_CONFIRM
		seqNum = bc.ReliableConn.GetNextSequenceNumber()
		confirmMsg := messages.MakeCalculationConfirm(seqNum)
		bc.SelfPlayer.Peer.Conn.WriteToUDP(confirmMsg.SerializeMessage(), bc.OpponentAddr)

		// Verbose logging for sent CALCULATION_CONFIRM
		netio.VerboseEventLog(
			"PokeProtocol: Sent CALCULATION_CONFIRM to opponent",
			&netio.LogOptions{
				MessageParams: confirmMsg.MessageParams,
			},
		)

		// Display what happened
		moveName := (*attackMsg.MessageParams)["move_name"].(string)
		attackerName := (*calcMsg.MessageParams)["attacker"].(string)
		fmt.Printf("\n%s used %s! Dealt %d damage.\n", attackerName, moveName, damage)

		// Log the event
		logEntry := fmt.Sprintf("%s used %s and dealt %d damage to %s (HP: %d/%d)",
			attackerName, moveName, damage, bc.SelfPlayer.PokemonStruct.Name,
			bc.SelfPlayer.PokemonStruct.HP, bc.SelfPlayer.PokemonStruct.MaxHP)
		bc.Game.BattleLog = append(bc.Game.BattleLog, logEntry)

		// Switch turns
		bc.switchTurn()
	}

	return nil
}

// Helper functions

func (bc *BattleContext) calculateDamage(
	attacker *poke.Pokemon,
	defender *poke.Pokemon,
	move poke.Move,
	attackBoost bool,
	defenseBoost bool,
) int {
	damage := CalculateDamage(attacker, defender, move, attackBoost, defenseBoost, bc.Game.RNG)
	return damage
}

func (bc *BattleContext) waitForMessage(msgType string) (*messages.Message, error) {
	buf := make([]byte, 65535)
	for {
		n, addr, err := bc.SelfPlayer.Peer.Conn.ReadFromUDP(buf)
		if err != nil {
			return nil, err
		}

		msg := messages.DeserializeMessage(buf[:n])

		// Handle spectators joining mid-battle (host only)
		if msg.MessageType == messages.SpectatorRequest && bc.IsHost {
			// Verbose logging for received SPECTATOR_REQUEST
			netio.VerboseEventLog(
				"PokeProtocol: Received SPECTATOR_REQUEST from new spectator",
				&netio.LogOptions{
					MS: addr.String(),
				},
			)

			spectatorName := "Spectator" + addr.String()
			spectator := peer.MakePD(spectatorName, nil, addr)
			bc.Game.AddSpectator(spectator)
			fmt.Printf("\nNew spectator joined: %s\n", addr.String())
			continue // Keep waiting for the actual battle message
		}

		// Ignore chat messages - they're handled by background listener
		if msg.MessageType == messages.ChatMessage {
			// Verbose logging for received CHAT_MESSAGE
			netio.VerboseEventLog(
				"PokeProtocol: Received CHAT_MESSAGE",
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
					MS:            addr.String(),
				},
			)

			// Display the chat message inline
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
			if bc.IsHost {
				// Check if message is from joiner (not from us or spectators)
				isFromOpponent := addr.IP.Equal(bc.OpponentAddr.IP) && addr.Port == bc.OpponentAddr.Port

				if isFromOpponent {
					// Message from joiner - always relay to spectators
					bc.broadcastToSpectators(buf[:n])
				} else {
					// Message from spectator - relay according to communication mode
					switch bc.Game.CommunicationMode {
					case "P": // P2P mode - spectator messages stay with spectators only
						bc.broadcastToSpectatorsExcept(buf[:n], addr)
					case "B": // Broadcast mode - relay to joiner AND other spectators
						bc.SelfPlayer.Peer.Conn.WriteToUDP(buf[:n], bc.OpponentAddr)
						bc.broadcastToSpectatorsExcept(buf[:n], addr)
					default: // Default to P2P behavior
						bc.broadcastToSpectatorsExcept(buf[:n], addr)
					}
				}
			}

			continue // Keep waiting for the actual battle message
		}

		// Check for GAME_OVER message - opponent's pokemon fainted
		if msg.MessageType == messages.GameOver {
			// Verbose logging for received GAME_OVER
			netio.VerboseEventLog(
				"PokeProtocol: Received GAME_OVER from opponent",
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
					MS:            addr.String(),
				},
			)

			bc.Game.State = StateGameOver
			return msg, fmt.Errorf("opponent_fainted")
		}

		if msg.MessageType == msgType {
			// Verbose logging for received battle message
			netio.VerboseEventLog(
				fmt.Sprintf("PokeProtocol: Received %s from opponent", msgType),
				&netio.LogOptions{
					MessageParams: msg.MessageParams,
					MS:            addr.String(),
				},
			)

			return msg, nil
		}
	}
}

func (bc *BattleContext) makeCalculationReport(
	attackerName string,
	moveName string,
	attackerHP int,
	damage int,
	defenderHP int,
	seqNum int,
) messages.Message {
	typeEff := 1.0 // Calculate type effectiveness
	statusMsg := GetStatusMessage(
		&bc.SelfPlayer.PokemonStruct,
		getOpponentPokemon(bc),
		poke.Move{Name: moveName},
		damage,
		typeEff,
	)

	return messages.MakeCalculationReport(
		attackerName,
		moveName,
		attackerHP,
		damage,
		defenderHP,
		statusMsg,
		seqNum,
	)
}

func (bc *BattleContext) switchTurn() {
	if bc.Game.CurrentTurn == "host" {
		bc.Game.CurrentTurn = "joiner"
	} else {
		bc.Game.CurrentTurn = "host"
	}
}

func getOpponentPokemon(bc *BattleContext) *poke.Pokemon {
	if bc.IsHost {
		// Opponent is joiner - we need to track their pokemon separately
		// This is a simplified version; in a real implementation, you'd track opponent state
		return &bc.Game.Joiner.PokemonStruct
	}
	return &bc.Game.Host.PokemonStruct
}
