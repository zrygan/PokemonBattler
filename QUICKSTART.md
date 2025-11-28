# Quick Start Guide - Pokemon Battler

## Installation

No installation required! Just ensure you have Go installed (1.21+).

## First Battle (Local Testing)

### Step 1: Open Two Terminals

**Terminal 1** - Navigate to project:
```powershell
cd C:\Users\zrgnt\Documents\zrygan\PokemonBattler
```

**Terminal 2** - Navigate to project:
```powershell
cd C:\Users\zrgnt\Documents\zrygan\PokemonBattler
```

---

### Step 2: Start Host (Terminal 1)

```powershell
go run .\host\host.go
```

**Optional:** Enable verbose logging to see network events:
```powershell
go run .\host\host.go -verbose
```

You'll see:
```
Enter your name: _
```

Enter a name (e.g., "Alice") and press Enter.

The host will start listening and display:
```
Listening on [your-ip]:50000
```

---

### Step 3: Start Joiner (Terminal 2)

```powershell
go run .\joiner\joiner.go
```

**Optional:** Enable verbose logging to see network events:
```powershell
go run .\joiner\joiner.go -verbose
```

You'll see:
```
Enter your name: _
```

Enter a name (e.g., "Bob") and press Enter.

After 3 seconds, you'll see discovered hosts:
```
Discovered Hosts
    Alice @ 127.0.0.1 50000

Send an invite to..._
```

Type the host name ("Alice") and press Enter.

---

### Step 4: Host Accepts Connection

Back in **Terminal 1**, you'll see:
```
Match found, received a HANDSHAKE_REQUEST message from Bob
Accept this player? [Y:default / N]: _
```

Press Enter (or type "Y") to accept.

---

### Step 5: Host Selects Communication Mode

Host chooses mode:
```
Select a communication mode:
P: peer-to-peer
B: broadcast
_
```

Type "P" for peer-to-peer and press Enter.

---

### Step 6: Both Players Select Pokemon

**Both terminals will now show:**
```
Select a pokemon: _
```

Choose any Pokemon (examples):
- Pikachu
- Charizard
- Blastoise
- Venusaur
- Mewtwo
- Gengar

**Note:** Names are case-sensitive. Use proper capitalization (e.g., "Pikachu" not "pikachu").

---

### Step 7: Allocate Stat Boosts

Both players allocate 10 points between Special Attack and Special Defense:

```
You can allocate 10 points to your special attack and special defense, use it wisely.
Special attack allocation: _
```

Enter a number (e.g., "6") and press Enter.

```
Special defense allocation: _
```

Enter remaining points (e.g., "4") and press Enter.

**Rules:**
- Total must be ≤ 10
- Both must be ≥ 0

---

### Step 8: Battle Begins!

You'll see:
```
=== BATTLE START ===
Your Pokemon: Pikachu (HP: 35/35)

Available Moves:
1. Tackle (Power: 40, Type: normal, Category: physical)
2. Electric Attack (Power: 60, Type: electric, Category: special)
3. Special Blast (Power: 70, Type: electric, Category: special)

Special Attack Boosts: 6
Special Defense Boosts: 4

--- Turn 1 ---
Your turn!  (if you're the host)
```

or

```
--- Turn 1 ---
Opponent's turn... waiting...  (if you're the joiner)
```

---

### Step 9: Take Your Turn

When it's your turn:

```
Select a move (enter number) or 'chat <message>': _
```

Type "1", "2", or "3" and press Enter to select a move.

**OR** type `chat Hello!` to send a chat message to your opponent!

If the move is special and you have boosts:
```
Use a Special Attack boost? (y/n): _
```

Type "y" or "n" and press Enter.

---

### Chatting During Battle

You can chat with your opponent at any time during the battle!

**To send a message:**
```
Select a move (enter number) or 'chat <message>': chat Good luck!
```

**You'll see:**
```
💬 You: Good luck!
```

**Your opponent sees:**
```
💬 [YourName]: Good luck!
```

**Examples:**
- `chat Nice move!`
- `chat That was close!`
- `chat gg wp`

**Note:** Spectators can also see all chat messages!

---

### Step 10: Watch the Battle!

After each turn, you'll see:
- Damage dealt
- Current HP
- Turn switches between players

Example output:
```
Your Pokemon: Pikachu (HP: 28/35)

--- Turn 2 ---
Opponent's turn... waiting...
```

---

### Step 11: Battle Ends

When a Pokemon faints:
```
Your Pokemon fainted! You lose!

=== BATTLE END ===
```

or

```
Opponent's Pokemon fainted! You win!

=== BATTLE END ===
```

---

## Quick Reference

### Popular Pokemon to Try
- **Electric:** Pikachu, Raichu, Zapdos
- **Fire:** Charmander, Charmeleon, Charizard
- **Water:** Squirtle, Wartortle, Blastoise
- **Grass:** Bulbasaur, Ivysaur, Venusaur
- **Psychic:** Abra, Kadabra, Alakazam
- **Dragon:** Dratini, Dragonair, Dragonite

### Stat Boost Strategies
- **Aggressive:** 8 Special Attack / 2 Special Defense
- **Balanced:** 5 Special Attack / 5 Special Defense
- **Defensive:** 3 Special Attack / 7 Special Defense
- **All-In:** 10 Special Attack / 0 Special Defense

### Move Categories
- **Physical:** Uses Attack vs Defense
- **Special:** Uses Special Attack vs Special Defense
  - Can use stat boosts!

---

## Troubleshooting

### "Pokemon not found"
- Check spelling (case-sensitive!)
- Try: Pikachu, Charizard, Mewtwo, Gengar

### "No hosts discovered"
- Make sure host started first
- Wait full 3 seconds
- Check firewall (allow UDP port 50000)

### Connection hangs
- Press Ctrl+C and restart both
- Ensure both on same network

### Invalid stat allocation
- Sum must be ≤ 10
- Both must be ≥ 0
- Try: 5 and 5

---

## Network Battle (Two Computers)

### Computer 1 (Host):
1. Note your IP address:
   ```powershell
   ipconfig
   ```
   Look for "IPv4 Address" (e.g., 192.168.1.100)

2. Run host:
   ```powershell
   go run .\host\host.go
   ```

### Computer 2 (Joiner):
1. Make sure on same network as host

2. Run joiner:
   ```powershell
   go run .\joiner\joiner.go
   ```

3. Should discover host automatically!

### Firewall Note:
If joiner can't find host, allow UDP port 50000:
```powershell
New-NetFirewallRule -DisplayName "Pokemon Battle" -Direction Inbound -Protocol UDP -LocalPort 50000 -Action Allow
```

---

## Tips for Your First Battle

1. **Start Simple:**
   - Use Pikachu vs Charizard
   - Allocate 5/5 on boosts

2. **Understand Type Matchups:**
   - Water beats Fire
   - Fire beats Grass
   - Grass beats Water
   - Electric beats Water
   - Ground beats Electric

3. **Use Boosts Wisely:**
   - Save for powerful special moves
   - Use when opponent is low HP

4. **Watch Your HP:**
   - Displayed after each turn
   - Plan ahead!

---

## Next Steps

Once you're comfortable with basic battles:

1. Try different Pokemon combinations
2. Experiment with stat boost strategies
3. Test type effectiveness matchups
4. Battle with friends over network
5. Read docs/IMPLEMENTATION.md for advanced features

---

## Help

- **Documentation:** `docs/IMPLEMENTATION.md`
- **Testing Guide:** `docs/TESTING.md`
- **Full Summary:** `docs/SUMMARY.md`

---

## Verbose Mode

All applications (host, joiner, and spectator) support verbose logging to help debug network issues or understand the protocol.

### Enabling Verbose Mode

Add the `-verbose` flag when running:

```powershell
# Host with verbose logging
go run .\host\host.go -verbose

# Joiner with verbose logging
go run .\joiner\joiner.go -verbose

# Spectator with verbose logging
go run .\spectator\spectator.go -verbose
```

### What Verbose Mode Shows

With verbose mode enabled, you'll see detailed network events:

```
🔴 LOG :: Found a JOINER, received a MMB_JOINING message

🔴 LOG :: Found a JOINER, sent a MMB_HOSTING message
	> MessageParams:
		name: Alice
		port: 50000
		ip: 192.168.1.100

🔴 LOG :: Match found, received a HANDSHAKE_REQUEST message from Bob
	> MessageParams:
		name: Bob
	> MessageSender: 192.168.1.101:50001
```

### When to Use Verbose Mode

- **Debugging:** When connections aren't working properly
- **Learning:** To understand the protocol message flow
- **Development:** When testing new features
- **Network Issues:** To see exactly what packets are being sent/received

### Default Behavior

By default, verbose mode is **OFF**. This means:
- Clean output focused on battle events
- No network-level details displayed
- Better user experience for casual play

To see network details, explicitly enable it with the `-verbose` flag.

---

## Spectator Mode

Want to watch a battle without playing? Use spectator mode!

### Step 1: Start a Battle
First, have two players start a battle using the normal steps above (host + joiner).

### Step 2: Open a Third Terminal

```powershell
cd C:\Users\zrgnt\Documents\zrygan\PokemonBattler
```

### Step 3: Run Spectator

```powershell
go run .\spectator\spectator.go
```

**Optional:** Enable verbose logging to see network events:
```powershell
go run .\spectator\spectator.go -verbose
```

You'll see:
```
Welcome to PokeBattler - Spectator Mode
Enter your name: _
```

Enter a name and the spectator will discover active battles.

### Step 4: Select Battle to Watch

```
🔍 Searching for active battles...
Listening for 5 seconds...

📡 Found battle: Alice @ 192.168.1.100:50000

Available Battles:
1. Alice @ 192.168.1.100:50000

Select battle to spectate (enter number): _
```

Type "1" and press Enter.

### Step 5: Watch the Battle!

You'll see all battle events in real-time:
```
🎮 === SPECTATING BATTLE ===
You are now observing the battle. Press Ctrl+C to exit.

⚔️  BATTLE: Pikachu vs Charizard
   Pikachu: 35/35 HP
   Charizard: 78/78 HP

⚡ Attack announced: Thunderbolt

📊 Pikachu used Thunderbolt!
   Damage: 25
   Status: Pikachu used Thunderbolt! It was super effective!

   Current HP:
   Pikachu: 35/35
   Charizard: 53/78
```

**Note:** Spectators see all moves, damage calculations, and battle results but cannot interact with the battle.

---

## Have Fun! ⚡🔥💧🌿

Enjoy your Pokemon battles!
