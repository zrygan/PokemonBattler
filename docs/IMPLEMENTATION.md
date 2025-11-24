# PokeProtocol Implementation

## Overview
This implementation provides a complete P2P Pokemon Battle Protocol (PokeProtocol) over UDP as specified in the RFC. The protocol includes reliability features, spectator support, and chat functionality with sticker support.

## Project Structure

```
PokemonBattler/
├── chat/           - Chat message handling (text and stickers)
├── data/           - Pokemon data (pokemon.csv)
├── game/           - Battle engine and game logic
│   ├── player/     - Player types and structures
│   ├── battle.go   - Damage calculation and battle mechanics
│   ├── battle_flow.go - Turn-based battle flow implementation
│   ├── battle_runner.go - Main battle loop
│   ├── setup.go    - Game setup functions
│   └── types.go    - Game state types
├── host/           - Host application
├── joiner/         - Joiner application
├── messages/       - Protocol message definitions
├── netio/          - Network I/O utilities
├── peer/           - Peer descriptor and connection management
├── poke/           - Pokemon data structures and type effectiveness
│   ├── mons/       - Pokemon database loader
│   ├── pokedata.go - CSV parser for Pokemon data
│   └── types.go    - Pokemon and Move types
└── reliability/    - UDP reliability layer (ACKs, retransmission)
```

## Implemented Features

### Core Protocol Messages
- ✅ HANDSHAKE_REQUEST / HANDSHAKE_RESPONSE
- ✅ SPECTATOR_REQUEST
- ✅ BATTLE_SETUP
- ✅ ATTACK_ANNOUNCE
- ✅ DEFENSE_ANNOUNCE
- ✅ CALCULATION_REPORT
- ✅ CALCULATION_CONFIRM
- ✅ RESOLUTION_REQUEST
- ✅ GAME_OVER
- ✅ CHAT_MESSAGE (text and sticker support)
- ✅ ACK (acknowledgment)

### Reliability Layer
- Sequence number tracking for all messages
- Automatic acknowledgment (ACK) system
- Retransmission with configurable timeout (default: 500ms)
- Maximum retry limit (default: 3 attempts)
- Connection loss detection

### Battle System
- Four-step turn handshake (ATTACK_ANNOUNCE → DEFENSE_ANNOUNCE → CALCULATION_REPORT → CALCULATION_CONFIRM)
- Synchronized damage calculation using seeded RNG
- Physical vs Special attack distinction
- Type effectiveness system (all 18 types supported)
- Special Attack and Special Defense boost system
- HP tracking and fainting detection

### Pokemon Data
- CSV-based Pokemon database (803 Pokemon)
- Stats: HP, Attack, Defense, Special Attack, Special Defense, Speed
- Type 1 and Type 2 support
- Move system with BasePower, Type, and DamageCategory

### Chat System
- Text messages
- Sticker support (Base64 encoded, max 10MB, 320x320px recommended)
- Asynchronous chat during battle
- Spectators can send/receive chat

### Communication Modes
- P2P (Peer-to-Peer): Direct UDP communication
- Broadcast: Local network broadcast for discovery
- Spectator: Observe-only mode

## Usage

### Running as Host
```powershell
go run .\host\host.go
```

The host will:
1. Wait for joiners to broadcast discovery messages
2. Accept or reject connection requests
3. Set communication mode (P2P or Broadcast)
4. Select Pokemon and allocate stat boosts
5. Start the battle (host always goes first)

### Running as Joiner
```powershell
go run .\joiner\joiner.go
```

The joiner will:
1. Broadcast to discover available hosts
2. Select a host to join
3. Wait for battle setup from host
4. Select Pokemon and allocate stat boosts
5. Enter the battle

### Pokemon Selection
When prompted, enter the Pokemon name exactly as it appears in the CSV (e.g., "Pikachu", "Charizard").

### Stat Boost Allocation
You have 10 points to distribute between Special Attack and Special Defense boosts.
These are consumable during battle for special moves.

### Battle Controls
During your turn:
1. Select a move by entering its number (1-4)
2. Choose whether to use a stat boost (if available and move is special)
3. Wait for calculation confirmation
4. View turn results

## Damage Calculation Formula

```
Damage = ((AttackerStat / DefenderStat) * BasePower * TypeEffectiveness) * RandomFactor

Where:
- AttackerStat: Attack (physical) or SpecialAttack (special)
- DefenderStat: Defense (physical) or SpecialDefense (special)
- BasePower: Move's base power (default 1.0)
- TypeEffectiveness: Type1Effectiveness * Type2Effectiveness
- RandomFactor: 0.85 to 1.0 (15% random variation)
```

### Stat Boosts
- Special Attack Boost: 1.5x multiplier to SpecialAttack stat
- Special Defense Boost: 1.5x multiplier to SpecialDefense stat
- Boosts are consumable resources (limited uses)

## Type Effectiveness
Implemented as per standard Pokemon type chart:
- Super Effective: 2.0x damage
- Not Very Effective: 0.5x damage
- No Effect: 0x damage
- Neutral: 1.0x damage
- Dual types: Multiply both effectiveness values

## Network Protocol

### Message Format
All messages use plain text key-value pairs:
```
message_type: MESSAGE_TYPE
key1: value1
key2: value2
...
```

### Reliability
- Every message (except ACK) includes a `sequence_number`
- Receiver must send ACK with corresponding `ack_number`
- Sender retransmits if no ACK received within timeout
- After max retries, connection is considered lost

## Configuration

### Timeouts and Retries
Edit `reliability/reliability.go`:
```go
const (
    DefaultTimeout    = 500 * time.Millisecond
    DefaultMaxRetries = 3
)
```

### Pokemon Data
Pokemon data is loaded from `data/pokemon.csv` on startup.
To add/modify Pokemon, edit the CSV file.

## Known Limitations

1. **Opponent Pokemon Tracking**: Currently simplified - full opponent state tracking would require additional synchronization
2. **Spectator Implementation**: Basic framework in place, full spectator mode requires additional message routing
3. **Move Database**: Moves are auto-generated based on Pokemon types; custom move sets not yet implemented
4. **Error Recovery**: Calculation discrepancies trigger error but don't have full resolution protocol implementation
5. **Network Issues**: No reconnection logic if connection drops

## Future Enhancements

- [ ] Complete spectator mode with full battle observation
- [ ] Custom move sets loaded from data file
- [ ] Status effects (paralysis, burn, sleep, etc.)
- [ ] Multi-turn moves and abilities
- [ ] Battle history and replay system
- [ ] GUI interface
- [ ] Tournament/ladder system
- [ ] Save/load battle state

## Testing

To test the implementation:

1. Start host in one terminal:
   ```powershell
   go run .\host\host.go
   ```

2. Start joiner in another terminal:
   ```powershell
   go run .\joiner\joiner.go
   ```

3. Follow the prompts to set up and play a battle

## Protocol Compliance

This implementation follows the PokeProtocol RFC specifications:
- ✅ Section 1: Introduction and objectives
- ✅ Section 2: Terminology
- ✅ Section 3: Protocol Architecture (P2P, Broadcast, Spectator modes)
- ✅ Section 4: Message Format and Types (all message types implemented)
- ✅ Section 5: State Management and Game Flow
- ✅ Section 6: Damage Calculation (with special attack/defense support)

## Dependencies

```
go.mod
module github.com/zrygan/pokemonbattler

go 1.21+
```

All dependencies are standard Go library packages.
