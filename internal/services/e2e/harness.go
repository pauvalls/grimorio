package e2e

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

// TestHarness wires all services with filesystem repositories for E2E testing.
type TestHarness struct {
	CampaignService         *services.CampaignService
	CharacterService        *services.CharacterService
	QuestService            *services.QuestService
	CanonService            *services.CanonService
	NarrativeStateService   *services.NarrativeStateService
	SessionPrepService      *services.SessionPrepService
	ValidationEngine        *services.ValidationEngine
	ConsistencyGateService  *services.ConsistencyGateService
	FactionService          *services.FactionService
	BaseDir                 string
}

// NewTestHarness creates a fully wired TestHarness using filesystem repositories in a temp dir.
func NewTestHarness(t *testing.T) *TestHarness {
	baseDir := t.TempDir()

	campaignRepo := repository.NewFilesystemCampaignRepository(baseDir)
	actRepo := repository.NewFilesystemActRepository(baseDir)
	charRepo := repository.NewFilesystemCharacterRepository(baseDir)
	npcRepo := repository.NewFilesystemNPCRepository(baseDir)
	questRepo := repository.NewFilesystemQuestRepository(baseDir)
	canonRepo := repository.NewFilesystemCanonRepository(baseDir)
	monsterRepo := repository.NewMemoryMonsterRepository()
	stateRepo := repository.NewFilesystemNarrativeStateRepository(baseDir)
	checkpointRepo := repository.NewMemoryCheckpointRepository()
	auditRepo := repository.NewMemoryAuditLogRepository()
	factionRepo := repository.NewFilesystemFactionRepository(baseDir)

	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		canonRepo, monsterRepo, baseDir, "",
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)
	canonService := services.NewCanonService(canonRepo, stateRepo, checkpointRepo)
	stateService := services.NewNarrativeStateService(stateRepo, canonRepo)
	validationEngine := services.NewValidationEngine(canonService, stateService, factionRepo, baseDir)
	gateService := services.NewConsistencyGateService(canonService, stateService, validationEngine, checkpointRepo, auditRepo)
	factionService := services.NewFactionService(canonRepo, factionRepo)
	sessionPrepService := services.NewSessionPrepService(canonRepo, stateRepo, factionRepo)

	return &TestHarness{
		CampaignService:        campaignService,
		CharacterService:       characterService,
		QuestService:           questService,
		CanonService:           canonService,
		NarrativeStateService:  stateService,
		SessionPrepService:     sessionPrepService,
		ValidationEngine:       validationEngine,
		ConsistencyGateService: gateService,
		FactionService:         factionService,
		BaseDir:                baseDir,
	}
}

// CleanupCampaign removes a campaign by name from the harness state.
func (h *TestHarness) CleanupCampaign(name string) error {
	return nil // Temp dir auto-cleaned on test exit
}
