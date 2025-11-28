package game

import (
	"fmt"
	"net"
	"strings"

	"github.com/zrygan/pokemonbattler/game/player"
	"github.com/zrygan/pokemonbattler/messages"
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
	if bc.IsHost && len(bc.Game.Spectators) > 0 {
		bc.Game.BroadcastToSpectators(msg)
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

		// Wait for DEFENSE_ANNOUNCE
		_, err := bc.waitForMessage(messages.DefenseAnnounce)
		if err != nil {
			return err
		}

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

		// Log the attack
		logEntry := fmt.Sprintf("%s used %s and dealt %d damage to opponent (HP: %d)",
			bc.SelfPlayer.PokemonStruct.Name, selectedMove.Name, damage, projectedHP)
		bc.Game.BattleLog = append(bc.Game.BattleLog, logEntry)

		// Wait for CALCULATION_CONFIRM from defender
		_, err = bc.waitForMessage(messages.CalculationConfirm)
		if err != nil {
			return err
		}

		// Switch turns
		bc.switchTurn()

	} else {
		// Defender's turn
		// Wait for ATTACK_ANNOUNCE
		attackMsg, err := bc.waitForMessage(messages.AttackAnnounce)
		if err != nil {
			return err
		}

		// Send DEFENSE_ANNOUNCE
		seqNum := bc.ReliableConn.GetNextSequenceNumber()
		defenseMsg := messages.MakeDefenseAnnounce(seqNum)
		bc.SelfPlayer.Peer.Conn.WriteToUDP(defenseMsg.SerializeMessage(), bc.OpponentAddr)

		// Wait for attacker's CALCULATION_REPORT
		calcMsg, err := bc.waitForMessage(messages.CalculationReport)
		if err != nil {
			return err
		}

		// Extract damage from calculation report
		damage := (*calcMsg.MessageParams)["damage_dealt"].(int)
		defenderHPRemaining := (*calcMsg.MessageParams)["defender_hp_remaining"].(int)

		// Apply damage to self
		bc.SelfPlayer.PokemonStruct.HP = defenderHPRemaining

		// Send CALCULATION_CONFIRM
		seqNum = bc.ReliableConn.GetNextSequenceNumber()
		confirmMsg := messages.MakeCalculationConfirm(seqNum)
		bc.SelfPlayer.Peer.Conn.WriteToUDP(confirmMsg.SerializeMessage(), bc.OpponentAddr)

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
			spectatorName := "Spectator" + addr.String()
			spectator := peer.MakePD(spectatorName, nil, addr)
			bc.Game.AddSpectator(spectator)
			fmt.Printf("\nNew spectator joined: %s\n", addr.String())
			continue // Keep waiting for the actual battle message
		}

		// Ignore chat messages - they're handled by background listener
		if msg.MessageType == messages.ChatMessage {
			// Display the chat message inline
			params := *msg.MessageParams
			senderName, _ := params["sender_name"].(string)
			contentType, _ := params["content_type"].(string)

			if contentType == "TEXT" {
				if messageText, ok := params["message_text"].(string); ok && messageText != "" {
					fmt.Printf("\n[%s]: %s\n", senderName, messageText)
				}
			} else if contentType == "STICKER" {
				if stickerID, ok := params["sticker_data"].(string); ok && stickerID != "" {
					// Display sticker with its visual representation
					if stickerText, exists := Stickers[strings.ToLower(stickerID)]; exists {
						fmt.Printf("\n[%s] sent sticker: %s\n", senderName, stickerText)
					} else {
						fmt.Printf("\n[%s] sent a sticker\n", senderName)
					}
				}
			}

			// Host relays all chat messages
			if bc.IsHost {
				// Check if message is from joiner (not from us or spectators)
				isFromOpponent := addr.IP.Equal(bc.OpponentAddr.IP) && addr.Port == bc.OpponentAddr.Port

				if isFromOpponent {
					// Message from joiner - relay to spectators only
					bc.broadcastToSpectators(buf[:n])
				} else {
					// Message from spectator - relay to joiner AND other spectators
					bc.SelfPlayer.Peer.Conn.WriteToUDP(buf[:n], bc.OpponentAddr)
					bc.broadcastToSpectators(buf[:n])
				}
			}

			continue // Keep waiting for the actual battle message
		}

		// Check for GAME_OVER message - opponent's pokemon fainted
		if msg.MessageType == messages.GameOver {
			bc.Game.State = StateGameOver
			return msg, fmt.Errorf("opponent_fainted")
		}

		if msg.MessageType == msgType {
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

func (bc *BattleContext) verifyCalculations(msg1 messages.Message, msg2 *messages.Message) bool {
	params1 := *msg1.MessageParams
	params2 := *msg2.MessageParams

	return params1["damage_dealt"] == params2["damage_dealt"] &&
		params1["defender_hp_remaining"] == params2["defender_hp_remaining"]
}

func (bc *BattleContext) handleCalculationDiscrepancy(myCalc messages.Message) error {
	seqNum := bc.ReliableConn.GetNextSequenceNumber()
	params := *myCalc.MessageParams

	resMsg := messages.MakeResolutionRequest(
		params["attacker"].(string),
		params["move_used"].(string),
		params["damage_dealt"].(int),
		params["defender_hp_remaining"].(int),
		seqNum,
	)

	bc.SelfPlayer.Peer.Conn.WriteToUDP(resMsg.SerializeMessage(), bc.OpponentAddr)

	return fmt.Errorf("calculation discrepancy detected")
}

func (bc *BattleContext) switchTurn() {
	if bc.Game.CurrentTurn == "host" {
		bc.Game.CurrentTurn = "joiner"
	} else {
		bc.Game.CurrentTurn = "host"
	}
}

func (bc *BattleContext) findMoveByName(pokemon *poke.Pokemon, moveName string) poke.Move {
	for _, move := range pokemon.Moves {
		if move.Name == moveName {
			return move
		}
	}
	return poke.Move{Name: moveName, BasePower: 50, Type: "normal", DamageCategory: poke.Physical}
}

func getOpponentPokemon(bc *BattleContext) *poke.Pokemon {
	if bc.IsHost {
		// Opponent is joiner - we need to track their pokemon separately
		// This is a simplified version; in a real implementation, you'd track opponent state
		return &bc.Game.Joiner.PokemonStruct
	}
	return &bc.Game.Host.PokemonStruct
}
