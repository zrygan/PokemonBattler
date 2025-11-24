# Testing Guide for PokeProtocol

## Quick Test Scenarios

### Scenario 1: Basic Local Battle
Test the complete battle flow on a single machine.

**Setup:**
1. Open two PowerShell terminals
2. Navigate to project directory in both

**Terminal 1 (Host):**
```powershell
go run .\host\host.go
```
- Enter name: "TestHost"
- Wait for joiner

**Terminal 2 (Joiner):**
```powershell
go run .\joiner\joiner.go
```
- Enter name: "TestJoiner"
- Select "TestHost" from discovered hosts
- Host accepts connection

**Battle Setup:**
- Both select Pokemon (e.g., "Pikachu", "Charizard")
- Allocate stat boosts (e.g., 5/5 or 7/3)

**Expected Result:**
- Battle starts
- Host takes first turn
- Turn alternation works
- Damage calculation synchronized
- Battle ends when Pokemon faints

---

### Scenario 2: Type Effectiveness Test
Verify type effectiveness calculations.

**Test Cases:**
1. **Fire vs Grass** (Super Effective)
   - Host: Charizard
   - Joiner: Venusaur
   - Expected: Fire attacks deal 2x damage

2. **Water vs Fire** (Super Effective)
   - Host: Blastoise
   - Joiner: Charizard
   - Expected: Water attacks deal 2x damage

3. **Electric vs Ground** (No Effect)
   - Host: Pikachu
   - Joiner: Sandslash
   - Expected: Electric attacks deal 0 damage

4. **Normal vs Ghost** (No Effect)
   - Host: Raticate
   - Joiner: Gengar
   - Expected: Normal attacks deal 0 damage

---

### Scenario 3: Stat Boost Testing
Verify special attack/defense boosts work correctly.

**Setup:**
- Host: Pick Pokemon with special moves
- Allocate all 10 points to Special Attack

**Test:**
1. Use a special move WITHOUT boost
2. Note damage
3. Next turn, use same move WITH boost
4. Verify damage is ~1.5x higher

---

### Scenario 4: Reliability Layer Test
Test ACK and retransmission system.

**Method 1: Artificial Delay**
Add delay in message processing to observe retransmissions.

**Method 2: Network Monitoring**
Use Wireshark to capture UDP packets:
- Filter: `udp.port == 50000`
- Look for:
  - Original message
  - ACK response
  - Retransmissions (if no ACK)

---

### Scenario 5: Chat Testing
Test chat message functionality.

**Note:** Current implementation has chat framework. To fully test:

1. Add chat input handling in battle loop
2. Send text message during battle
3. Verify opponent receives message
4. Test sticker encoding/decoding:
   ```go
   stickerData, _ := chat.EncodeStickerFromFile("test.png")
   chat.DecodeStickerToFile(stickerData, "received.png")
   ```

---

## Manual Testing Checklist

### Connection Phase
- [ ] Host starts and listens successfully
- [ ] Joiner broadcasts discovery message
- [ ] Host receives and responds to discovery
- [ ] Joiner displays available hosts
- [ ] Joiner can select a host
- [ ] Handshake request sent
- [ ] Handshake response received with seed
- [ ] Communication mode set successfully

### Setup Phase
- [ ] Pokemon selection works
- [ ] Invalid Pokemon names rejected
- [ ] Stat boost allocation validates (sum ≤ 10)
- [ ] BATTLE_SETUP messages exchanged
- [ ] Both players ready to start

### Battle Phase
- [ ] Host takes first turn
- [ ] Move selection UI works
- [ ] ATTACK_ANNOUNCE sent
- [ ] DEFENSE_ANNOUNCE received
- [ ] Damage calculation synchronized
- [ ] CALCULATION_REPORT exchanged
- [ ] CALCULATION_CONFIRM sent
- [ ] Turn switches correctly
- [ ] HP updates display correctly
- [ ] Stat boosts consume when used
- [ ] Type effectiveness applied correctly

### End Game
- [ ] Battle ends when Pokemon faints
- [ ] GAME_OVER message sent
- [ ] Winner/loser displayed correctly
- [ ] Application exits gracefully

---

## Automated Testing Ideas

### Unit Tests

**Test 1: Damage Calculation**
```go
func TestDamageCalculation(t *testing.T) {
    rng := rand.New(rand.NewSource(42))
    
    attacker := &poke.Pokemon{
        Attack: 100,
        SpecialAttack: 120,
    }
    
    defender := &poke.Pokemon{
        Defense: 80,
        SpecialDefense: 100,
        Type1: "water",
    }
    
    move := poke.Move{
        BasePower: 50,
        Type: "electric",
        DamageCategory: poke.Special,
    }
    
    damage := game.CalculateDamage(attacker, defender, move, false, false, rng)
    
    // Verify damage is calculated correctly
    // Should be: (120/100) * 50 * 2.0 * random = ~120 damage
    if damage < 100 || damage > 130 {
        t.Errorf("Expected damage around 120, got %d", damage)
    }
}
```

**Test 2: Type Effectiveness**
```go
func TestTypeEffectiveness(t *testing.T) {
    tests := []struct {
        attackType string
        defenseType string
        expected float64
    }{
        {"fire", "grass", 2.0},
        {"water", "fire", 2.0},
        {"electric", "ground", 0.0},
        {"normal", "ghost", 0.0},
        {"fire", "water", 0.5},
    }
    
    for _, tt := range tests {
        result := poke.GetTypeEffectiveness(tt.attackType, tt.defenseType)
        if result != tt.expected {
            t.Errorf("Type %s vs %s: expected %.1f, got %.1f",
                tt.attackType, tt.defenseType, tt.expected, result)
        }
    }
}
```

**Test 3: Message Serialization**
```go
func TestMessageSerialization(t *testing.T) {
    original := messages.MakeAttackAnnounce("Thunderbolt", 5)
    
    serialized := original.SerializeMessage()
    deserialized := messages.DeserializeMessage(serialized)
    
    if deserialized.MessageType != messages.AttackAnnounce {
        t.Error("Message type mismatch")
    }
    
    if (*deserialized.MessageParams)["move_name"] != "Thunderbolt" {
        t.Error("Move name mismatch")
    }
    
    if (*deserialized.MessageParams)["sequence_number"] != 5 {
        t.Error("Sequence number mismatch")
    }
}
```

---

## Network Testing

### Local Network Test
1. Connect two machines to same WiFi/LAN
2. Note host machine's IP address
3. Run host on one machine
4. Run joiner on another
5. Verify discovery works across network

### Firewall Test
Ensure UDP port 50000 is open:

**Windows:**
```powershell
New-NetFirewallRule -DisplayName "Pokemon Battle" -Direction Inbound -Protocol UDP -LocalPort 50000 -Action Allow
```

**Test connectivity:**
```powershell
Test-NetConnection -ComputerName <host-ip> -Port 50000 -InformationLevel Detailed
```

---

## Performance Testing

### Latency Test
Measure round-trip time for turn completion:
1. Start battle
2. Time from ATTACK_ANNOUNCE to CALCULATION_CONFIRM
3. Should be < 100ms on local network
4. Should be < 500ms on internet

### Throughput Test
Send multiple messages rapidly:
1. Test chat message sending
2. Verify ACK system handles load
3. Check for message loss

---

## Error Handling Tests

### Disconnection Test
1. Start battle
2. Kill joiner process mid-turn
3. Verify host detects timeout
4. Check max retries logic

### Calculation Discrepancy Test
Intentionally modify damage calculation on one side:
1. Change RNG seed
2. Observe RESOLUTION_REQUEST
3. Verify error handling

### Invalid Input Test
- [ ] Invalid Pokemon names
- [ ] Negative stat allocations
- [ ] Sum of boosts > 10
- [ ] Invalid move selection
- [ ] Malformed network messages

---

## Common Issues and Solutions

### Issue: "No hosts discovered"
**Solution:**
- Check both are on same network
- Verify firewall allows UDP 50000
- Try pinging host IP
- Check broadcast address (255.255.255.255)

### Issue: "Connection timeout"
**Solution:**
- Increase timeout in reliability.go
- Check network latency
- Verify ACK messages being sent

### Issue: "Pokemon not found"
**Solution:**
- Check CSV file exists in data/
- Verify Pokemon name spelling (case-sensitive)
- Check CSV parsing logs

### Issue: "Calculation mismatch"
**Solution:**
- Verify both using same seed
- Check RNG initialization
- Ensure same damage formula

---

## Debugging Tips

### Enable Verbose Logging
The `netio.VerboseEventLog` calls already exist. To see them:
1. Check netio package for log level settings
2. Add more logging in battle flow if needed

### Network Capture
Use Wireshark with filter:
```
udp.port == 50000
```

### Step-by-Step Debugging
Add breakpoints:
1. Message sending
2. Message receiving
3. Damage calculation
4. State transitions

---

## Test Data

### Recommended Pokemon for Testing

**High Attack (Physical):**
- Machamp (Attack: 130)
- Dragonite (Attack: 134)

**High Special Attack:**
- Alakazam (Sp. Attack: 135)
- Gengar (Sp. Attack: 130)

**Balanced:**
- Pikachu
- Charizard
- Blastoise

**Type Testing:**
- Fire: Charizard, Arcanine
- Water: Blastoise, Gyarados
- Grass: Venusaur, Exeggutor
- Electric: Pikachu, Zapdos
- Ground: Sandslash, Rhydon
- Ghost: Gengar

---

## Success Criteria

A successful implementation should:
1. ✅ Connect host and joiner reliably
2. ✅ Exchange messages without loss
3. ✅ Calculate damage identically on both sides
4. ✅ Apply type effectiveness correctly
5. ✅ Handle stat boosts properly
6. ✅ Alternate turns correctly
7. ✅ Detect battle end conditions
8. ✅ Recover from network delays
9. ✅ Handle invalid input gracefully
10. ✅ Display battle state clearly

---

## Reporting Issues

When reporting bugs, include:
1. Steps to reproduce
2. Expected behavior
3. Actual behavior
4. Network logs (if applicable)
5. Pokemon used
6. Seed value (from handshake)
7. Turn number when error occurred

---

## Next Steps After Testing

1. Document any bugs found
2. Add unit tests for core functions
3. Create integration test suite
4. Performance profiling
5. User acceptance testing
6. Load testing (multiple battles)

---

## Quick Test Script

```powershell
# Build both applications
go build -o host.exe .\host\host.go
go build -o joiner.exe .\joiner\joiner.go

# Run tests
go test ./...

# Start host in background
Start-Process .\host.exe

# Wait a moment
Start-Sleep -Seconds 2

# Start joiner in new window
Start-Process .\joiner.exe
```

---

This testing guide should help verify all aspects of the PokeProtocol implementation!
