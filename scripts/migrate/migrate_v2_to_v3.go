package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MigrationConfig holds migration configuration.
type MigrationConfig struct {
	CampaignPath    string
	BackupPath      string
	DryRun          bool
	RollbackEnabled bool
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate_v2_to_v3 <campaign_path> [--dry-run]")
		os.Exit(1)
	}

	campaignPath := os.Args[1]
	dryRun := false
	for _, arg := range os.Args[2:] {
		if arg == "--dry-run" {
			dryRun = true
		}
	}

	config := MigrationConfig{
		CampaignPath:    campaignPath,
		BackupPath:      campaignPath + ".backup." + time.Now().Format("20060102_150405"),
		DryRun:          dryRun,
		RollbackEnabled: !dryRun,
	}

	if err := migrateCampaign(config); err != nil {
		fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Migration completed successfully!")
	if config.RollbackEnabled {
		fmt.Printf("Backup created at: %s\n", config.BackupPath)
		fmt.Println("To rollback: cp -r " + config.BackupPath + " " + config.CampaignPath)
	}
}

func migrateCampaign(config MigrationConfig) error {
	fmt.Printf("Starting migration from v2 to v3...\n")
	if config.DryRun {
		fmt.Println("[DRY RUN] No changes will be made")
	}

	// Step 1: Create backup
	if !config.DryRun {
		if err := createBackup(config); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Println("✓ Backup created")
	}

	// Step 2: Detect old area format
	oldAreas, err := detectOldAreaFormat(config.CampaignPath)
	if err != nil {
		return fmt.Errorf("failed to detect old areas: %w", err)
	}

	if len(oldAreas) == 0 {
		fmt.Println("✓ No old area format detected - campaign already migrated or new")
		return nil
	}

	fmt.Printf("✓ Found %d areas in old format\n", len(oldAreas))

	// Step 3: Convert to unified numbered format
	if !config.DryRun {
		if err := convertAreasToUnified(config.CampaignPath, oldAreas); err != nil {
			return fmt.Errorf("failed to convert areas: %w", err)
		}
		fmt.Println("✓ Areas converted to unified format")
	}

	// Step 4: Update version marker
	if !config.DryRun {
		if err := updateVersionMarker(config.CampaignPath, "3.0.0"); err != nil {
			return fmt.Errorf("failed to update version marker: %w", err)
		}
		fmt.Println("✓ Version marker updated to 3.0.0")
	}

	return nil
}

func createBackup(config MigrationConfig) error {
	// Create backup directory
	if err := os.MkdirAll(filepath.Dir(config.BackupPath), 0755); err != nil {
		return err
	}

	// Copy campaign directory
	return copyDir(config.CampaignPath, config.BackupPath)
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func detectOldAreaFormat(campaignPath string) ([]map[string]interface{}, error) {
	oldAreas := []map[string]interface{}{}

	// Look for canon.json or area files
	canonPath := filepath.Join(campaignPath, "canon.json")
	if _, err := os.Stat(canonPath); os.IsNotExist(err) {
		return oldAreas, nil
	}

	data, err := os.ReadFile(canonPath)
	if err != nil {
		return nil, err
	}

	var canon map[string]interface{}
	if err := json.Unmarshal(data, &canon); err != nil {
		return nil, err
	}

	// Check acts for old area format
	if acts, ok := canon["acts"].([]interface{}); ok {
		for _, act := range acts {
			actMap, ok := act.(map[string]interface{})
			if !ok {
				continue
			}

			if areas, ok := actMap["areas"].([]interface{}); ok {
				for _, area := range areas {
					areaMap, ok := area.(map[string]interface{})
					if !ok {
						continue
					}

					// Old format missing area_number field
					if _, hasNumber := areaMap["area_number"]; !hasNumber {
						oldAreas = append(oldAreas, areaMap)
					}
				}
			}
		}
	}

	return oldAreas, nil
}

func convertAreasToUnified(campaignPath string, oldAreas []map[string]interface{}) error {
	// Add area_number field to each area
	for i, area := range oldAreas {
		area["area_number"] = i + 1
	}

	// Write updated canon.json
	canonPath := filepath.Join(campaignPath, "canon.json")
	data, err := os.ReadFile(canonPath)
	if err != nil {
		return err
	}

	var canon map[string]interface{}
	if err := json.Unmarshal(data, &canon); err != nil {
		return err
	}

	// Update areas in canon
	if acts, ok := canon["acts"].([]interface{}); ok {
		areaIndex := 0
		for actIdx := range acts {
			actMap, ok := acts[actIdx].(map[string]interface{})
			if !ok {
				continue
			}

			if areas, ok := actMap["areas"].([]interface{}); ok {
				for areaIdx := range areas {
					if areaIndex < len(oldAreas) {
						areas[areaIdx] = oldAreas[areaIndex]
						areaIndex++
					}
				}
			}
		}
	}

	updatedData, err := json.MarshalIndent(canon, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(canonPath, updatedData, 0644)
}

func updateVersionMarker(campaignPath, version string) error {
	versionPath := filepath.Join(campaignPath, "VERSION")
	return os.WriteFile(versionPath, []byte(version), 0644)
}
