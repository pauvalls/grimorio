// migrate-v1-to-v2 converts Grimorio campaigns from v1 format to v2 format.
// It creates canon.json and narrative_state.json for each existing campaign.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <campaigns-dir>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Example: %s ./campaigns\n", os.Args[0])
		os.Exit(1)
	}

	baseDir := os.Args[1]
	if err := migrate(baseDir); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration completed successfully.")
}

func migrate(baseDir string) error {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read campaigns directory: %w", err)
	}

	migrated := 0
	skipped := 0
	failed := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		campaignName := entry.Name()
		campaignDir := filepath.Join(baseDir, campaignName)

		// Check if already v2
		canonPath := filepath.Join(campaignDir, "canon.json")
		if _, err := os.Stat(canonPath); err == nil {
			fmt.Printf("  [SKIP] %s: already v2 (canon.json exists)\n", campaignName)
			skipped++
			continue
		}

		// Create backup
		backupDir := filepath.Join(baseDir, campaignName+".v1-backup")
		if err := backupCampaign(campaignDir, backupDir); err != nil {
			fmt.Printf("  [FAIL] %s: backup failed: %v\n", campaignName, err)
			failed++
			continue
		}

		// Migrate
		if err := migrateCampaign(campaignDir, campaignName); err != nil {
			fmt.Printf("  [FAIL] %s: migration failed: %v\n", campaignName, err)
			failed++
			continue
		}

		fmt.Printf("  [OK]   %s: migrated to v2\n", campaignName)
		migrated++
	}

	fmt.Printf("\nResults: %d migrated, %d skipped, %d failed\n", migrated, skipped, failed)
	return nil
}

func backupCampaign(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := backupCampaign(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func migrateCampaign(campaignDir, campaignName string) error {
	// Load campaign metadata
	campaignPath := filepath.Join(campaignDir, "campaign.json")
	var campaign domain.Campaign
	if data, err := os.ReadFile(campaignPath); err == nil {
		if err := json.Unmarshal(data, &campaign); err != nil {
			return fmt.Errorf("failed to unmarshal campaign JSON: %w", err)
		}
	}

	now := time.Now()

	// Create canon document
	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    campaignName,
		CreatedAt:     now,
		UpdatedAt:     now,
		Facts: []domain.CanonFact{
			{
				ID:        "fact-001",
				Category:  "lore",
				Statement: fmt.Sprintf("Campaign '%s' migrated from v1 to v2.", campaignName),
				Source:    "migration_v1_to_v2",
				Immutable: false,
				CreatedAt: now,
			},
		},
		Entities:   []domain.CanonEntity{},
		Timeline:   []domain.CanonTimelineEvent{},
		Rules:      []domain.CanonRule{},
		Relationships: []domain.CanonRelationship{},
	}

	// Try to extract entities from existing NPCs
	npcsDir := filepath.Join(campaignDir, "npcs")
	if entries, err := os.ReadDir(npcsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			// Extract name without extension
			if ext := filepath.Ext(name); ext == ".json" || ext == ".md" {
				entityName := name[:len(name)-len(ext)]
				canon.Entities = append(canon.Entities, domain.CanonEntity{
					ID:         fmt.Sprintf("npc-%s", entityName),
					Name:       entityName,
					Type:       domain.EntityTypeNPC,
					CanonState: domain.EntityStateAlive,
				})
			}
		}
	}

	// Save canon
	canonPath := filepath.Join(campaignDir, "canon.json")
	canonData, err := json.MarshalIndent(canon, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal canon: %w", err)
	}
	if err := os.WriteFile(canonPath, canonData, 0644); err != nil {
		return fmt.Errorf("failed to write canon.json: %w", err)
	}

	// Create narrative state
	state := &domain.NarrativeState{
		SchemaVersion:   domain.SchemaVersionV2,
		CampaignID:      campaignName,
		CurrentSession:  0,
		LastUpdated:     now,
		RevealedClues:   []domain.RevealedClue{},
		ActiveQuests:    []domain.QuestState{},
		CompletedQuests: []domain.QuestState{},
		FailedQuests:    []domain.QuestState{},
		DeadNPCs:        []domain.NPCDeathRecord{},
		KeyItems:        []domain.KeyItem{},
		SessionLog:      []domain.SessionRecord{},
		DMOverrides:     []domain.DMOverride{},
	}

	// Try to load existing quests into active quests
	questsDir := filepath.Join(campaignDir, "quests")
	if entries, err := os.ReadDir(questsDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if ext := filepath.Ext(name); ext == ".json" || ext == ".md" {
				questName := name[:len(name)-len(ext)]
				state.ActiveQuests = append(state.ActiveQuests, domain.QuestState{
					ID:        fmt.Sprintf("quest-%s", questName),
					Name:      questName,
					Status:    "active",
					SourceAct: "",
				})
			}
		}
	}

	// Save narrative state
	statePath := filepath.Join(campaignDir, "narrative_state.json")
	stateData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal narrative state: %w", err)
	}
	if err := os.WriteFile(statePath, stateData, 0644); err != nil {
		return fmt.Errorf("failed to write narrative_state.json: %w", err)
	}

	return nil
}
