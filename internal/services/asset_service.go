package services

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/svg"
)

// AssetService handles asset generation (maps, images, dividers)
type AssetService struct {
	baseDir      string
	imgConfig    image.Config
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
	// Note: The caller handles file writing to maintain compatibility
	return svgContent, nil
}

// GenerateDivider generates a decorative SVG divider
func (s *AssetService) GenerateDivider(campaign, filename, style string, width int) (string, error) {
	if style == "" {
		style = "ornate"
	}

	svgContent := svg.GenerateDivider(width, style)
	// Note: The caller handles file writing
	return svgContent, nil
}

// GenerateImage generates an image using AI with fallback providers
func (s *AssetService) GenerateImage(campaign, filename, prompt, imgType string) (string, error) {
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

// ensureImageExt ensures filename has an image extension, defaulting to .png
func ensureImageExt(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".webp" {
		return filename
	}
	return filename + ".png"
}

// GenerateImagesBatch generates multiple images SEQUENTIALLY with fallback providers
// Sequential generation avoids rate limiting on free APIs
func (s *AssetService) GenerateImagesBatch(campaign string, images []BatchImageSpec) ([]BatchImageResult, error) {
	providers := s.getProviders()
	results := make([]BatchImageResult, len(images))

	for i, spec := range images {
		outputPath := filepath.Join(s.assetsDir(campaign), ensureImageExt(spec.Filename))
		providerName, err := image.GenerateAndSaveWithFallback(providers, spec.Prompt, outputPath)

		if err != nil {
			results[i] = BatchImageResult{
				Filename:    spec.Filename,
				Success:     false,
				Error:       err.Error(),
				ProviderUsed: "",
			}
		} else {
			results[i] = BatchImageResult{
				Filename:     spec.Filename,
				Success:      true,
				Path:         outputPath,
				ProviderUsed: providerName,
			}
		}

		// Delay between requests to avoid rate limiting on free APIs
		// Only delay if there are more images to process
		if i < len(images)-1 {
			time.Sleep(3 * time.Second)
		}
	}

	return results, nil
}

// BatchImageSpec represents an image to generate
type BatchImageSpec struct {
	Filename string
	Prompt   string
	Type     string
}

// BatchImageResult represents the result of generating an image
type BatchImageResult struct {
	Filename     string
	Success      bool
	Path         string
	Error        string
	ProviderUsed string
}
