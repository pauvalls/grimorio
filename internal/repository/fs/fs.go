// Package fs provides filesystem-backed V3 repository implementations.
package fs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
)

// readJSON reads and unmarshals a JSON file at the given path.
func readJSON(path string, v any) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	if err := json.Unmarshal(bytes, v); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}
	return nil
}

// writeJSON marshals and writes v to path as indented JSON.
func writeJSON(path string, v any) error {
	bytes, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}

// deleteFile removes the file at path. Returns nil if the file does not exist.
func deleteFile(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// campaignDir returns the directory for a campaign.
func campaignDir(baseDir, campaignID string) string {
	return filepath.Join(baseDir, campaignID)
}

// ensureSubdir ensures a subdirectory exists under the campaign directory.
func ensureSubdir(baseDir, campaignID, subdir string) (string, error) {
	dir := filepath.Join(campaignDir(baseDir, campaignID), subdir)
	return dir, os.MkdirAll(dir, 0755)
}

// entityPath returns the path for an entity JSON file.
func entityPath(baseDir, campaignID, subdir, id string) string {
	filename := domain.SanitizeFilename(id) + ".json"
	return filepath.Join(campaignDir(baseDir, campaignID), subdir, filename)
}

// listJSONFiles reads all JSON files in a directory and unmarshals them.
func listJSONFiles[T any](dir string) ([]*T, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*T{}, nil
		}
		return nil, err
	}

	var result []*T
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var item T
		path := filepath.Join(dir, entry.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(bytes, &item); err != nil {
			continue
		}
		result = append(result, &item)
	}
	return result, nil
}
