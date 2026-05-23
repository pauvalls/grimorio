package grimorio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestGoReleaserConfigBundlesRuntimeAssets verifies that the GoReleaser
// configuration includes agents/ and skills/ directories in release archives
// for all target platforms, with the correct archive formats.
func TestGoReleaserConfigBundlesRuntimeAssets(t *testing.T) {
	yamlBytes, err := os.ReadFile(".goreleaser.yaml")
	require.NoError(t, err, "should be able to read .goreleaser.yaml")

	var config map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &config)
	require.NoError(t, err, "should be able to parse .goreleaser.yaml as YAML")

	// Verify builds section covers all required platforms.
	builds, ok := config["builds"].([]interface{})
	require.True(t, ok, "builds section should exist")
	require.Len(t, builds, 1, "should have exactly one build config")

	build := builds[0].(map[string]interface{})
	goos, ok := build["goos"].([]interface{})
	require.True(t, ok, "builds should specify goos")
	assert.ElementsMatch(t, []interface{}{"linux", "darwin", "windows"}, goos,
		"should build for linux, darwin, and windows")

	goarch, ok := build["goarch"].([]interface{})
	require.True(t, ok, "builds should specify goarch")
	assert.ElementsMatch(t, []interface{}{"amd64", "arm64"}, goarch,
		"should build for amd64 and arm64")

	// Verify archives section includes agents/ and skills/.
	archives, ok := config["archives"].([]interface{})
	require.True(t, ok, "archives section should exist")
	require.Len(t, archives, 1, "should have exactly one archive config")

	archive := archives[0].(map[string]interface{})

	// Verify archive format is tar.gz by default.
	format, ok := archive["format"].(string)
	require.True(t, ok, "archive format should be specified")
	assert.Equal(t, "tar.gz", format, "default archive format should be tar.gz")

	// Verify format override for Windows -> zip.
	formatOverrides, ok := archive["format_overrides"].([]interface{})
	require.True(t, ok, "format_overrides should exist")
	require.Len(t, formatOverrides, 1, "should have one format override")

	override := formatOverrides[0].(map[string]interface{})
	assert.Equal(t, "windows", override["goos"], "should override format for windows")
	assert.Equal(t, "zip", override["format"], "windows archives should use zip")

	// Verify archive wraps everything in a single directory.
	wrapInDir, ok := archive["wrap_in_directory"].(bool)
	require.True(t, ok, "wrap_in_directory should be set")
	assert.True(t, wrapInDir, "archive should wrap files in a single directory")

	// Verify files section includes agents/ and skills/.
	files, ok := archive["files"].([]interface{})
	require.True(t, ok, "archive files section should exist")

	var hasAgents, hasSkills bool
	for _, f := range files {
		switch v := f.(type) {
		case string:
			if v == "agents/*" || v == "agents/**/*" {
				hasAgents = true
			}
			if v == "skills/**/*" {
				hasSkills = true
			}
		case map[string]interface{}:
			src, _ := v["src"].(string)
			if src == "agents/*" || src == "agents/**/*" {
				hasAgents = true
			}
			if src == "skills/**/*" {
				hasSkills = true
			}
		}
	}

	assert.True(t, hasAgents, "archive should include agents/ directory")
	assert.True(t, hasSkills, "archive should include skills/ directory")
}
