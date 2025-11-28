package poke

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Personality traits that affect flavor text during battle
type Personality string

const (
	PersonalityBrave   Personality = "Brave"
	PersonalityTimid   Personality = "Timid"
	PersonalityJolly   Personality = "Jolly"
	PersonalitySerious Personality = "Serious"
	PersonalitySassy   Personality = "Sassy"
	PersonalityCalm    Personality = "Calm"
	PersonalityPlayful Personality = "Playful"
	PersonalityProud   Personality = "Proud"
)

var AllPersonalities = []Personality{
	PersonalityBrave, PersonalityTimid, PersonalityJolly, PersonalitySerious,
	PersonalitySassy, PersonalityCalm, PersonalityPlayful, PersonalityProud,
}

// PokemonProfile stores nickname, personality, and friendship data
type PokemonProfile struct {
	OriginalName string      `json:"original_name"`
	Nickname     string      `json:"nickname"`
	Personality  Personality `json:"personality"`
	Friendship   int         `json:"friendship"`    // 0-100
	Victories    int         `json:"victories"`     // Number of battle wins
	TotalBattles int         `json:"total_battles"` // Total battles participated
}

// GetDisplayName returns the nickname if set, otherwise the original name
func (p *PokemonProfile) GetDisplayName() string {
	if p.Nickname != "" {
		return p.Nickname
	}
	return p.OriginalName
}

// IncreaseFriendship increases friendship level (max 100)
func (p *PokemonProfile) IncreaseFriendship(amount int) {
	p.Friendship += amount
	if p.Friendship > 100 {
		p.Friendship = 100
	}
}

// DecreaseFriendship decreases friendship level (min 0)
func (p *PokemonProfile) DecreaseFriendship(amount int) {
	p.Friendship -= amount
	if p.Friendship < 0 {
		p.Friendship = 0
	}
}

// RecordVictory records a battle victory
func (p *PokemonProfile) RecordVictory() {
	p.Victories++
	p.TotalBattles++
	p.IncreaseFriendship(5) // Gain 5 friendship per victory
}

// RecordDefeat records a battle loss
func (p *PokemonProfile) RecordDefeat() {
	p.TotalBattles++
	p.IncreaseFriendship(2) // Gain 2 friendship even in defeat
}

// GetFlavorText returns personality-specific flavor text for different battle events
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
		case PersonalitySerious:
			return fmt.Sprintf("%s focuses intently on the opponent.", displayName)
		case PersonalitySassy:
			return fmt.Sprintf("%s struts confidently into battle!", displayName)
		case PersonalityCalm:
			return fmt.Sprintf("%s enters peacefully, unfazed by the situation.", displayName)
		case PersonalityPlayful:
			return fmt.Sprintf("%s playfully prances around, ready for action!", displayName)
		case PersonalityProud:
			return fmt.Sprintf("%s holds its head high with pride!", displayName)
		}
	case "critical_hit":
		switch p.Personality {
		case PersonalityBrave:
			return fmt.Sprintf("%s strikes with courage!", displayName)
		case PersonalityJolly:
			return fmt.Sprintf("%s landed a lucky hit! It seems pleased!", displayName)
		case PersonalitySassy:
			return fmt.Sprintf("%s delivers a devastating blow with style!", displayName)
		case PersonalityProud:
			return fmt.Sprintf("%s shows off its superior skill!", displayName)
		default:
			return fmt.Sprintf("%s lands a powerful strike!", displayName)
		}
	case "low_hp":
		switch p.Personality {
		case PersonalityBrave:
			return fmt.Sprintf("%s refuses to give up!", displayName)
		case PersonalityTimid:
			return fmt.Sprintf("%s looks frightened and unsteady...", displayName)
		case PersonalitySerious:
			return fmt.Sprintf("%s analyzes the situation carefully.", displayName)
		case PersonalityCalm:
			return fmt.Sprintf("%s remains composed despite the danger.", displayName)
		default:
			return fmt.Sprintf("%s is hanging in there!", displayName)
		}
	case "victory":
		switch p.Personality {
		case PersonalityBrave:
			return fmt.Sprintf("%s stands victorious!", displayName)
		case PersonalityJolly:
			return fmt.Sprintf("%s jumps for joy!", displayName)
		case PersonalitySassy:
			return fmt.Sprintf("%s strikes a victory pose!", displayName)
		case PersonalityProud:
			return fmt.Sprintf("%s basks in its glorious victory!", displayName)
		default:
			return fmt.Sprintf("%s won the battle!", displayName)
		}
	}

	return ""
}

// SaveProfile saves the Pokemon profile to a JSON file
func (p *PokemonProfile) SaveProfile(trainerName string) error {
	// Create profiles directory if it doesn't exist
	profileDir := filepath.Join(".", "profiles")
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Create filename based on trainer and pokemon name
	filename := filepath.Join(profileDir, fmt.Sprintf("%s_%s.json", trainerName, p.OriginalName))

	// Marshal to JSON
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	return nil
}

// LoadProfile loads a Pokemon profile from a JSON file
func LoadProfile(trainerName string, pokemonName string) (*PokemonProfile, error) {
	filename := filepath.Join(".", "profiles", fmt.Sprintf("%s_%s.json", trainerName, pokemonName))

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Unmarshal JSON
	var profile PokemonProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

// NewPokemonProfile creates a new Pokemon profile with default values
func NewPokemonProfile(originalName string) *PokemonProfile {
	return &PokemonProfile{
		OriginalName: originalName,
		Nickname:     "",
		Personality:  PersonalitySerious, // Default personality
		Friendship:   50,                 // Start at 50 friendship
		Victories:    0,
		TotalBattles: 0,
	}
}

// GetFriendshipLevel returns a descriptive friendship level
func (p *PokemonProfile) GetFriendshipLevel() string {
	switch {
	case p.Friendship >= 90:
		return "Best Friends"
	case p.Friendship >= 70:
		return "Great Friends"
	case p.Friendship >= 50:
		return "Good Friends"
	case p.Friendship >= 30:
		return "Friends"
	case p.Friendship >= 10:
		return "Acquaintances"
	default:
		return "Strangers"
	}
}

// DisplayProfile shows a formatted profile summary
func (p *PokemonProfile) DisplayProfile() {
	displayName := p.GetDisplayName()
	if p.Nickname != "" {
		fmt.Printf("\n=== %s (%s) ===\n", displayName, p.OriginalName)
	} else {
		fmt.Printf("\n=== %s ===\n", displayName)
	}
	fmt.Printf("Personality: %s\n", p.Personality)
	fmt.Printf("Friendship: %d/100 (%s)\n", p.Friendship, p.GetFriendshipLevel())
	fmt.Printf("Battle Record: %d-%d (%.1f%% win rate)\n",
		p.Victories,
		p.TotalBattles-p.Victories,
		float64(p.Victories)/float64(max(p.TotalBattles, 1))*100)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
