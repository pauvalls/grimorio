package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pauvalls/grimorio/internal/validators"
)

// Valid chapter part names in order
var validPartNames = []string{"opener", "general-features", "npcs", "encounters", "areas-1", "areas-2", "closing"}

// partOrder maps part name to its sort order
var partOrder = map[string]int{
	"opener":           1,
	"general-features": 2,
	"npcs":             3,
	"encounters":       4,
	"areas-1":          5,
	"areas-2":          6,
	"closing":          7,
}

// ChapterPartResult is returned by SaveChapterPart
type ChapterPartResult struct {
	Status           string        `json:"status"`
	PartsSaved       int           `json:"parts_saved"`
	DraftPath        string        `json:"draft_path"`
	AccumulatedWords int           `json:"accumulated_words"`
	PartsReceived    []string      `json:"parts_received"`
	Context          *DraftContext `json:"context,omitempty"`
}

// DraftContext holds structured entity references extracted from accumulated draft content
type DraftContext struct {
	NPCs       []string `json:"npcs"`
	Areas      []string `json:"areas"`
	Encounters []string `json:"encounters"`
}

// FinalizeResult is returned by FinalizeChapter
type FinalizeResult struct {
	Status     string `json:"status"`
	Chapter    string `json:"chapter"`
	Areas      int    `json:"areas"`
	NPCs       int    `json:"npcs"`
	Encounters int    `json:"encounters"`
	WordCount  int    `json:"word_count"`
}

// SaveChapterPart saves a chapter part to the draft directory
func (s *CampaignService) SaveChapterPart(campaignID string, chapterNum int, partName, content string) (*ChapterPartResult, error) {
	if !s.campaignRepo.Exists(campaignID) {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}

	// Validate part name
	order, ok := partOrder[partName]
	if !ok {
		return nil, fmt.Errorf("invalid part name %q, valid names: %s", partName, strings.Join(validPartNames, ", "))
	}

	// Create draft directory
	draftDir := filepath.Join(s.baseDir, campaignID, "chapters", ".draft", fmt.Sprintf("chapter_%02d", chapterNum))
	if err := os.MkdirAll(draftDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create draft directory: %w", err)
	}

	// Write part file: {order}-{partName}.md
	filename := fmt.Sprintf("%02d-%s.md", order, partName)
	partPath := filepath.Join(draftDir, filename)
	if err := os.WriteFile(partPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write part file: %w", err)
	}

	// Calculate accumulated stats
	partsSaved, partsList, totalWords, err := s.draftStats(draftDir)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate draft stats: %w", err)
	}

	// Extract structured context from accumulated draft content
	accumulated, _, err := s.assembleDraft(draftDir)
	if err != nil {
		accumulated = ""
	}
	ctx := extractDraftContext(accumulated)

	return &ChapterPartResult{
		Status:           "ok",
		PartsSaved:       partsSaved,
		DraftPath:        draftDir,
		AccumulatedWords: totalWords,
		PartsReceived:    partsList,
		Context:          ctx,
	}, nil
}

// FinalizeChapter assembles draft parts into a final chapter, validates, and saves
func (s *CampaignService) FinalizeChapter(campaignID string, chapterNum int, title string) (*FinalizeResult, error) {
	if !s.campaignRepo.Exists(campaignID) {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}

	draftDir := filepath.Join(s.baseDir, campaignID, "chapters", ".draft", fmt.Sprintf("chapter_%02d", chapterNum))

	// Check draft directory exists
	if _, err := os.Stat(draftDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("no draft found for chapter %d — call save_chapter_part first", chapterNum)
	}

	// Read all part files in order
	content, partsList, err := s.assembleDraft(draftDir)
	if err != nil {
		return nil, fmt.Errorf("failed to assemble draft: %w", err)
	}

	if len(partsList) == 0 {
		return nil, fmt.Errorf("draft directory is empty — no parts saved")
	}

	// Prepend chapter title if not already present
	if !strings.HasPrefix(strings.TrimSpace(content), "# ") {
		chapterHeader := fmt.Sprintf("# Chapter %d: %s\n\n", chapterNum, title)
		content = chapterHeader + content
	}

	// Run validation
	wordCount := validators.CountWords(content)
	areaResult := validators.ValidateAreaMarkdown(content)
	if !areaResult.Valid {
		return nil, fmt.Errorf("chapter validation failed: %v", validationErrors(areaResult.Errors))
	}

	// Parse entities
	parser := NewEntityParser()
	result, err := parser.ParseChapter(content, campaignID, chapterNum)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chapter: %w", err)
	}

	// Write final chapter file
	dir := filepath.Join(s.baseDir, campaignID, "chapters")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chapters directory: %w", err)
	}

	filename := fmt.Sprintf("chapter_%02d.md", chapterNum)
	finalPath := filepath.Join(dir, filename)
	tempPath := filepath.Join(dir, "."+filename+".tmp")

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write chapter: %w", err)
	}

	// Sync inline NPCs to canon
	if len(result.NPCs) > 0 {
		if err := s.syncCanonEntities(campaignID, result.NPCs, nil, nil); err != nil {
			_ = os.Remove(tempPath)
			return nil, fmt.Errorf("failed to sync canon: %w", err)
		}
	}

	// Atomic rename
	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("failed to finalize chapter: %w", err)
	}

	// Clean up draft directory
	if err := os.RemoveAll(draftDir); err != nil {
		// Non-fatal: chapter is saved, just log warning
		_ = err
	}

	// Advisory CR audit on the chapter content. The hook is
	// strictly advisory: it never returns an error and never
	// blocks the finalize. It is also nil-tolerant so legacy
	// callers that don't wire the hook keep working.
	s.invokeChapterAudit(context.Background(), content, campaignID)

	return &FinalizeResult{
		Status:     "ok",
		Chapter:    filename,
		Areas:      len(result.Areas),
		NPCs:       len(result.NPCs),
		Encounters: len(result.Encounters),
		WordCount:  wordCount,
	}, nil
}

// draftStats returns the number of parts, their names, and total word count
func (s *CampaignService) draftStats(draftDir string) (int, []string, int, error) {
	entries, err := os.ReadDir(draftDir)
	if err != nil {
		return 0, nil, 0, fmt.Errorf("failed to read draft directory: %w", err)
	}

	var parts []string
	totalWords := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(draftDir, entry.Name()))
		if err != nil {
			continue
		}
		totalWords += validators.CountWords(string(data))
		// Extract part name from filename: "01-opener.md" → "opener"
		name := strings.TrimSuffix(entry.Name(), ".md")
		if idx := strings.Index(name, "-"); idx >= 0 {
			name = name[idx+1:]
		}
		parts = append(parts, name)
	}

	sort.Strings(parts)
	return len(parts), parts, totalWords, nil
}

// assembleDraft reads all part files in order and concatenates them
func (s *CampaignService) assembleDraft(draftDir string) (string, []string, error) {
	entries, err := os.ReadDir(draftDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read draft directory: %w", err)
	}

	// Sort entries by name (which includes order prefix)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var parts []string
	var sb strings.Builder

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(draftDir, entry.Name()))
		if err != nil {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.Write(data)

		name := strings.TrimSuffix(entry.Name(), ".md")
		if idx := strings.Index(name, "-"); idx >= 0 {
			name = name[idx+1:]
		}
		parts = append(parts, name)
	}

	return sb.String(), parts, nil
}

// validationErrors converts validator errors to a readable string
func validationErrors(errs []validators.ValidationError) string {
	var msgs []string
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("[%s] %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

var (
	draftAreaHeadingRe = regexp.MustCompile(`(?i)^#{3}\s+(?:Área|Area)\s+(\d+|[A-Z]\d*)[:\s]*(.+?)\s*$`)
	draftEncounterRe   = regexp.MustCompile(`(?i)^#{3}\s+(?:Encuentro|Encounter)\s+(\d+)[:\s]*(.+?)\s*$`)
)

// extractDraftContext parses accumulated draft content to extract NPC names, area numbers, and encounter names
func extractDraftContext(content string) *DraftContext {
	ctx := &DraftContext{
		NPCs:       []string{},
		Areas:      []string{},
		Encounters: []string{},
	}

	if content == "" {
		return ctx
	}

	lines := strings.Split(content, "\n")
	inNPCSection := false
	inEncounterSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "## ") {
			inNPCSection = false
			inEncounterSection = false

			if strings.Contains(lower, "npc") || strings.Contains(lower, "personaje") {
				inNPCSection = true
			} else if strings.Contains(lower, "encuentro") || strings.Contains(lower, "encounter") {
				inEncounterSection = true
			}
			continue
		}

		if matches := draftAreaHeadingRe.FindStringSubmatch(trimmed); matches != nil {
			ctx.Areas = append(ctx.Areas, matches[1])
			continue
		}

		if inEncounterSection {
			if matches := draftEncounterRe.FindStringSubmatch(trimmed); matches != nil {
				ctx.Encounters = append(ctx.Encounters, strings.TrimSpace(matches[2]))
			}
			continue
		}

		if inNPCSection {
			if strings.HasPrefix(trimmed, "### ") {
				name := strings.TrimPrefix(trimmed, "### ")
				name = strings.TrimSpace(name)
				if name != "" {
					ctx.NPCs = append(ctx.NPCs, name)
				}
			}
		}
	}

	return ctx
}
