package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/svg"
)

// rateLimiter ensures image generation is always sequential
// This prevents rate limiting on free AI image APIs
var imageRateLimiter sync.Mutex

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

// GenerateImage generates an image using AI with fallback providers
// ALWAYS sequential - waits for previous image generation to complete
func (s *AssetService) GenerateImage(campaign, filename, prompt, imgType string) (string, error) {
	// Try to acquire lock — return error immediately if another generation is in progress
	if !imageRateLimiter.TryLock() {
		return "", fmt.Errorf("image generation already in progress, try again later")
	}
	defer func() {
		// Delay before releasing to avoid rate limiting
		time.Sleep(3 * time.Second)
		imageRateLimiter.Unlock()
	}()

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


