package poke

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TeamManager handles Pokemon profile creation and management
type TeamManager struct {
	TrainerName string
}

// NewTeamManager creates a new team manager for a trainer
func NewTeamManager(trainerName string) *TeamManager {
	return &TeamManager{
		TrainerName: trainerName,
	}
}

// CustomizePokemon allows a player to customize their Pokemon before battle
func (tm *TeamManager) CustomizePokemon(pokemon *Pokemon) (*PokemonProfile, error) {
	reader := bufio.NewReader(os.Stdin)

	// Try to load existing profile
	profile, err := LoadProfile(tm.TrainerName, pokemon.Name)
	if err == nil {
		fmt.Printf("\n🎉 Welcome back! Found existing profile for %s!\n", pokemon.Name)
		profile.DisplayProfile()

		fmt.Print("\nWould you like to use this profile? (y/n): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			return profile, nil
		}
	}

	// Create new profile
	profile = NewPokemonProfile(pokemon.Name)

	fmt.Printf("\n✨ Let's customize your %s! ✨\n", pokemon.Name)

	// Set nickname
	fmt.Print("\nGive your Pokemon a nickname (or press Enter to skip): ")
	nickname, _ := reader.ReadString('\n')
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		profile.Nickname = nickname
		fmt.Printf("Great! Your %s will be called \"%s\"!\n", pokemon.Name, nickname)
	}

	// Choose personality
	fmt.Println("\n🎭 Choose a personality for your Pokemon:")
	for i, personality := range AllPersonalities {
		fmt.Printf("%d. %s\n", i+1, personality)
	}

	for {
		fmt.Print("\nEnter personality number (1-8): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err == nil && choice >= 1 && choice <= len(AllPersonalities) {
			profile.Personality = AllPersonalities[choice-1]
			fmt.Printf("Perfect! Your Pokemon has a %s personality!\n", profile.Personality)
			break
		}
		fmt.Println("Invalid choice. Please enter a number between 1 and 8.")
	}

	// Show flavor text
	fmt.Printf("\n%s\n", profile.GetFlavorText("battle_start"))

	// Save profile
	if err := profile.SaveProfile(tm.TrainerName); err != nil {
		fmt.Printf("Warning: Could not save profile: %v\n", err)
	} else {
		fmt.Println("\n✓ Profile saved successfully!")
	}

	return profile, nil
}

// UpdateProfileAfterBattle updates the profile after a battle concludes
func (tm *TeamManager) UpdateProfileAfterBattle(profile *PokemonProfile, won bool) error {
	if won {
		profile.RecordVictory()
		fmt.Printf("\n%s\n", profile.GetFlavorText("victory"))
		fmt.Printf("🎊 %s gained friendship! (+5)\n", profile.GetDisplayName())
	} else {
		profile.RecordDefeat()
		fmt.Printf("\n%s tried their best...\n", profile.GetDisplayName())
		fmt.Printf("💙 %s gained friendship! (+2)\n", profile.GetDisplayName())
	}

	// Show updated stats
	profile.DisplayProfile()

	// Save updated profile
	return profile.SaveProfile(tm.TrainerName)
}

// ListProfiles lists all saved Pokemon profiles for the trainer
func (tm *TeamManager) ListProfiles() error {
	profileDir := "./profiles"

	// Check if directory exists
	if _, err := os.Stat(profileDir); os.IsNotExist(err) {
		fmt.Println("\nNo Pokemon profiles found yet. Customize a Pokemon to create your first profile!")
		return nil
	}

	// Read directory
	entries, err := os.ReadDir(profileDir)
	if err != nil {
		return fmt.Errorf("failed to read profiles directory: %w", err)
	}

	// Filter for this trainer's profiles
	trainerProfiles := []string{}
	prefix := tm.TrainerName + "_"

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) && strings.HasSuffix(entry.Name(), ".json") {
			trainerProfiles = append(trainerProfiles, entry.Name())
		}
	}

	if len(trainerProfiles) == 0 {
		fmt.Println("\nNo Pokemon profiles found for this trainer.")
		return nil
	}

	fmt.Printf("\n📚 Your Pokemon Profiles:\n")
	fmt.Println("=========================")

	for _, filename := range trainerProfiles {
		// Extract Pokemon name from filename
		parts := strings.TrimSuffix(filename, ".json")
		parts = strings.TrimPrefix(parts, prefix)

		// Load and display profile
		profile, err := LoadProfile(tm.TrainerName, parts)
		if err == nil {
			profile.DisplayProfile()
		}
	}

	return nil
}

// ShowPreBattleMessage displays a personality-based pre-battle message
func ShowPreBattleMessage(profile *PokemonProfile) {
	if profile == nil {
		return
	}
	fmt.Printf("\n%s\n", profile.GetFlavorText("battle_start"))
}

// ShowLowHealthMessage displays a personality-based low health warning
func ShowLowHealthMessage(profile *PokemonProfile) {
	if profile == nil {
		return
	}
	fmt.Printf("\n⚠️  %s\n", profile.GetFlavorText("low_hp"))
}

// ShowCriticalHitMessage displays a personality-based critical hit message
func ShowCriticalHitMessage(profile *PokemonProfile) {
	if profile == nil {
		return
	}
	fmt.Printf("\n💥 %s\n", profile.GetFlavorText("critical_hit"))
}
