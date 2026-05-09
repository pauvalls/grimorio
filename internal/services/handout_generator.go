package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
	"strings"
)

// HandoutGenerator generates various handout types.
type HandoutGenerator struct {
	handoutRepo HandoutRepositoryV3
}

// NewHandoutGenerator creates a new HandoutGenerator.
func NewHandoutGenerator(handoutRepo HandoutRepositoryV3) *HandoutGenerator {
	return &HandoutGenerator{handoutRepo: handoutRepo}
}

// GenerateLetter generates a letter handout.
func (s *HandoutGenerator) GenerateLetter(ctx context.Context, campaignID, sender, recipient, purpose, style string) (*domain.Handout, error) {
	handout := &domain.Handout{
		ID:         fmt.Sprintf("letter_%s_%d", campaignID, len(sender)),
		CampaignID: campaignID,
		Type:       domain.HandoutTypeLetter,
		Title:      fmt.Sprintf("Letter from %s", sender),
		Content:    generateLetterContent(sender, recipient, purpose, style),
		DMNotes:    fmt.Sprintf("Give to players when they meet %s or find the letter", sender),
		Format:     domain.FormatText,
		Style:      getHandoutStyle(style),
	}

	if err := handout.Validate(); err != nil {
		return nil, fmt.Errorf("generated letter validation failed: %w", err)
	}

	return handout, nil
}

// GenerateClue generates a clue handout.
func (s *HandoutGenerator) GenerateClue(ctx context.Context, campaignID, questID, clueType, location string, discoveryDC int) (*domain.Handout, error) {
	handout := &domain.Handout{
		ID:         fmt.Sprintf("clue_%s_%s", campaignID, questID),
		CampaignID: campaignID,
		Type:       domain.HandoutTypeClue,
		Title:      fmt.Sprintf("Clue: %s", clueType),
		Content:    generateClueContent(clueType, location),
		DMNotes:    fmt.Sprintf("Hide in %s, DC %d to discover", location, discoveryDC),
		Format:     domain.FormatText,
		Style:      domain.StyleMysterious,
		QuestRefs:  []string{questID},
		RevealConditions: []string{
			fmt.Sprintf("DC %d Investigation check", discoveryDC),
			fmt.Sprintf("Found in %s", location),
		},
	}

	if err := handout.Validate(); err != nil {
		return nil, fmt.Errorf("generated clue validation failed: %w", err)
	}

	return handout, nil
}

// GenerateDocument generates a document or journal handout.
func (s *HandoutGenerator) GenerateDocument(ctx context.Context, campaignID, docType, title, style string, pages int) (*domain.Handout, error) {
	handout := &domain.Handout{
		ID:         fmt.Sprintf("doc_%s_%d", campaignID, len(title)),
		CampaignID: campaignID,
		Type:       getDocumentType(docType),
		Title:      title,
		Content:    generateDocumentContent(docType, title, pages),
		DMNotes:    fmt.Sprintf("%s handout with %d pages", docType, pages),
		Format:     domain.FormatText,
		Style:      getHandoutStyle(style),
	}

	if err := handout.Validate(); err != nil {
		return nil, fmt.Errorf("generated document validation failed: %w", err)
	}

	return handout, nil
}

// ExportHandout exports a handout in the specified format.
func (s *HandoutGenerator) ExportHandout(ctx context.Context, handoutID, format string) ([]byte, error) {
	// TODO: Implement export with proper CSS styling
	return []byte{}, nil
}

func generateLetterContent(sender, recipient, purpose, style string) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("To %s,\n\n", recipient))
	
	switch style {
	case "formal":
		content.WriteString(fmt.Sprintf("From %s.\n\n", sender))
		content.WriteString(fmt.Sprintf("I write to you regarding %s.\n\n", purpose))
		content.WriteString("Your humble servant,\n")
	case "urgent":
		content.WriteString(fmt.Sprintf("%s sends word: %s\n\n", sender, purpose))
		content.WriteString("ACT NOW!\n")
	default:
		content.WriteString(fmt.Sprintf("%s here. About %s...\n\n", sender, purpose))
	}
	
	content.WriteString(fmt.Sprintf("[Signature: %s]", sender))
	return content.String()
}

func generateClueContent(clueType, location string) string {
	switch clueType {
	case "note":
		return "A crumpled note with hasty writing..."
	case "map":
		return "A partial map showing location markings..."
	case "symbol":
		return "A strange symbol carved into surface..."
	default:
		return fmt.Sprintf("A clue found in %s...", location)
	}
}

func generateDocumentContent(docType, title string, pages int) string {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", title))
	
	for i := 1; i <= pages; i++ {
		content.WriteString(fmt.Sprintf("## Page %d\n\n", i))
		content.WriteString("[Content for this page...]\n\n")
	}
	
	return content.String()
}

func getHandoutStyle(style string) domain.HandoutStyle {
	switch style {
	case "formal":
		return domain.StyleFormal
	case "informal":
		return domain.StyleInformal
	case "ancient":
		return domain.StyleAncient
	case "urgent":
		return domain.StyleUrgent
	case "mysterious":
		return domain.StyleMysterious
	default:
		return domain.StyleFormal
	}
}

func getDocumentType(docType string) domain.HandoutType {
	switch docType {
	case "journal":
		return domain.HandoutTypeJournal
	case "proclamation":
		return domain.HandoutTypeProclamation
	case "artifact":
		return domain.HandoutTypeArtifact
	default:
		return domain.HandoutTypeDocument
	}
}
