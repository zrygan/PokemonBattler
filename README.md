# PokemonBattler

A peer-to-peer Pokemon battle simulator implementing the PokeProtocol over UDP.

## Features

- **Turn-based Pokemon battles** between two players over a network
- **UDP-based protocol** with custom reliability layer (ACKs and retransmission)
- **Type effectiveness system** for all 18 Pokemon types
- **Physical and Special attack distinction** with stat boost system
- **803 Pokemon** loaded from CSV database
- **Chat functionality** with text messages and sticker support
- **Spectator mode** for observing battles
- **P2P and Broadcast communication modes**

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Two machines on the same network (or two terminals for local testing)

### Running a Battle

**Terminal 1 (Host):**
```powershell
go run .\host\host.go
```

**Terminal 2 (Joiner):**
```powershell
go run .\joiner\joiner.go
```

Follow the prompts to:
1. Select Pokemon
2. Allocate stat boosts (10 points total)
3. Battle!

## Project Structure

```
PokemonBattler/
├── chat/           - Chat and sticker handling
├── data/           - Pokemon database (CSV)
├── game/           - Battle engine and logic
├── host/           - Host application
├── joiner/         - Joiner application
├── messages/       - Protocol message types
├── reliability/    - UDP reliability layer
├── poke/           - Pokemon data structures
└── docs/           - Documentation
```

## Protocol Overview

The PokeProtocol implements:
- **Handshaking** for connection establishment
- **Battle Setup** for Pokemon selection and configuration
- **Turn Flow**: ATTACK_ANNOUNCE → DEFENSE_ANNOUNCE → CALCULATION_REPORT → CALCULATION_CONFIRM
- **Damage Calculation** synchronized using seeded RNG
- **Chat Messages** for communication during battle

See [docs/IMPLEMENTATION.md](docs/IMPLEMENTATION.md) for complete details.

## Documentation

- [Implementation Guide](docs/IMPLEMENTATION.md) - Complete technical documentation
- [Protocol Specification](docs/RFC.md) - Full protocol RFC (if available)

## Example Battle Flow

```
1. Host starts and listens for joiners
2. Joiner broadcasts discovery message
3. Host responds with hosting announcement
4. Joiner selects host and sends handshake request
5. Host accepts and sends handshake response with seed
6. Both players select Pokemon and allocate stat boosts
7. Battle begins (host goes first)
8. Turn-based combat with synchronized damage calculation
9. Battle ends when a Pokemon faints
```

## Development

### Building
```powershell
# Build host
go build -o host.exe .\host\host.go

# Build joiner
go build -o joiner.exe .\joiner\joiner.go
```

### Testing
Run host and joiner in separate terminals or on different machines on the same network.

## License

[Add your license here]

## Contributors

[Add contributors here]
