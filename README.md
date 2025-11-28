# 🎮 PokemonBattler

A feature-rich, peer-to-peer Pokemon battle simulator implementing a custom PokeProtocol over UDP with personality-driven Pokemon companions.

## ✨ Features

### 🔥 Core Battle System
- **Turn-based Pokemon battles** between two players over a network
- **UDP-based PokeProtocol** with custom reliability layer (ACKs and retransmission)
- **Complete type effectiveness system** for all 18 Pokemon types
- **Physical vs Special attack mechanics** with consumable stat boost system
- **803 Pokemon** loaded from comprehensive CSV database
- **Synchronized damage calculation** using seeded RNG for fair play

### 💬 Communication & Social
- **Real-time chat functionality** with text messages and sticker support
- **Spectator mode** for observing battles in real-time
- **P2P and Broadcast communication modes** for flexible network setups
- **24 built-in stickers** (/smile, /gg, /attack, /defend, etc.)

### 🎭 **Pokemon Personality & Profile System** ⭐ *NEW!*
- **Custom nicknames** with emoji support (name your Pikachu "Sparky⚡")
- **8 unique personalities** with distinct battle flavor text (Brave, Timid, Jolly, etc.)
- **Friendship system** that grows with each battle (0-100 scale)
- **Battle statistics tracking** (wins, losses, win rate per Pokemon)
- **Persistent profiles** saved locally and loaded automatically
- **Personality-based flavor text** during battles for immersive experience

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** 
- Two machines on the same network (or two terminals for local testing)

### 🎯 Running a Battle

**Terminal 1 (Host):**
```bash
go run ./host/host.go
```

**Terminal 2 (Joiner):**
```bash
go run ./joiner/joiner.go
```

**Terminal 3 (Spectator - Optional):**
```bash
go run ./spectator/spectator.go
```

### 📋 Battle Setup Flow
1. **Enter your trainer name** (used for Pokemon profiles)
2. **View existing Pokemon profiles** (optional)
3. **Select your Pokemon** from 803 available
4. **Customize Pokemon** with nickname and personality
5. **Allocate stat boosts** (10 points between Special Attack/Defense)
6. **Battle begins!** Host goes first

### 🎮 During Battle
- Select moves (1-4) and decide on stat boost usage
- **Chat anytime** with `chat <message>` or use stickers like `/gg`
- Watch your Pokemon's **personality shine** through flavor text
- See **real-time friendship updates** after each battle

## 📁 Project Structure

```
PokemonBattler/
├── 🎯 Applications
│   ├── host/           - Host application (battle coordinator)
│   ├── joiner/         - Joiner application (battle participant)
│   └── spectator/      - Spectator mode (battle observer)
├── 🎮 Game Engine
│   ├── game/           - Battle engine and core logic
│   │   ├── player/     - Player data structures
│   │   ├── battle.go   - Damage calculation & mechanics
│   │   ├── battle_flow.go - Turn-based battle flow
│   │   ├── battle_runner.go - Main battle loop
│   │   └── setup.go    - Game setup functions
├── 🐾 Pokemon System
│   ├── poke/           - Pokemon data structures & profiles
│   │   ├── mons/       - Pokemon database loader
│   │   ├── personality.go - NEW: Personality system
│   │   ├── team_manager.go - NEW: Profile management
│   │   └── types.go    - Pokemon & move definitions
├── 📡 Networking
│   ├── messages/       - Protocol message definitions
│   ├── netio/          - Network I/O utilities
│   ├── peer/           - Peer connection management
│   └── reliability/    - UDP reliability layer
├── 💾 Data & Config
│   ├── data/           - Pokemon CSV database
│   ├── profiles/       - Pokemon profiles (auto-generated)
│   └── docs/           - Documentation
└── 💬 Communication
    └── chat/           - Chat and sticker handling
```

## 🔧 Protocol Overview

### Core PokeProtocol Messages
```
HANDSHAKE_REQUEST ↔️ HANDSHAKE_RESPONSE
BATTLE_SETUP → BATTLE_SETUP
ATTACK_ANNOUNCE → DEFENSE_ANNOUNCE → CALCULATION_REPORT → CALCULATION_CONFIRM
CHAT_MESSAGE (async)
GAME_OVER
```

### Reliability Features
- **Sequence numbering** for all messages
- **Automatic ACK system** with retransmission
- **Timeout handling** (500ms default)
- **Connection loss detection**

## 🎭 Pokemon Personalities

| Personality | Battle Start | Low HP | Victory |
|-------------|--------------|--------|---------|
| **Brave** | "looks determined!" | "refuses to give up!" | "stands victorious!" |
| **Timid** | "nervously takes field..." | "looks frightened..." | "surprised by victory!" |
| **Jolly** | "bounces excitedly!" | "still hanging in there!" | "jumps for joy!" |
| **Sassy** | "struts confidently!" | "maintains composure..." | "strikes victory pose!" |
| **Serious** | "focuses intently..." | "analyzes situation..." | "nods with satisfaction" |
| **Calm** | "enters peacefully..." | "remains composed..." | "smiles quietly" |
| **Playful** | "prances around!" | "still wants to play!" | "celebrates playfully!" |
| **Proud** | "holds head high!" | "maintains dignity..." | "basks in glory!" |

## 🛠️ Development

### Building Applications
```bash
# Build all applications
go build -o pokemonbattler.exe .

# Build individually
go build -o host.exe ./host/host.go
go build -o joiner.exe ./joiner/joiner.go
go build -o spectator.exe ./spectator/spectator.go
```

### Running Tests
```bash
# Test compilation
go build .

# Run with verbose logging
go run ./host/host.go -verbose
```

## 📖 Documentation

| Document | Description |
|----------|-------------|
| [📋 Implementation Guide](docs/IMPLEMENTATION.md) | Complete technical documentation |
| [🎭 Pokemon Personality System](docs/POKEMON_PERSONALITY.md) | Personality system deep-dive |
| [📊 Feature Summary](FEATURE_SUMMARY.md) | Quick feature overview |
| [🎨 Feature Diagram](docs/FEATURE_DIAGRAM.txt) | Visual system flow |

## 🎯 Example Battle Session

```
=== HOST TERMINAL ===
Welcome to PokeBattler
What is your trainer name? Ash

📚 Your Pokemon Profiles:
=== Sparky (Pikachu) ===
Personality: Jolly
Friendship: 85/100 (Great Friends)
Battle Record: 12-3 (80.0% win rate)

Select a pokemon: Pikachu
🎉 Welcome back! Found existing profile!

=== BATTLE START ===
Sparky bounces excitedly onto the battlefield!
Your Pokemon: Sparky (HP: 100/100)

Your turn!
Select move: 1. Thunderbolt
Sparky used Thunderbolt! Dealt 45 damage.

⚠️ Sparky is hanging in there!
🎊 Victory! Sparky jumps for joy!
Sparky gained friendship! (+5)

Battle Record: 13-3 (81.3% win rate)
```

## 🎮 Chat Commands & Stickers

### Text Chat
```
chat Hello! Good luck!
chat That was a great move!
```

### Stickers
```
/smile   →  :)
/gg      →  ASCII "GG"
/fire    →  (~)
/attack  →  >>--->>
/defend  →  [SHIELD]
/lucky   →  Lucky!
/ouch    →  Ouch!
```

## 🏆 Advanced Features

- **Friendship Levels**: Strangers → Acquaintances → Friends → Good Friends → Great Friends → Best Friends
- **Battle Statistics**: Individual Pokemon win/loss tracking
- **Profile Persistence**: Automatic save/load across sessions  
- **Type Effectiveness**: Complete 18-type interaction matrix
- **Stat Boost Strategy**: Consumable Special Attack/Defense boosts
- **Spectator Broadcasting**: Real-time battle observation
- **Cross-platform**: Works on Windows, macOS, Linux

## 🐛 Troubleshooting

### Common Issues
- **Port conflicts**: Application auto-increments ports if busy
- **Profile errors**: Check `profiles/` directory permissions
- **Network issues**: Ensure both machines on same subnet
- **Pokemon not found**: Check spelling (case-insensitive search available)

### Debug Commands
```bash
# Verbose logging
go run ./host/host.go -v

# Check network connectivity
ping [opponent_ip]

# List profiles
ls profiles/
```

### Declaration of AI Use

The following Generative AI (GenAI) tools were used: Claude Sonnet 4. These were used for generating in-code documentation and populate this README documentation, explaning code, exploring Go's network library with required functions or functionalities, and debugging. All the codes are manually tested and verified by the author.

### References

> This project does not use any implementations outside of Go's standard library.

The Go Programming Language. go.dev.