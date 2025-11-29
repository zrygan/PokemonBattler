---
marp: true
theme: default
paginate: true
---

# P2P Pokémon Battle Protocol Implementation
## CSNETWK Machine Problem 1T2526

**Team:** zrygan  
**Project:** PokemonBattler  
**Protocol:** PokeProtocol over UDP

---

## Project Overview

- **Language:** Go (Golang)
- **Transport:** UDP with custom reliability layer
- **Architecture:** P2P with Host/Joiner/Spectator roles
- **Features:** Turn-based battles, chat, stickers, team management
- **Total Score Target:** 125/100 points + 15 bonus points

---

## Core Protocol Implementation (30 points)

### UDP Sockets & Message Handling (10 points)

```go
// netio/login.go - UDP Socket Setup
func StartListening(port string) *net.UDPConn {
    addr, err := net.ResolveUDPAddr("udp", ":"+port)
    if err != nil {
        panic(err)
    }
    
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        panic(err)
    }
    
    return conn
}
```

---

### Handshake Process (10 points)

```go
// messages/handshake_request.go
func MakeHandshakeRequest() Message {
    params := make(map[string]interface{})
    return Message{
        MessageType:   HandshakeRequest,
        MessageParams: &params,
    }
}

// Host responds with seed for synchronization
func MakeHandshakeResponse(seed int) Message {
    params := map[string]interface{}{
        "seed": seed,
    }
    return Message{
        MessageType:   HandshakeResponse,
        MessageParams: &params,
    }
}
```

---

### Message Serialization (10 points)

```go
// messages/serialization.go
func (m Message) SerializeMessage() []byte {
    var buffer strings.Builder
    buffer.WriteString(fmt.Sprintf("message_type: %s\n", m.MessageType))
    
    if m.MessageParams != nil {
        for key, value := range *m.MessageParams {
            buffer.WriteString(fmt.Sprintf("%s: %v\n", key, value))
        }
    }
    
    return []byte(buffer.String())
}
```

---

## Game Logic & State Management (30 points)

### Turn-Based Flow (15 points)

```go
// game/battle_flow.go - 4-Step Turn Process
func HandleBattleMessage(msg messages.Message, self peer.PeerDescriptor, 
                        opponent *peer.PeerDescriptor, pokemon *poke.Pokemon, 
                        commMode string) {
    switch msg.MessageType {
    case messages.AttackAnnounce:
        // Step 1: Receive attack announcement
        handleAttackAnnounce(msg, self, opponent, commMode)
        
    case messages.DefenseAnnounce:
        // Step 2: Send defense acknowledgment
        sendDefenseAnnounce(self, opponent, commMode)
```

---

### Turn-Based Flow (15 points; *continuation*)

```go  
    case messages.CalculationReport:
        // Step 3: Process damage calculation
        handleCalculationReport(msg, self, opponent, pokemon, commMode)
        
    case messages.CalculationConfirm:
        // Step 4: Confirm and switch turns
        handleCalculationConfirm(msg)
    }
}
```

---

### Win/Loss Condition (5 points)

```go
// game/battle_runner.go - Game Over Detection
func checkGameOver(pokemon *poke.Pokemon, opponent *poke.Pokemon, 
                  self peer.PeerDescriptor, opponentPeer *peer.PeerDescriptor) bool {
    if opponent.HP <= 0 {
        gameOverMsg := messages.MakeGameOver(pokemon.Name, opponent.Name, 1)
        
        // Send to opponent and spectators
        sendMessage(gameOverMsg, self, opponentPeer, "P2P")
        
        fmt.Printf("\n🏆 Victory! %s defeated %s!\n", pokemon.Name, opponent.Name)
        return true
    }
    return false
}
```

---

### Battle State Synchronization (10 points)

```go
// game/battle_flow.go - Synchronization Check
func validateCalculation(received messages.Message, localDamage int, 
                        localHP int) bool {
    params := *received.MessageParams
    receivedDamage := params["damage_dealt"].(int)
    receivedHP := params["defender_hp_remaining"].(int)
    
    if receivedDamage != localDamage || receivedHP != localHP {
        // Send RESOLUTION_REQUEST for discrepancy
        return false
    }
    return true
}
```

---

## Reliability & Error Handling (30 points)

### Sequence Numbers & ACKs (10 points)

```go
// reliability/reliability.go - Reliable Connection
type ReliableConnection struct {
    sequenceNumber int
    expectedSeqNum int
    pendingAcks    map[int]*PendingMessage
    timeout        time.Duration
    maxRetries     int
}
```

---

### Sequence Numbers & ACKs (10 points; *cont*)

```go
func (rc *ReliableConnection) SendReliable(msg messages.Message, 
                                          dest *net.UDPAddr) (int, error) {
    seqNum := rc.GetNextSequenceNumber()
    rc.pendingAcks[seqNum] = &PendingMessage{
        Message:     msg,
        Destination: dest,
        RetriesLeft: rc.maxRetries,
        LastSent:    time.Now(),
    }
    
    return seqNum, rc.conn.WriteToUDP(msg.SerializeMessage(), dest)
}
```

---

### Retransmission Logic (10 points)

```go
// reliability/reliability.go - Timeout & Retry
func (rc *ReliableConnection) CheckRetransmissions() []int {
    var failedSeqNums []int
    now := time.Now()
    
    for seqNum, pending := range rc.pendingAcks {
        if now.Sub(pending.LastSent) > rc.timeout {
            if pending.RetriesLeft > 0 {
                // Retransmit message
                rc.conn.WriteToUDP(pending.Message.SerializeMessage(), 
                                  pending.Destination)
                pending.LastSent = now
                pending.RetriesLeft--
            } else {
                // Max retries exceeded
                failedSeqNums = append(failedSeqNums, seqNum)
                delete(rc.pendingAcks, seqNum)
            }
        }
    }
    return failedSeqNums
}
```

---

### Discrepancy Resolution (10 points)

```go
// messages/resolution_request.go
func MakeResolutionRequest(attacker, moveUsed string, damage, defenderHP, 
                          seqNum int) Message {
    params := map[string]interface{}{
        "attacker":              attacker,
        "move_used":             moveUsed,
        "damage_dealt":          damage,
        "defender_hp_remaining": defenderHP,
        "sequence_number":       seqNum,
    }
    return Message{
        MessageType:   ResolutionRequest,
        MessageParams: &params,
    }
}
```

---

## Features (25 points)

### Damage Calculation (10 points)

```go
// game/battle_runner.go - RFC-Compliant Damage Formula
func calculateDamage(attacker, defender *poke.Pokemon, move poke.Move) int {
    var attackerStat, defenderStat float64
    
    if move.DamageCategory == poke.Physical {
        attackerStat = float64(attacker.Attack)
        defenderStat = float64(defender.Defense)
    } else {
        attackerStat = float64(attacker.SpecialAttack)
        defenderStat = float64(defender.SpecialDefense)
    }
```
---
### Damage Calculation (10 points; *cont*)

```
    // Type effectiveness calculation
    type1Eff := poke.GetTypeEffectiveness(move.Type, defender.Type1)
    type2Eff := poke.GetTypeEffectiveness(move.Type, defender.Type2)
    if defender.Type2 == "" {
        type2Eff = 1.0
    }
    
    typeEffectiveness := type1Eff * type2Eff
    
    damage := int((attackerStat/defenderStat) * move.BasePower * typeEffectiveness)
    return damage
}
```

---

### Chat Functionality - Text (5 points)

```go
// messages/chat_message.go
func MakeChatMessage(sender, contentType, messageText, stickerData string, 
                    seqNum int) Message {
    params := map[string]interface{}{
        "sender_name":    sender,
        "content_type":   contentType,
        "message_text":   messageText,
        "sticker_data":   stickerData,
        "sequence_number": seqNum,
    }
    return Message{
        MessageType:   ChatMessage,
        MessageParams: &params,
    }
}
```

---

### Chat Functionality - Stickers (5 points)

```go
// game/estickers.go - Base64 Encoded Sticker Support
func LoadEsticker(filePath string) (string, error) {
    // Validate file exists and is image
    if !isImageFile(filePath) {
        return "", fmt.Errorf("file is not a supported image format")
    }
    
    // Check file size (max 10MB)
    if fileInfo.Size() > 10*1024*1024 {
        return "", fmt.Errorf("file size exceeds 10MB limit")
    }
    
    // Validate dimensions (320x320px)
    if !isCorrectDimensions(filePath) {
        return "", fmt.Errorf("image must be exactly 320x320 pixels")
    }
    
    // Encode to Base64
    fileBytes, _ := os.ReadFile(filePath)
    return base64.StdEncoding.EncodeToString(fileBytes), nil
}
```

---

### Verbose Mode (5 points)

```go
// netio/verbose.go - Protocol Message Logging
var Verbose bool

func VerboseEventLog(event string, options *LogOptions) {
    if !Verbose {
        return
    }
    
    timestamp := time.Now().Format("15:04:05")
    logMessage := fmt.Sprintf("[%s] PokeProtocol :: %s", timestamp, event)
    
    if options != nil {
        if options.Name != "" {
            logMessage += fmt.Sprintf(" | Name: %s", options.Name)
        }
        if options.IP != "" && options.Port != "" {
            logMessage += fmt.Sprintf(" | %s:%s", options.IP, options.Port)
        }
    }
    
    fmt.Println(logMessage)
}
```

---

## Code Quality & Design (10 points)

### Readability & Comments (5 points)

```go
// game/battle_flow.go - Well-documented code structure
// HandleBattleMessage processes incoming battle-related messages and routes them
// to appropriate handlers based on the current battle state and message type.
// This function maintains the 4-step turn flow as specified in the PokeProtocol RFC.
func HandleBattleMessage(msg messages.Message, self peer.PeerDescriptor, 
                        opponent *peer.PeerDescriptor, pokemon *poke.Pokemon, 
                        commMode string) {
    // Verbose logging for protocol debugging
    netio.VerboseEventLog(
        fmt.Sprintf("Processing %s message", msg.MessageType),
        &netio.LogOptions{Name: opponent.Name},
    )
    
    switch msg.MessageType {
    // ... handler implementations with clear comments
    }
}
```

---

### Separation of Concerns (5 points)

```
Project Structure:
├── messages/          # Protocol message definitions
├── netio/            # Network I/O and utilities  
├── game/             # Battle logic and state management
├── poke/             # Pokemon data structures and team management
├── peer/             # Peer descriptor and networking
├── reliability/      # UDP reliability layer
├── host/             # Host application entry point
├── joiner/           # Joiner application entry point
└── spectator/        # Spectator application entry point
```

 ---

## Bonus Features (15 points)

### Pokemon Team & Personality Management System

#### Profile Creation & Persistence
```go
// poke/personality.go - Pokemon Personality System
type PokemonProfile struct {
    OriginalName string      `json:"original_name"`
    Nickname     string      `json:"nickname"`
    Personality  string      `json:"personality"`
    Friendship   int         `json:"friendship"`
    Victories    int         `json:"victories"`
    TotalBattles int         `json:"total_battles"`
}
```
---
```go
func (p *PokemonProfile) SaveProfile(trainerName string) error {
    profileDir := filepath.Join(".", "profiles")
    os.MkdirAll(profileDir, 0755)
    
    filename := filepath.Join(profileDir, 
        fmt.Sprintf("%s_%s.json", trainerName, p.OriginalName))
    
    data, _ := json.MarshalIndent(p, "", "  ")
    return os.WriteFile(filename, data, 0644)
}
```

---

### Dynamic Personality System

#### Battle-Responsive Flavor Text
```go
// poke/personality.go - Personality-Based Messages  
func (p *PokemonProfile) GetFlavorText(event string) string {
    displayName := p.GetDisplayName()
    
    switch event {
    case "battle_start":
        switch p.Personality {
        case PersonalityBrave:
            return fmt.Sprintf("%s looks determined and ready to fight!", displayName)
        case PersonalityTimid:
            return fmt.Sprintf("%s nervously takes the field...", displayName)
        case PersonalityJolly:
            return fmt.Sprintf("%s bounces excitedly onto the battlefield!", displayName)
        }
```
---
```go
    case "critical_hit":
        switch p.Personality {
        case PersonalitySassy:
            return fmt.Sprintf("%s delivers a devastating blow with style!", displayName)
        }
    }
    return ""
}
```

---

### Team Management Interface

#### Interactive Customization
```go
// poke/team_manager.go - Interactive Pokemon Customization
func (tm *TeamManager) CustomizePokemon(pokemon *Pokemon) (*PokemonProfile, error) {
    // Try to load existing profile
    profile, err := LoadProfile(tm.TrainerName, pokemon.Name)
    if err == nil {
        fmt.Printf("\nWelcome back! Found existing profile for %s!\n", pokemon.Name)
        profile.DisplayProfile()
        // Option to use existing or create new
    }
    
    // Create new profile with nickname and personality selection
    profile = NewPokemonProfile(pokemon.Name)
```
---
```go
    // Interactive nickname setting
    fmt.Print("\nGive your Pokemon a nickname (or press Enter to skip): ")
    nickname, _ := reader.ReadString('\n')
    
    // Personality selection from 8 options
    fmt.Println("\nChoose a personality for your Pokemon:")
    for i, personality := range AllPersonalities {
        fmt.Printf("%d. %s\n", i+1, personality)
    }
    
    return profile, nil
}
```

---

### Friendship & Battle History Tracking

#### Persistent Statistics
```go
// poke/personality.go - Battle Record Tracking
func (p *PokemonProfile) RecordVictory() {
    p.Victories++
    p.TotalBattles++
    p.IncreaseFriendship(5) // Gain 5 friendship per victory
}

func (p *PokemonProfile) RecordDefeat() {
    p.TotalBattles++
    p.IncreaseFriendship(2) // Gain 2 friendship even in defeat
}
```
---
```go
func (p *PokemonProfile) DisplayProfile() {
    displayName := p.GetDisplayName()
    fmt.Printf("\n=== %s ===\n", displayName)
    fmt.Printf("Personality: %s\n", p.Personality)
    fmt.Printf("Friendship: %d/100 (%s)\n", p.Friendship, p.GetFriendshipLevel())
    fmt.Printf("Battle Record: %d-%d (%.1f%% win rate)\n",
        p.Victories,
        p.TotalBattles-p.Victories,
        float64(p.Victories)/float64(max(p.TotalBattles, 1))*100)
}
```
 
---

## Communication Modes Implementation

### P2P vs Broadcast Mode Support
```go
// game/battle_flow.go - Dual Communication Mode
func sendMessage(msg messages.Message, self peer.PeerDescriptor, 
                opponent *peer.PeerDescriptor, commMode string) {
    msgBytes := msg.SerializeMessage()
    
    if commMode == "BROADCAST" {
        // Broadcast to all spectators and participants
        broadcastAddr := &net.UDPAddr{
            IP:   net.IPv4bcast,
            Port: 50000,
        }
        self.Conn.WriteToUDP(msgBytes, broadcastAddr)
    } else {
        // Direct P2P communication
        self.Conn.WriteToUDP(msgBytes, opponent.Addr)
    }
}
```

---

## AI Usage Acknowledgment

**AI Tools Used:**
- GitHub Copilot for code generation assistance
- ChatGPT for protocol understanding and debugging

**Verification Process:**
- All AI-generated code was thoroughly reviewed and tested
- Protocol compliance verified through extensive testing
- Full understanding demonstrated through custom implementations
- Original architecture and design decisions made by team

**Academic Integrity:** All code represents original work with proper AI tool acknowledgment as per course policy.

---

## Thank You