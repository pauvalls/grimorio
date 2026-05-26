package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// DMContextMonsterRepository defines the repository interface for monsters within the DM context service.
type DMContextMonsterRepository interface {
	List(ctx context.Context, campaignID string) ([]*domain.Monster, error)
}

// DMContextAreaRepository defines the repository interface for areas within the DM context service.
type DMContextAreaRepository interface {
	ListAll(ctx context.Context, campaignID string) ([]*domain.Area, error)
}

// DMContextService aggregates all campaign data into a single JSON payload for the AI Dungeon Master.
type DMContextService struct {
	canonRepo      repository.CanonRepository
	stateRepo      repository.NarrativeStateRepository
	charRepo       repository.CharacterRepository
	npcRepo        repository.NPCRepository
	questRepo      repository.QuestRepository
	monsterRepo    DMContextMonsterRepository
	areaRepo       DMContextAreaRepository
	factionRepo    repository.FactionReputationRepository
	sessionPrepSvc *SessionPrepService
	baseDir        string
}

// NewDMContextService creates a new DM context aggregation service.
func NewDMContextService(
	canonRepo repository.CanonRepository,
	stateRepo repository.NarrativeStateRepository,
	charRepo repository.CharacterRepository,
	npcRepo repository.NPCRepository,
	questRepo repository.QuestRepository,
	monsterRepo DMContextMonsterRepository,
	areaRepo DMContextAreaRepository,
	factionRepo repository.FactionReputationRepository,
	sessionPrepSvc *SessionPrepService,
	baseDir string,
) *DMContextService {
	return &DMContextService{
		canonRepo:      canonRepo,
		stateRepo:      stateRepo,
		charRepo:       charRepo,
		npcRepo:        npcRepo,
		questRepo:      questRepo,
		monsterRepo:    monsterRepo,
		areaRepo:       areaRepo,
		factionRepo:    factionRepo,
		sessionPrepSvc: sessionPrepSvc,
		baseDir:        baseDir,
	}
}

// GetContext aggregates all campaign data into a DMContextPayload.
// If compressionEnabled is true, sessions older than compressionThreshold are condensed.
func (s *DMContextService) GetContext(ctx context.Context, campaignID string, sessionNum int, includePrologue bool, includePDFText bool, compressionEnabled bool, compressionThreshold int) (*domain.DMContextPayload, []string, error) {
	warnings := []string{}
	payload := &domain.DMContextPayload{
		CampaignID:   campaignID,
		SessionNum:   sessionNum,
		GeneratedAt:  time.Now(),
		Characters:   []domain.CharacterContext{},
		Areas:        make(map[string]domain.AreaContext),
		NPCs:         make(map[string]domain.NPCContext),
		Bestiary:     make(map[string]domain.MonsterContext),
		Factions:     make(map[string]domain.FactionContext),
		Quests:       []domain.QuestContext{},
		PDFAvailable: false,
		DMNotes: domain.DMNotesContext{
			Warnings:  []string{},
			Reminders: []string{},
		},
	}

	// Load canon (required)
	canonDoc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, warnings, fmt.Errorf("canon document not found for campaign: %s", campaignID)
	}
	payload.Canon = buildCanonContext(canonDoc)

	// Load narrative state (required)
	state, err := s.stateRepo.Load(campaignID)
	if err != nil {
		return nil, warnings, fmt.Errorf("narrative state not found for campaign: %s", campaignID)
	}
	if sessionNum <= 0 {
		sessionNum = state.CurrentSession
		if sessionNum <= 0 {
			sessionNum = 1
		}
	}
	payload.SessionNum = sessionNum
	
	// Apply compression if enabled
	if compressionEnabled && compressionThreshold > 0 {
		state.SessionLog = s.compressSessionLog(state.SessionLog, sessionNum, compressionThreshold)
	}
	
	payload.NarrativeState = buildNarrativeContext(state)

	// Load session prep (optional)
	if s.sessionPrepSvc != nil {
		prep, prepWarnings, err := s.sessionPrepSvc.GetPrep(ctx, campaignID, sessionNum)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to load session prep: %v", err))
		} else {
			payload.SessionPrep = &domain.DMContextSessionPrep{
				PreviouslyOn:    prep.PreviouslyOn,
				ActiveQuests:    orEmptySlice(prep.ActiveQuests),
				RelevantNPCs:    orEmptySlice(prep.RelevantNPCs),
				Reminders:       orEmptySlice(prep.Reminders),
				LikelyScenarios: orEmptySlice(prep.LikelyScenarios),
			}
			warnings = append(warnings, prepWarnings...)
		}
	}

	// Load characters (optional)
	chars, err := s.charRepo.List(campaignID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to load characters: %v", err))
	} else {
		for _, c := range chars {
			payload.Characters = append(payload.Characters, domain.CharacterContext{
				Name:       c.Name,
				Race:       c.Race,
				Class:      c.Class,
				Level:      c.Level,
				Background: c.Background,
				Alignment:  c.Alignment,
				HP:         c.HP,
				AC:         c.AC,
				Stats:      c.Stats,
			})
		}
	}

	// Load NPCs (optional)
	npcs, err := s.npcRepo.List(campaignID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to load NPCs: %v", err))
	} else {
		for _, n := range npcs {
			npcCtx := domain.NPCContext{
				Name:        n.Name,
				Description: n.Description,
				Faction:     n.Faction,
			}
			if n.Stats != nil {
				npcCtx.Stats = domain.NPCStats{
					HP: n.Stats.HP,
					AC: n.Stats.AC,
				}
			}
			// Enrich from canon entities
			for _, e := range canonDoc.Entities {
				if e.Type == domain.EntityTypeNPC && e.Name == n.Name {
					npcCtx.Motivation = e.Motivation
					npcCtx.Secret = e.Secret
					if voice, ok := e.Properties["dialogue_voice"].(string); ok {
						npcCtx.DialogueVoice = voice
					}
					if traits, ok := e.Properties["personality_traits"].([]string); ok {
						npcCtx.Personality = traits
					}
					if tactics, ok := e.Properties["tactics"].(string); ok {
						npcCtx.Tactics = tactics
					}
					break
				}
			}
			payload.NPCs[n.Name] = npcCtx
		}
	}

	// Load quests (optional)
	quests, err := s.questRepo.List(campaignID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to load quests: %v", err))
	} else {
		for _, q := range quests {
			giver := ""
			if len(q.RelatedNPCs) > 0 {
				giver = q.RelatedNPCs[0]
			}
			payload.Quests = append(payload.Quests, domain.QuestContext{
				ID:     q.ID,
				Title:  q.Title,
				Status: string(q.Status),
				Type:   string(q.Type),
				Giver:  giver,
			})
		}
	}

	// Load bestiary (optional)
	if s.monsterRepo != nil {
		monsters, err := s.monsterRepo.List(ctx, campaignID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to load bestiary: %v", err))
		} else {
			for _, m := range monsters {
				mc := domain.MonsterContext{
					Name:    m.Name,
					CR:      m.CR,
					AC:      m.Stats.AC,
					HP:      m.Stats.HP,
					Tactics: "",
				}
				// Enrich descriptive cues from canon entity properties
				for _, e := range canonDoc.Entities {
					if e.Type == domain.EntityTypeMonster && e.Name == m.Name {
						if cues, ok := e.Properties["descriptive_cues"].(map[string]any); ok {
							mc.DescriptiveCues = make(map[string]string)
							for k, v := range cues {
								if s, ok := v.(string); ok {
									mc.DescriptiveCues[k] = s
								}
							}
						}
						if tactics, ok := e.Properties["tactics"].(string); ok {
							mc.Tactics = tactics
						}
						break
					}
				}
				if mc.DescriptiveCues == nil {
					mc.DescriptiveCues = map[string]string{}
				}
				payload.Bestiary[m.Name] = mc
			}
		}
	}

	// Load areas (optional)
	if s.areaRepo != nil {
		areas, err := s.areaRepo.ListAll(ctx, campaignID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to load areas: %v", err))
		} else {
			for _, a := range areas {
				payload.Areas[a.ID] = domain.AreaContext{
					ID:              a.ID,
					ChapterID:       a.ChapterID,
					AreaNumber:      a.AreaNumber,
					Title:           a.Title,
					Summary:         a.Summary,
					PlayerReadAloud: a.PlayerReadAloud,
					Encounters:      a.Encounters,
					NPCs:            a.NPCs,
					Treasure:        a.Treasure,
				}
			}
		}
	}

	// Load factions (optional)
	if s.factionRepo != nil {
		matrix, err := s.factionRepo.Load(campaignID)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to load faction reputation: %v", err))
		} else {
			for _, entry := range matrix.Entries {
				name := entry.FactionID
				attitude := "neutral"
				// Lookup name from canon entities
				for _, e := range canonDoc.Entities {
					if e.Type == domain.EntityTypeFaction && e.ID == entry.FactionID {
						name = e.Name
						if a, ok := e.Properties["attitude"].(string); ok {
							attitude = a
						}
						break
					}
				}
				payload.Factions[entry.FactionID] = domain.FactionContext{
					ID:         entry.FactionID,
					Name:       name,
					Reputation: entry.Score,
					Status:     entry.Status,
					Attitude:   attitude,
				}
			}
		}
	}

	// Load prologue (optional)
	if includePrologue {
		prologuePath := filepath.Join(s.baseDir, campaignID, "prologue.md")
		if data, err := os.ReadFile(prologuePath); err == nil {
			payload.Prologue = &domain.PrologueContext{
				Tone: "heroic",
				Parts: []domain.ProloguePartContext{
					{
						Order:       1,
						Title:       "Prólogo",
						Content:     string(data),
						IsReadAloud: true,
					},
				},
			}
		}
	}

	// Check PDF availability
	pdfPath := filepath.Join(s.baseDir, campaignID+".pdf")
	if _, err := os.Stat(pdfPath); err == nil {
		payload.PDFAvailable = true
		payload.PDFPath = pdfPath
		if includePDFText {
			// Extract text from PDF using pdftotext if available
			text, err := s.extractPDFText(pdfPath)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to extract PDF text: %v", err))
			} else {
				payload.PDFText = text
			}
		}
	}

	// Build DM notes from warnings
	payload.DMNotes.Warnings = warnings
	payload.DMNotes.Reminders = s.buildReminders(canonDoc, state)

	// Size cap check
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, warnings, fmt.Errorf("failed to serialize context payload: %w", err)
	}
	if len(jsonBytes) > 100*1024 {
		warnings = append(warnings, fmt.Sprintf("payload size exceeds 100KB cap: %d bytes", len(jsonBytes)))
	}

	return payload, warnings, nil
}

// extractPDFText attempts to extract text from a PDF file using pdftotext (poppler-utils).
// Falls back to a placeholder message if the tool is not available.
func (s *DMContextService) extractPDFText(pdfPath string) (string, error) {
	// Check if pdftotext is available
	_, err := os.Stat("/usr/bin/pdftotext")
	if err != nil {
		// Try alternative paths
		_, err = os.Stat("/usr/local/bin/pdftotext")
		if err != nil {
			return "", fmt.Errorf("pdftotext not available (install poppler-utils to enable PDF text extraction)")
		}
	}

	// Create temp file for text output
	tmpFile, err := os.CreateTemp("", "grimorio-pdf-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func(name string) { _ = os.Remove(name) }(tmpFile.Name())
	_ = tmpFile.Close()

	// Run pdftotext
	cmd := exec.Command("pdftotext", pdfPath, tmpFile.Name())
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}

	// Read extracted text
	text, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read extracted text: %w", err)
	}

	// Truncate if too large (>50KB)
	const maxPDFTextSize = 50 * 1024
	if len(text) > maxPDFTextSize {
		return string(text[:maxPDFTextSize]) + "\n\n[... PDF text truncated at 50KB ...]", nil
	}

	return string(text), nil
}

func buildCanonContext(doc *domain.CanonDocument) *domain.CanonContext {
	if doc == nil {
		return &domain.CanonContext{
			Facts:         []domain.CanonFact{},
			Entities:      []domain.CanonEntity{},
			Timeline:      []domain.CanonTimelineEvent{},
			Rules:         []domain.CanonRule{},
			Relationships: []domain.CanonRelationship{},
		}
	}
	return &domain.CanonContext{
		Facts:         orEmptySlice(doc.Facts),
		Entities:      orEmptySlice(doc.Entities),
		Timeline:      orEmptySlice(doc.Timeline),
		Rules:         orEmptySlice(doc.Rules),
		Relationships: orEmptySlice(doc.Relationships),
	}
}

func buildNarrativeContext(state *domain.NarrativeState) *domain.NarrativeContext {
	if state == nil {
		return &domain.NarrativeContext{
			CurrentSession:  0,
			RevealedClues:   []domain.RevealedClue{},
			ActiveQuests:    []domain.QuestState{},
			CompletedQuests: []domain.QuestState{},
			FailedQuests:    []domain.QuestState{},
			DeadNPCs:        []domain.NPCDeathRecord{},
			KeyItems:        []domain.KeyItem{},
			SessionLog:      []domain.SessionRecord{},
		}
	}
	return &domain.NarrativeContext{
		CurrentSession:  state.CurrentSession,
		RevealedClues:   orEmptySlice(state.RevealedClues),
		ActiveQuests:    orEmptySlice(state.ActiveQuests),
		CompletedQuests: orEmptySlice(state.CompletedQuests),
		FailedQuests:    orEmptySlice(state.FailedQuests),
		DeadNPCs:        orEmptySlice(state.DeadNPCs),
		KeyItems:        orEmptySlice(state.KeyItems),
		SessionLog:      orEmptySlice(state.SessionLog),
	}
}

func (s *DMContextService) buildReminders(doc *domain.CanonDocument, state *domain.NarrativeState) []string {
	reminders := []string{}
	if doc == nil || state == nil {
		return reminders
	}

	// Check for dead NPCs whose canon state is still alive
	deadSet := make(map[string]bool)
	for _, d := range state.DeadNPCs {
		deadSet[d.NPCID] = true
	}
	for _, e := range doc.Entities {
		if e.Type == domain.EntityTypeNPC && deadSet[e.ID] && e.CanonState == domain.EntityStateAlive {
			reminders = append(reminders, fmt.Sprintf("NPC %s is dead in narrative state but alive in canon", e.Name))
		}
	}

	// Pending DM overrides
	for _, o := range state.DMOverrides {
		reminders = append(reminders, fmt.Sprintf("DM Override pending: %s on %s (%s → %s)", o.Reason, o.TargetID, o.Field, o.NewValue))
	}

	return reminders
}

// compressSessionLog condenses old sessions into a single summary
func (s *DMContextService) compressSessionLog(
	sessionLog []domain.SessionRecord,
	currentSession int,
	threshold int,
) []domain.SessionRecord {
	if len(sessionLog) <= threshold {
		return sessionLog
	}

	compressed := []domain.SessionRecord{}
	condensedSessions := []domain.SessionRecord{}

	// Separate old vs recent sessions
	for _, session := range sessionLog {
		if session.SessionNum <= (currentSession - threshold) {
			condensedSessions = append(condensedSessions, session)
		} else {
			compressed = append(compressed, session)
		}
	}

	// Build condensed summary from old sessions
	if len(condensedSessions) > 0 {
		condensedText := s.buildCondensedSummary(condensedSessions)

		// Prepend placeholder record
		compressed = append([]domain.SessionRecord{
			{
				SessionNum:       0, // Special value indicating condensed range
				CondensedSummary: condensedText,
				Summary:          fmt.Sprintf("Sessions 1-%d condensed", currentSession-threshold),
			},
		}, compressed...)
	}

	return compressed
}

// buildCondensedSummary extracts key information from old sessions
func (s *DMContextService) buildCondensedSummary(sessions []domain.SessionRecord) string {
	var builder strings.Builder

	builder.WriteString("=== SESIONES CONDENSADAS ===\n\n")

	// Collect key decisions from condensed sessions
	builder.WriteString("Decisiones Mayores:\n")
	for _, session := range sessions {
		for _, decision := range session.KeyDecisions {
			if decision.ImpactScope == "campaign" {
				builder.WriteString(fmt.Sprintf("- %s (sesión %d)\n", decision.Description, session.SessionNum))
			}
		}
	}

	// Add session summaries
	builder.WriteString("\nResúmenes de Sesiones:\n")
	for _, session := range sessions {
		builder.WriteString(fmt.Sprintf("- Sesión %d: %s\n", session.SessionNum, session.Summary))
	}

	return builder.String()
}

// filterNPCsByRelevance filters NPCs based on current location and active quests
func (s *DMContextService) filterNPCsByRelevance(
	npcs map[string]domain.NPCContext,
	currentLocation string,
	activeQuests []domain.QuestState,
	canon *domain.CanonDocument,
) map[string]domain.NPCContext {
	relevant := make(map[string]domain.NPCContext)

	// Build set of relevant NPC IDs
	relevantIDs := make(map[string]bool)

	// Add quest givers
	for _, quest := range activeQuests {
		if quest.GiverNPC != "" {
			relevantIDs[quest.GiverNPC] = true
		}
	}

	// Add NPCs at current location
	for _, entity := range canon.Entities {
		if entity.Type == domain.EntityTypeNPC {
			if location, ok := entity.Properties["location"].(string); ok {
				if location == currentLocation {
					relevantIDs[entity.ID] = true
				}
			}
		}
	}

	// Filter NPCs
	for name, npc := range npcs {
		if relevantIDs[npc.Name] || relevantIDs[name] {
			relevant[name] = npc
		}
	}

	// If no relevant NPCs found, return all (fallback)
	if len(relevant) == 0 {
		return npcs
	}

	return relevant
}

func orEmptySlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
