package piper

import (
	"regexp"
	"strings"
)

// TextFilter removes non-narratable content from DM text before TTS synthesis.
type TextFilter interface {
	Filter(text string) string
}

// DefaultTextFilter implements TextFilter by removing markdown tables
// and <thinking>...</thinking> blocks.
type DefaultTextFilter struct{}

var (
	// thinkingRegex matches <thinking>...</thinking> blocks (multiline, non-greedy).
	thinkingRegex = regexp.MustCompile(`(?s)<thinking>.*?</thinking>`)
)

// Filter removes markdown table lines and thinking blocks from the input text.
func (f *DefaultTextFilter) Filter(text string) string {
	// Remove thinking blocks first
	text = thinkingRegex.ReplaceAllString(text, "")

	// Remove markdown table lines (lines starting with |)
	lines := strings.Split(text, "\n")
	var filtered []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "|") {
			continue
		}
		filtered = append(filtered, line)
	}

	// Rejoin and clean up extra whitespace
	result := strings.Join(filtered, "\n")
	result = strings.TrimSpace(result)
	// Collapse multiple consecutive blank lines into one
	result = collapseBlankLines(result)
	return result
}

func collapseBlankLines(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	prevBlank := false
	for _, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}
	return strings.Join(result, "\n")
}
