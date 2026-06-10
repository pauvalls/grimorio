package e2e

import (
	"context"
	"fmt"

	"github.com/pauvalls/grimorio/internal/domain"
)

// executeStep dispatches a test step to the appropriate harness method.
func executeStep(ctx context.Context, harness *TestHarness, step TestStep) error {
	dispatch := map[string]func(context.Context, *TestHarness, TestStep) error{
		"create_campaign":      executeCreateCampaign,
		"save_areas":           executeSaveAreas,
		"save_npcs":            executeSaveNPCs,
		"save_encounters":      executeSaveEncounters,
		"save_bestiary":        executeSaveBestiary,
		"save_lore":            executeSaveLore,
		"save_introduction":    executeSaveIntroduction,
		"save_setting_guide":   executeSaveSettingGuide,
		"save_appendices":      executeSaveAppendices,
		"save_maps":            executeSaveMaps,
		"save_chapter":         executeSaveChapter,
		"compile_pdf":          executeCompilePDF,
		"generate_character":   executeGenerateCharacter,
		"save_characters":      executeSaveCharacters,
		"create_quest":         executeCreateQuest,
		"update_quest_status":  executeUpdateQuestStatus,
		"generate_adventure_bible": executeGenerateAdventureBible,
		"update_narrative_state":   executeUpdateNarrativeState,
		"check_consistency":    executeCheckConsistency,
		"generate_session_prep": executeGenerateSessionPrep,
	}

	fn, ok := dispatch[step.Action]
	if !ok {
		return fmt.Errorf("unsupported action: %s", step.Action)
	}

	return fn(ctx, harness, step)
}

func getStringParam(params map[string]any, key string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getIntParam(params map[string]any, key string) int {
	if v, ok := params[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		}
	}
	return 0
}

func executeCreateCampaign(_ context.Context, h *TestHarness, step TestStep) error {
	name := getStringParam(step.Params, "name")
	if name == "" {
		name = "test_campaign"
	}
	_, err := h.CampaignService.CreateCampaign(name, getStringParam(step.Params, "title"), getStringParam(step.Params, "setting"))
	return err
}

func executeSaveAreas(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveAct(
		getStringParam(step.Params, "campaign"),
		getIntParam(step.Params, "chapter_number"),
		getStringParam(step.Params, "title"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveNPCs(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveNPCs(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveEncounters(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveEncounters(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveBestiary(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveBestiary(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveLore(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveLore(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveIntroduction(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveIntroduction(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveSettingGuide(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveSettingGuide(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveAppendices(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveAppendices(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveMaps(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveMaps(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "content"),
	)
}

func executeSaveChapter(_ context.Context, h *TestHarness, step TestStep) error {
	return h.CampaignService.SaveChapter(
		getStringParam(step.Params, "campaign"),
		getIntParam(step.Params, "chapter_number"),
		getStringParam(step.Params, "title"),
		getStringParam(step.Params, "content"),
	)
}

func executeCompilePDF(ctx context.Context, h *TestHarness, step TestStep) error {
	_, err := h.CampaignService.CompilePDF(ctx, getStringParam(step.Params, "campaign"), getStringParam(step.Params, "title"))
	return err
}

func executeGenerateCharacter(_ context.Context, h *TestHarness, step TestStep) error {
	_, err := h.CharacterService.CreateCharacter(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "name"),
		getStringParam(step.Params, "race"),
		getStringParam(step.Params, "class"),
		getIntParam(step.Params, "level"),
		getStringParam(step.Params, "background"),
		getStringParam(step.Params, "alignment"),
	)
	return err
}

func executeSaveCharacters(_ context.Context, h *TestHarness, step TestStep) error {
	char := &domain.Character{
		CampaignID: getStringParam(step.Params, "campaign"),
		Name:       getStringParam(step.Params, "name"),
		Race:       getStringParam(step.Params, "race"),
		Class:      getStringParam(step.Params, "class"),
		Level:      getIntParam(step.Params, "level"),
		Background: getStringParam(step.Params, "background"),
		Alignment:  getStringParam(step.Params, "alignment"),
		Status:     "alive",
	}
	return h.CharacterService.SaveCharacter(char)
}

func executeCreateQuest(_ context.Context, h *TestHarness, step TestStep) error {
	_, err := h.QuestService.CreateQuest(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "quest_title"),
		domain.QuestType(getStringParam(step.Params, "quest_type")),
		getStringParam(step.Params, "hook"),
		"",
		getStringParam(step.Params, "stakes"),
		nil,
	)
	return err
}

func executeUpdateQuestStatus(_ context.Context, h *TestHarness, step TestStep) error {
	return h.QuestService.UpdateQuestStatus(
		getStringParam(step.Params, "campaign"),
		getStringParam(step.Params, "quest_id"),
		domain.QuestStatus(getStringParam(step.Params, "status")),
		getStringParam(step.Params, "notes"),
	)
}

func executeGenerateAdventureBible(ctx context.Context, h *TestHarness, step TestStep) error {
	brief := domain.CampaignBrief{
		Name: getStringParam(step.Params, "campaign_id"),
	}
	_, err := h.CanonService.InitializeCanon(ctx, brief)
	return err
}

func executeUpdateNarrativeState(_ context.Context, h *TestHarness, step TestStep) error {
	return fmt.Errorf("update_narrative_state not yet implemented in E2E harness")
}

func executeCheckConsistency(_ context.Context, h *TestHarness, step TestStep) error {
	return fmt.Errorf("check_consistency not yet implemented in E2E harness")
}

func executeGenerateSessionPrep(_ context.Context, h *TestHarness, step TestStep) error {
	return fmt.Errorf("generate_session_prep not yet implemented in E2E harness")
}
