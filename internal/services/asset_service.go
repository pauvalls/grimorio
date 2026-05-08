package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/svg"
	"golang.org/x/time/rate"
)

// campaignLimiters holds a per-campaign rate limiter for image generation.
// This prevents one campaign from starving others while still respecting
// API rate limits per campaign.
var (
	campaignLimiters   = make(map[string]*rate.Limiter)
	campaignLimitersMu sync.RWMutex
)

// getCampaignLimiter returns (or creates) a rate limiter for the given campaign.
// The limiter allows 1 request every 3 seconds with a burst of 1.
func getCampaignLimiter(campaign string) *rate.Limiter {
	campaignLimitersMu.RLock()
	lim, ok := campaignLimiters[campaign]
	campaignLimitersMu.RUnlock()
	if ok {
		return lim
	}

	campaignLimitersMu.Lock()
	defer campaignLimitersMu.Unlock()
	// Double-check after acquiring write lock.
	if lim, ok := campaignLimiters[campaign]; ok {
		return lim
	}
	lim = rate.NewLimiter(rate.Every(3*time.Second), 1)
	campaignLimiters[campaign] = lim
	return lim
}

// AssetService handles asset generation (maps, images, dividers)
type AssetService struct {
	baseDir       string
	imgConfig     image.Config
	imageProvider image.Provider // injectable for testing
}

// NewAssetService creates a new asset service
func NewAssetService(baseDir string, imgConfig image.Config) *AssetService {
	return &AssetService{
		baseDir:   baseDir,
		imgConfig: imgConfig,
	}
}

// NewAssetServiceWithProvider creates a new asset service with a custom provider (for testing)
func NewAssetServiceWithProvider(baseDir string, provider image.Provider) *AssetService {
	return &AssetService{
		baseDir:       baseDir,
		imageProvider: provider,
	}
}

func (s *AssetService) campaignDir(campaign string) string {
	return filepath.Join(s.baseDir, campaign)
}

func (s *AssetService) assetsDir(campaign string) string {
	return filepath.Join(s.campaignDir(campaign), "assets")
}

// GenerateMap generates a procedural SVG battle map
func (s *AssetService) GenerateMap(campaign, filename, style, title string, rooms int, labels []string) (string, error) {
	svgCfg := svg.DefaultBattleMapConfig()
	if style != "" {
		svgCfg.Style = svg.MapStyle(style)
	}
	if title != "" {
		svgCfg.Title = title
	}
	svgCfg.NumRooms = rooms
	if len(labels) > 0 {
		svgCfg.Labels = labels
	}

	svgContent := svg.GenerateBattleMap(svgCfg)
	assetsDir := s.assetsDir(campaign)
	outputPath := filepath.Join(assetsDir, ensureSVGExt(filename))
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create assets dir: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write map to %s: %w", outputPath, err)
	}
	return outputPath, nil
}

// GenerateDivider generates a decorative SVG divider
func (s *AssetService) GenerateDivider(campaign, filename, style string, width int) (string, error) {
	if style == "" {
		style = "ornate"
	}

	svgContent := svg.GenerateDivider(width, style)
	assetsDir := s.assetsDir(campaign)
	outputPath := filepath.Join(assetsDir, ensureSVGExt(filename))
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create assets dir: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write divider to %s: %w", outputPath, err)
	}
	return outputPath, nil
}

// GenerateImage generates an image using AI with fallback providers.
// Rate limiting is per-campaign so different campaigns can generate in parallel.
func (s *AssetService) GenerateImage(campaign, filename, prompt, imgType string) (string, error) {
	lim := getCampaignLimiter(campaign)
	if err := lim.Wait(context.Background()); err != nil {
		return "", fmt.Errorf("rate limiter error: %w", err)
	}

	providers := s.getProviders()

	outputPath := filepath.Join(s.assetsDir(campaign), ensureImageExt(filename))
	_, err := image.GenerateAndSaveWithFallback(providers, prompt, outputPath)
	if err != nil {
		return "", fmt.Errorf("image generation failed (tried %d providers): %w", len(providers), err)
	}

	return outputPath, nil
}

// getProviders returns the list of providers to try
// If a custom provider is injected (for testing), use only that
// Otherwise use the fallback chain: primary -> raphael -> pollinations
func (s *AssetService) getProviders() []image.Provider {
	if s.imageProvider != nil {
		return []image.Provider{s.imageProvider}
	}
	return image.NewProviderChain(s.imgConfig)
}

// InsertImageReference inserts an image reference into a markdown file.
// It finds the section by heading and appends the image reference after the section content.
// If section is empty, appends to the end of the file.
func (s *AssetService) InsertImageReference(campaign, markdownFile, section, alt, filename string) error {
	if markdownFile == "" {
		return nil // No markdown file specified, skip
	}

	// Resolve markdown path
	mdPath := markdownFile
	if !filepath.IsAbs(mdPath) {
		mdPath = filepath.Join(s.campaignDir(campaign), markdownFile)
	}

	// Read markdown content
	content, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("failed to read markdown file %s: %w", mdPath, err)
	}

	// Build image reference
	imgRef := fmt.Sprintf("\n![%s](assets/%s)\n", alt, ensureImageExt(filename))

	var newContent string
	if section == "" {
		// Append to end of file
		newContent = string(content) + imgRef
	} else {
		// Find the section and insert after it
		newContent = insertAfterSection(string(content), section, imgRef)
	}

	if err := os.WriteFile(mdPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file %s: %w", mdPath, err)
	}

	return nil
}

// insertAfterSection finds a section heading in markdown and inserts text after its content.
// It looks for the next heading at the same or higher level to determine section boundaries.
func insertAfterSection(content, section, insertion string) string {
	lines := strings.Split(content, "\n")
	var result []string
	found := false
	inserted := false
	sectionLevel := 0

	for _, line := range lines {
		if !found {
			// Check if this line is the section heading
			level := headingLevel(line)
			if level > 0 {
				headingText := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "#"), "#"))
				headingText = strings.TrimSpace(headingText)
				if strings.EqualFold(headingText, section) {
					found = true
					sectionLevel = level
					result = append(result, line)
					continue
				}
			}
			result = append(result, line)
		} else if !inserted {
			// Check if we've reached the next section at same or higher level
			level := headingLevel(line)
			if level > 0 && level <= sectionLevel {
				// Insert before this new section
				result = append(result, insertion)
				inserted = true
			}
			result = append(result, line)
		} else {
			result = append(result, line)
		}
	}

	// If section was found but never inserted (end of file)
	if found && !inserted {
		result = append(result, insertion)
	}

	return strings.Join(result, "\n")
}

// headingLevel returns the heading level (1-6) or 0 if not a heading
func headingLevel(line string) int {
	line = strings.TrimSpace(line)
	for i := 1; i <= 6; i++ {
		prefix := strings.Repeat("#", i) + " "
		if strings.HasPrefix(line, prefix) {
			return i
		}
	}
	return 0
}

// ensureSVGExt ensures filename has a .svg extension
func ensureSVGExt(filename string) string {
	if strings.ToLower(filepath.Ext(filename)) == ".svg" {
		return filename
	}
	return filename + ".svg"
}

// ensureImageExt ensures filename has an image extension, defaulting to .png
func ensureImageExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
		return filename
	}
	return filename + ".png"
}
