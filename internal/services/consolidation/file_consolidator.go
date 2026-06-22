package consolidation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileConsolidator detects duplicate, obsolete, and stale generated files.
type FileConsolidator struct {
	generatedNames []string
}

// NewFileConsolidator creates a new file consolidator.
func NewFileConsolidator() *FileConsolidator {
	return &FileConsolidator{
		generatedNames: []string{"campaign.md", "INDEX.md", "index.md"},
	}
}

// Name returns the analyzer name.
func (f *FileConsolidator) Name() string {
	return "file_consolidator"
}

// Analyze detects duplicate files and stale generated artifacts.
func (f *FileConsolidator) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	result := &AnalysisResult{Rule: "file_integrity", Passed: true}

	// Duplicate detection by content hash.
	hashes := make(map[string][]CampaignFile)
	for _, file := range files {
		h := sha256.Sum256([]byte(file.Content))
		key := hex.EncodeToString(h[:])
		hashes[key] = append(hashes[key], file)
	}

	for _, group := range hashes {
		if len(group) <= 1 {
			continue
		}
		result.Passed = false
		result.Severity = "warning"
		var paths []string
		for _, f := range group {
			paths = append(paths, f.RelPath)
		}
		sort.Strings(paths)
		result.Issues = append(result.Issues, DomainIssue{
			Rule:      "duplicate_file",
			Severity:  "warning",
			Message:   fmt.Sprintf("Duplicate file content found in %s", strings.Join(paths, ", ")),
			Locations: paths,
			Suggestion: "Remove exact duplicate files, keeping the canonical copy.",
		})
		result.Fixes = append(result.Fixes, domainFix{
			Rule:      "duplicate_file",
			Target:    paths[0],
			Before:    strings.Join(paths, ", "),
			After:     paths[0],
			Locations: paths,
		})
	}

	// Stale generated files.
	var sourceNewest time.Time
	for _, file := range files {
		if f.isGenerated(file.RelPath) {
			continue
		}
		if file.ModTime.After(sourceNewest) {
			sourceNewest = file.ModTime
		}
	}

	for _, file := range files {
		if !f.isGenerated(file.RelPath) {
			continue
		}
		if !sourceNewest.IsZero() && file.ModTime.Before(sourceNewest) {
			result.Passed = false
			if result.Severity == "" {
				result.Severity = "warning"
			}
			result.Issues = append(result.Issues, DomainIssue{
				Rule:      "stale_generated_file",
				Severity:  "warning",
				Message:   fmt.Sprintf("Generated file '%s' is older than the newest source (%s)", file.RelPath, sourceNewest.Format(time.RFC3339)),
				Locations: []string{file.RelPath},
				Suggestion: "Regenerate the file to reflect source changes.",
			})
		}
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d file integrity issue(s)", len(result.Issues))
	}
	return result, nil
}

// StaleFiles returns generated files older than the newest source.
func (f *FileConsolidator) StaleFiles(files []CampaignFile) []CampaignFile {
	var sourceNewest time.Time
	for _, file := range files {
		if f.isGenerated(file.RelPath) {
			continue
		}
		if file.ModTime.After(sourceNewest) {
			sourceNewest = file.ModTime
		}
	}

	var stale []CampaignFile
	for _, file := range files {
		if !f.isGenerated(file.RelPath) {
			continue
		}
		if !sourceNewest.IsZero() && file.ModTime.Before(sourceNewest) {
			stale = append(stale, file)
		}
	}
	return stale
}

func (f *FileConsolidator) isGenerated(relPath string) bool {
	base := filepath.Base(relPath)
	for _, g := range f.generatedNames {
		if strings.EqualFold(base, g) {
			return true
		}
	}
	return false
}

// RemoveDuplicateFiles deletes exact duplicate files, keeping the first path.
func RemoveDuplicateFiles(files []CampaignFile) ([]string, error) {
	hashes := make(map[string][]CampaignFile)
	for _, file := range files {
		h := sha256.Sum256([]byte(file.Content))
		key := hex.EncodeToString(h[:])
		hashes[key] = append(hashes[key], file)
	}

	var removed []string
	for _, group := range hashes {
		if len(group) <= 1 {
			continue
		}
		for i := 1; i < len(group); i++ {
			if err := os.Remove(group[i].Path); err != nil && !os.IsNotExist(err) {
				return removed, fmt.Errorf("failed to remove duplicate %s: %w", group[i].Path, err)
			}
			removed = append(removed, group[i].RelPath)
		}
	}
	return removed, nil
}
