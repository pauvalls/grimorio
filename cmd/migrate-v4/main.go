package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Migration script V2 → V3 for canon-session-context-refactor
// Adds chapter tracking and XP fields to narrative_state.json

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate_v4 <campaign_path>")
		fmt.Println("Example: migrate_v4 ~/campaigns/la-peste-de-velmorath")
		os.Exit(1)
	}

	campaignPath := os.Args[1]
	narrativeStatePath := filepath.Join(campaignPath, "canon", "narrative_state.json")

	fmt.Printf("🔧 Migrating %s to V3...\n", campaignPath)

	// Load narrative state
	data, err := os.ReadFile(narrativeStatePath)
	if err != nil {
		fmt.Printf("❌ Failed to read narrative_state.json: %v\n", err)
		os.Exit(1)
	}

	// Parse as map to preserve all fields
	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		fmt.Printf("❌ Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Add V3 fields if not present
	modified := false

	if _, exists := state["current_chapter_id"]; !exists {
		state["current_chapter_id"] = ""
		fmt.Println("  + Added current_chapter_id (empty)")
		modified = true
	}

	if _, exists := state["completed_chapters"]; !exists {
		state["completed_chapters"] = []interface{}{}
		fmt.Println("  + Added completed_chapters (empty)")
		modified = true
	}

	if _, exists := state["party_level"]; !exists {
		// Try to infer from xp_total if present
		if xpTotal, ok := state["xp_total"].(float64); ok {
			level := calculateLevelFromXP(int(xpTotal))
			state["party_level"] = level
			fmt.Printf("  + Added party_level: %d (inferred from xp_total: %.0f)\n", level, xpTotal)
		} else {
			state["party_level"] = 1
			fmt.Println("  + Added party_level: 1 (default)")
		}
		modified = true
	}

	if _, exists := state["xp_ledger"]; !exists {
		// Create ledger from session_log if available
		ledger := []interface{}{}
		if sessionLog, ok := state["session_log"].([]interface{}); ok {
			for _, session := range sessionLog {
				if sMap, ok := session.(map[string]interface{}); ok {
					if xpAwarded, ok := sMap["xp_awarded"].(float64); ok && xpAwarded > 0 {
						entry := map[string]interface{}{
							"session_num": sMap["session_num"],
							"amount":      xpAwarded,
							"reason":      "session",
							"timestamp":   time.Now().Unix(),
						}
						ledger = append(ledger, entry)
					}
				}
			}
			if len(ledger) > 0 {
				state["xp_ledger"] = ledger
				fmt.Printf("  + Added xp_ledger (%d entries from session_log)\n", len(ledger))
				modified = true
			} else {
				state["xp_ledger"] = []interface{}{}
				fmt.Println("  + Added xp_ledger (empty)")
				modified = true
			}
		} else {
			state["xp_ledger"] = []interface{}{}
			fmt.Println("  + Added xp_ledger (empty)")
			modified = true
		}
	}

	if !modified {
		fmt.Println("✅ Already up to date (V3)")
		return
	}

	// Write updated state
	output, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("❌ Failed to marshal JSON: %v\n", err)
		os.Exit(1)
	}

	// Backup original
	backupPath := narrativeStatePath + ".backup"
	if err := os.Rename(narrativeStatePath, backupPath); err != nil {
		fmt.Printf("❌ Failed to backup: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  💾 Backup: %s\n", backupPath)

	// Write new file
	if err := os.WriteFile(narrativeStatePath, output, 0644); err != nil {
		fmt.Printf("❌ Failed to write: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Migration complete! V3 fields added.\n")
	fmt.Println("\n📝 V3 fields added:")
	fmt.Println("  - current_chapter_id (string)")
	fmt.Println("  - completed_chapters ([]string)")
	fmt.Println("  - party_level (int, inferred from xp_total)")
	fmt.Println("  - xp_ledger ([]XPEntry, from session_log)")
}

// calculateLevelFromXP calculates D&D 5e party level from total XP
func calculateLevelFromXP(xpTotal int) int {
	thresholds := []int{
		0,      // Level 1
		300,    // Level 2
		900,    // Level 3
		2700,   // Level 4
		6500,   // Level 5
		14000,  // Level 6
		23000,  // Level 7
		34000,  // Level 8
		48000,  // Level 9
		64000,  // Level 10
		85000,  // Level 11
		100000, // Level 12
		120000, // Level 13
		140000, // Level 14
		165000, // Level 15
		195000, // Level 16
		225000, // Level 17
		265000, // Level 18
		305000, // Level 19
		355000, // Level 20
	}

	for i := len(thresholds) - 1; i >= 0; i-- {
		if xpTotal >= thresholds[i] {
			return i + 1
		}
	}

	return 1
}
