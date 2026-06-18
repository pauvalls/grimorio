package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	consolidationPkg "github.com/pauvalls/grimorio/internal/services/consolidation"
)

// CampaignHealthCheck executes automated health checks on a campaign
type CampaignHealthCheck struct {
	canonRepo    repository.CanonRepository
	stateRepo    repository.NarrativeStateRepository
	questRepo    repository.QuestRepository
	factionRepo  repository.FactionReputationRepository
	npcRepo      repository.NPCRepository
	baseDir      string
	consolidator *ConsolidationAdapter
}

// HealthReport represents the output of a health check
type HealthReport struct {
	CampaignID    string          `json:"campaign_id"`
	GeneratedAt   time.Time       `json:"generated_at"`
	OverallHealth domain.HealthStatus `json:"overall_health"`
	Findings      []HealthFinding `json:"findings"`
	Summary       HealthSummary   `json:"summary"`
	ExecutedAt    time.Time       `json:"executed_at"`
	Duration      string          `json:"duration"` // Human-readable duration
}

// HealthFinding represents a single health check result
type HealthFinding struct {
	Rule           string            `json:"rule"`
	Severity       domain.Severity   `json:"severity"`
	Message        string            `json:"message"`
	EntityID       string            `json:"entity_id,omitempty"`
	EntityType     string            `json:"entity_type,omitempty"`
	Recommendation string            `json:"recommendation,omitempty"`
	Metadata       map[string]any    `json:"metadata,omitempty"`
}

// HealthSummary provides aggregate statistics
type HealthSummary struct {
	TotalFindings int `json:"total_findings"`
	CriticalCount int `json:"critical_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// NewCampaignHealthCheck creates a new health check service
func NewCampaignHealthCheck(
	canonRepo repository.CanonRepository,
	stateRepo repository.NarrativeStateRepository,
	questRepo repository.QuestRepository,
	factionRepo repository.FactionReputationRepository,
	npcRepo repository.NPCRepository,
	baseDir string,
) *CampaignHealthCheck {
	return &CampaignHealthCheck{
		canonRepo:    canonRepo,
		stateRepo:    stateRepo,
		questRepo:    questRepo,
		factionRepo:  factionRepo,
		npcRepo:      npcRepo,
		baseDir:      baseDir,
		consolidator: NewConsolidationAdapter(baseDir),
	}
}

// RunHealthCheck executes all health checks and returns a report
func (s *CampaignHealthCheck) RunHealthCheck(ctx context.Context, campaignID string) (*HealthReport, error) {
	start := time.Now()

	// Load required data
	canon, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	state, err := s.stateRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load narrative state: %w", err)
	}

	// Initialize report
	report := &HealthReport{
		CampaignID:  campaignID,
		GeneratedAt: time.Now(),
		ExecutedAt:  start,
		Findings:    []HealthFinding{},
	}

	// Run all health checks
	s.checkStaleQuests(state, report)
	s.checkFactionContradictions(canon, state, report)
	s.checkOrphanedClues(state, report)
	s.checkDeadNPCMismatch(canon, state, report)
	s.checkMcGuffinDrift(canon, state, report)

	// Consolidation-driven health findings (cross-file coherence).
	s.checkDuplicateFiles(ctx, campaignID, report)
	s.checkStaleGeneratedFiles(ctx, campaignID, report)
	s.checkMissingMaps(ctx, campaignID, report)
	s.checkLoreContradictions(ctx, campaignID, report)
	s.checkEntityNameCollisions(ctx, campaignID, report)
	s.checkBossStatBlockDrift(ctx, campaignID, report)

	// Sort findings by severity, rule, entity_id
	sortFindings(report.Findings)

	// Calculate summary and overall health
	report.Summary = calculateSummary(report.Findings)
	report.OverallHealth = calculateOverallHealth(report.Summary)
	report.Duration = time.Since(start).String()

	// Persist report
	if err := s.saveReport(campaignID, report); err != nil {
		// Non-fatal: save failure should not block health check
		_ = err // Intentionally ignoring error - logged in production
	}

	return report, nil
}

// GetHealthReport retrieves the most recent health report for a campaign
func (s *CampaignHealthCheck) GetHealthReport(ctx context.Context, campaignID string) (*HealthReport, error) {
	reportPath := filepath.Join(s.baseDir, campaignID, ".health_report.json")

	data, err := os.ReadFile(reportPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no health report found for campaign %s", campaignID)
		}
		return nil, fmt.Errorf("failed to read health report: %w", err)
	}

	var report HealthReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse health report: %w", err)
	}

	return &report, nil
}

// checkStaleQuests detects quests active for >10 sessions
func (s *CampaignHealthCheck) checkStaleQuests(state *domain.NarrativeState, report *HealthReport) {
	const staleThreshold = 10

	for _, quest := range state.ActiveQuests {
		if quest.Status != "active" {
			continue
		}

		// Find when quest was created (earliest session mentioning it)
		createdSession := findQuestCreationSession(state.SessionLog, quest.ID)
		sessionsActive := state.CurrentSession - createdSession

		if sessionsActive > staleThreshold {
			report.Findings = append(report.Findings, HealthFinding{
				Rule:     "stale_quest",
				Severity: domain.SeverityWarning,
				Message: fmt.Sprintf("Quest '%s' has been active for %d sessions without progress",
					quest.Name, sessionsActive),
				EntityID:       quest.ID,
				EntityType:     "quest",
				Recommendation: "Consider completing, failing, or replacing this quest",
				Metadata: map[string]any{
					"sessions_active": sessionsActive,
					"created_session": createdSession,
				},
			})
		}
	}
}

// checkFactionContradictions detects ally factions with hostile reputation
func (s *CampaignHealthCheck) checkFactionContradictions(
	canon *domain.CanonDocument,
	state *domain.NarrativeState,
	report *HealthReport,
) {
	// Load faction reputation matrix
	matrix, err := s.factionRepo.Load(canon.CampaignID)
	if err != nil {
		return // Non-fatal: skip this check
	}

	for _, entity := range canon.Entities {
		if entity.Type != domain.EntityTypeFaction {
			continue
		}

		// Get canon attitude
		canonAttitude, ok := entity.Properties["attitude"].(string)
		if !ok {
			continue
		}

		// Get reputation score
		var repScore int8
		for _, entry := range matrix.Entries {
			if entry.FactionID == entity.ID {
				repScore = entry.Score
				break
			}
		}

		// Check for contradiction: ally/friendly but hostile reputation (≤ -80)
		isFriendly := canonAttitude == "ally" || canonAttitude == "friendly"
		isHostile := repScore <= -80

		if isFriendly && isHostile {
			report.Findings = append(report.Findings, HealthFinding{
				Rule:     "faction_contradiction",
				Severity: domain.SeverityCritical,
				Message: fmt.Sprintf("Faction '%s' is marked as '%s' in canon but reputation is %d (hostile)",
					entity.Name, canonAttitude, repScore),
				EntityID:       entity.ID,
				EntityType:     "faction",
				Recommendation: "Update canon attitude or investigate reputation discrepancy",
				Metadata: map[string]any{
					"canon_attitude":   canonAttitude,
					"reputation_score": repScore,
				},
			})
		}
	}
}

// checkOrphanedClues detects clues with unrevealed prerequisites
func (s *CampaignHealthCheck) checkOrphanedClues(state *domain.NarrativeState, report *HealthReport) {
	// Build set of revealed clue IDs
	revealedSet := make(map[string]bool)
	for _, clue := range state.RevealedClues {
		revealedSet[clue.ID] = true
	}

	for _, clue := range state.RevealedClues {
		for _, prereqID := range clue.Prerequisites {
			if !revealedSet[prereqID] {
				report.Findings = append(report.Findings, HealthFinding{
					Rule:     "orphaned_clue",
					Severity: domain.SeverityWarning,
					Message: fmt.Sprintf("Clue '%s' requires prerequisite '%s' which was never revealed",
						clue.ID, prereqID),
					EntityID:       clue.ID,
					EntityType:     "clue",
					Recommendation: "Reveal prerequisite clue or remove prerequisite requirement",
					Metadata: map[string]any{
						"missing_prerequisite": prereqID,
					},
				})
			}
		}
	}
}

// checkDeadNPCMismatch detects NPCs dead in state but alive in canon
func (s *CampaignHealthCheck) checkDeadNPCMismatch(
	canon *domain.CanonDocument,
	state *domain.NarrativeState,
	report *HealthReport,
) {
	// Build set of dead NPC IDs
	deadSet := make(map[string]bool)
	for _, death := range state.DeadNPCs {
		deadSet[death.NPCID] = true
	}

	for _, entity := range canon.Entities {
		if entity.Type != domain.EntityTypeNPC {
			continue
		}

		if deadSet[entity.ID] && entity.CanonState != domain.EntityStateDead {
			// Find death record for details
			var deathRecord domain.NPCDeathRecord
			for _, d := range state.DeadNPCs {
				if d.NPCID == entity.ID {
					deathRecord = d
					break
				}
			}

			report.Findings = append(report.Findings, HealthFinding{
				Rule:     "dead_npc_mismatch",
				Severity: domain.SeverityCritical,
				Message: fmt.Sprintf("NPC '%s' is dead in narrative state (session %d) but canon state is '%s'",
					entity.Name, deathRecord.Session, entity.CanonState),
				EntityID:       entity.ID,
				EntityType:     "npc",
				Recommendation: "Update canon entity state to 'dead'",
				Metadata: map[string]any{
					"death_session":       deathRecord.Session,
					"current_canon_state": entity.CanonState,
				},
			})
		}
	}
}

// checkMcGuffinDrift detects McGuffin location mismatches
func (s *CampaignHealthCheck) checkMcGuffinDrift(
	canon *domain.CanonDocument,
	state *domain.NarrativeState,
	report *HealthReport,
) {
	for _, item := range state.KeyItems {
		if !item.IsMcGuffin {
			continue
		}

		// Find corresponding entity in canon
		for _, entity := range canon.Entities {
			if entity.Type != domain.EntityTypeItem || entity.ID != item.ID {
				continue
			}

			expectedLocation, ok := entity.Properties["expected_location"].(string)
			if !ok || expectedLocation == "" {
				continue
			}

			if item.Holder != expectedLocation {
				report.Findings = append(report.Findings, HealthFinding{
					Rule:     "mcguffin_drift",
					Severity: domain.SeverityCritical,
					Message: fmt.Sprintf("McGuffin '%s' is held by '%s' but expected at '%s'",
						item.Name, item.Holder, expectedLocation),
					EntityID:       item.ID,
					EntityType:     "item",
					Recommendation: "Update narrative state or adjust expected location",
					Metadata: map[string]any{
						"current_holder":    item.Holder,
						"expected_location": expectedLocation,
					},
				})
			}
		}
	}
}

// sortFindings sorts findings by severity, rule, entity_id
func sortFindings(findings []HealthFinding) {
	severityOrder := map[domain.Severity]int{
		domain.SeverityCritical: 0,
		domain.SeverityWarning:  1,
		domain.SeverityInfo:     2,
	}

	sort.Slice(findings, func(i, j int) bool {
		// Primary: severity
		if severityOrder[findings[i].Severity] != severityOrder[findings[j].Severity] {
			return severityOrder[findings[i].Severity] < severityOrder[findings[j].Severity]
		}
		// Secondary: rule name
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		// Tertiary: entity ID
		return findings[i].EntityID < findings[j].EntityID
	})
}

// calculateSummary computes aggregate statistics
func calculateSummary(findings []HealthFinding) HealthSummary {
	summary := HealthSummary{
		TotalFindings: len(findings),
	}

	for _, f := range findings {
		switch f.Severity {
		case domain.SeverityCritical:
			summary.CriticalCount++
		case domain.SeverityWarning:
			summary.WarningCount++
		case domain.SeverityInfo:
			summary.InfoCount++
		}
	}

	return summary
}

// calculateOverallHealth determines overall health status
func calculateOverallHealth(summary HealthSummary) domain.HealthStatus {
	if summary.CriticalCount > 0 {
		return domain.HealthStatusCritical
	}
	if summary.WarningCount > 0 {
		return domain.HealthStatusFair
	}
	if summary.InfoCount > 0 {
		return domain.HealthStatusGood
	}
	return domain.HealthStatusExcellent
}

// findQuestCreationSession finds the session where a quest was first mentioned
func findQuestCreationSession(sessionLog []domain.SessionRecord, questID string) int {
	for _, session := range sessionLog {
		// Simplified: in production, check quest creation metadata
		if session.SessionNum >= 1 {
			return session.SessionNum
		}
	}
	return 1
}

// saveReport persists the health report to filesystem
func (s *CampaignHealthCheck) saveReport(campaignID string, report *HealthReport) error {
	campaignDir := filepath.Join(s.baseDir, campaignID)
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		return fmt.Errorf("failed to create campaign directory: %w", err)
	}

	reportPath := filepath.Join(campaignDir, ".health_report.json")

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal health report: %w", err)
	}

	if err := os.WriteFile(reportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write health report: %w", err)
	}

	return nil
}

// computeCheckpointHash computes SHA256 hash of canon+state snapshots
func computeCheckpointHash(checkpoint *domain.PipelineCheckpoint) (string, error) {
	hasher := sha256.New()

	// Serialize canon snapshot
	canonData, err := json.Marshal(checkpoint.CanonSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canon snapshot: %w", err)
	}

	// Serialize state snapshot
	stateData, err := json.Marshal(checkpoint.StateSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state snapshot: %w", err)
	}

	// Hash both
	if _, err := hasher.Write(canonData); err != nil {
		return "", err
	}
	if _, err := hasher.Write(stateData); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// ----- Consolidation-driven health findings -----

// checkDuplicateFiles emits a warning for each set of identical-content files.
func (s *CampaignHealthCheck) checkDuplicateFiles(ctx context.Context, campaignID string, report *HealthReport) {
	res, err := s.consolidator.RunAnalyzer(ctx, campaignID, consolidationPkg.NewFileConsolidator())
	if err != nil || res == nil {
		return
	}
	hasDuplicates := false
	for _, issue := range res.Issues {
		if issue.Rule != "duplicate_file" {
			continue
		}
		hasDuplicates = true
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "duplicate_file",
			Severity:       domain.SeverityWarning,
			Message:        issue.Message,
			EntityID:       campaignID,
			EntityType:     "file",
			Recommendation: issue.Suggestion,
		})
	}
	if hasDuplicates && res.Passed {
		// Defensive: in practice Passed will already be false, but emit
		// a summary finding if the analyzer's internal state diverges.
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "duplicate_file",
			Severity:       domain.SeverityWarning,
			Message:        res.Message,
			EntityID:       campaignID,
			EntityType:     "file",
			Recommendation: "Remove duplicate files, keeping the canonical copy.",
		})
	}
}

// checkStaleGeneratedFiles reports campaign.md/INDEX.md older than sources.
func (s *CampaignHealthCheck) checkStaleGeneratedFiles(ctx context.Context, campaignID string, report *HealthReport) {
	freshness, err := s.consolidator.VerifyFreshness(ctx, campaignID)
	if err != nil || freshness == nil {
		return
	}
	if freshness.CampaignMDStale {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "stale_generated_file",
			Severity:       domain.SeverityWarning,
			Message:        "campaign.md is stale relative to sources",
			EntityID:       "campaign.md",
			EntityType:     "file",
			Recommendation: "Regenerate campaign.md to reflect current sources.",
		})
	}
	if freshness.IndexStale {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "stale_generated_file",
			Severity:       domain.SeverityWarning,
			Message:        "INDEX.md is stale relative to sources",
			EntityID:       "INDEX.md",
			EntityType:     "file",
			Recommendation: "Run regenerate_index to refresh the breadcrumb index.",
		})
	}
}

// checkMissingMaps reports broken map/asset references.
func (s *CampaignHealthCheck) checkMissingMaps(ctx context.Context, campaignID string, report *HealthReport) {
	campaignDir := filepath.Join(s.baseDir, campaignID)
	res, err := s.consolidator.RunAnalyzer(ctx, campaignID, consolidationPkg.NewMapReferenceChecker(campaignDir))
	if err != nil || res == nil {
		return
	}
	if !res.Passed {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "missing_map",
			Severity:       domain.SeverityCritical,
			Message:        res.Message,
			EntityType:     "asset",
			Recommendation: "Generate the missing map/asset or remove the broken reference.",
			Metadata: map[string]any{
				"locations": res.Locations,
			},
		})
	}
	for _, issue := range res.Issues {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "missing_map",
			Severity:       domain.SeverityCritical,
			Message:        issue.Message,
			EntityType:     "asset",
			Recommendation: issue.Suggestion,
		})
	}
}

// checkLoreContradictions surfaces treaty/event/primordial contradictions.
func (s *CampaignHealthCheck) checkLoreContradictions(ctx context.Context, campaignID string, report *HealthReport) {
	res, err := s.consolidator.RunAnalyzer(ctx, campaignID, consolidationPkg.NewLoreCoherence())
	if err != nil || res == nil {
		return
	}
	// Single overall finding for the check, then per-issue findings.
	if !res.Passed {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "lore_contradiction",
			Severity:       domain.SeverityCritical,
			Message:        res.Message,
			EntityType:     "lore",
			Recommendation: "Resolve treaty dates and event placement before exporting.",
			Metadata: map[string]any{
				"locations": res.Locations,
			},
		})
	}
	for _, issue := range res.Issues {
		// Treaty date contradictions are canon-breaking; treat as critical.
		severity := domain.SeverityCritical
		if issue.Severity == "warning" {
			severity = domain.SeverityWarning
		}
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "lore_contradiction",
			Severity:       severity,
			Message:        issue.Message,
			EntityType:     "lore",
			Recommendation: issue.Suggestion,
		})
	}
}

// checkEntityNameCollisions surfaces near-duplicate entity names.
func (s *CampaignHealthCheck) checkEntityNameCollisions(ctx context.Context, campaignID string, report *HealthReport) {
	res, err := s.consolidator.RunAnalyzer(ctx, campaignID, consolidationPkg.NewEntityResolver(0.85))
	if err != nil || res == nil {
		return
	}
	// Emit a single finding for the overall check when collisions exist —
	// high-similarity matches become auto-fixes in the consolidator, so
	// the issues slice may be empty even when Passed=false.
	if !res.Passed {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "entity_name_collision",
			Severity:       domain.SeverityWarning,
			Message:        res.Message,
			EntityType:     "entity",
			Recommendation: "Review the candidate collisions and resolve any ambiguity via resolve_ambiguity.",
			Metadata: map[string]any{
				"locations": res.Locations,
			},
		})
	}
	for _, issue := range res.Issues {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "entity_name_collision",
			Severity:       domain.SeverityWarning,
			Message:        issue.Message,
			EntityType:     "entity",
			Recommendation: issue.Suggestion,
		})
	}
}

// checkBossStatBlockDrift reports conflicting CR values for the same boss.
func (s *CampaignHealthCheck) checkBossStatBlockDrift(ctx context.Context, campaignID string, report *HealthReport) {
	res, err := s.consolidator.RunAnalyzer(ctx, campaignID, consolidationPkg.NewStatBlockResolver())
	if err != nil || res == nil {
		return
	}
	if !res.Passed {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "boss_stat_block_drift",
			Severity:       domain.SeverityCritical,
			Message:        res.Message,
			EntityType:     "monster",
			Recommendation: "Adopt the bestiary CR as canonical.",
			Metadata: map[string]any{
				"locations": res.Locations,
			},
		})
	}
	for _, issue := range res.Issues {
		report.Findings = append(report.Findings, HealthFinding{
			Rule:           "boss_stat_block_drift",
			Severity:       domain.SeverityCritical,
			Message:        issue.Message,
			EntityType:     "monster",
			Recommendation: issue.Suggestion,
		})
	}
}
