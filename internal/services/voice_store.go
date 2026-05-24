package services

import (
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/dm"
)

// CampaignVoiceStore provides per-campaign voice registry operations.
type CampaignVoiceStore interface {
	GetVoiceID(campaignID, npcName string) (string, error)
	SetVoicePrompt(campaignID, npcName, voicePrompt string) error
	ListVoices(campaignID string) map[string]string
}

// FileCampaignVoiceStore implements CampaignVoiceStore using disk-backed JSON.
type FileCampaignVoiceStore struct {
	baseDir string
}

// NewFileCampaignVoiceStore creates a new disk-backed campaign voice store.
func NewFileCampaignVoiceStore(baseDir string) *FileCampaignVoiceStore {
	return &FileCampaignVoiceStore{baseDir: baseDir}
}

func (s *FileCampaignVoiceStore) registryFor(campaignID string) *dm.FileVoiceRegistry {
	dir := filepath.Join(s.baseDir, campaignID)
	return dm.NewVoiceRegistry(dir)
}

// GetVoiceID returns the voice ID for an NPC, or error if not found.
func (s *FileCampaignVoiceStore) GetVoiceID(campaignID, npcName string) (string, error) {
	voiceID := "npc-" + dm.Slugify(npcName)
	reg := s.registryFor(campaignID)
	_, err := reg.GetVoice(voiceID)
	if err != nil {
		return "", err
	}
	return voiceID, nil
}

// SetVoicePrompt creates or updates a voice prompt for an NPC.
func (s *FileCampaignVoiceStore) SetVoicePrompt(campaignID, npcName, voicePrompt string) error {
	voiceID := "npc-" + dm.Slugify(npcName)
	reg := s.registryFor(campaignID)
	entry := dm.NPCVoiceEntry{
		NPCID:       voiceID,
		NPCName:     npcName,
		VoicePrompt: voicePrompt,
		Language:    "es",
		Speed:       1.0,
	}
	return reg.SetVoice(voiceID, entry)
}

// ListVoices returns all assigned voices for a campaign.
func (s *FileCampaignVoiceStore) ListVoices(campaignID string) map[string]string {
	reg := s.registryFor(campaignID)
	entries := reg.ListVoices()
	result := make(map[string]string, len(entries))
	for k, v := range entries {
		result[k] = v.VoicePrompt
	}
	return result
}
