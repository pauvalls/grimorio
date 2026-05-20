package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDMContextPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		payload DMContextPayload
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid payload",
			payload: DMContextPayload{
				CampaignID:  "test-campaign",
				SessionNum:  1,
				GeneratedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing campaign_id",
			payload: DMContextPayload{
				SessionNum: 1,
			},
			wantErr: true,
			errMsg:  "campaign ID is required",
		},
		{
			name: "invalid campaign_id format",
			payload: DMContextPayload{
				CampaignID: "La Llave",
				SessionNum: 1,
			},
			wantErr: true,
			errMsg:  "campaign ID must be kebab-case",
		},
		{
			name: "session_num too low",
			payload: DMContextPayload{
				CampaignID: "test-campaign",
				SessionNum: 0,
			},
			wantErr: true,
			errMsg:  "session number must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDMContextPayload_JSONRoundTrip(t *testing.T) {
	original := DMContextPayload{
		CampaignID:  "test-campaign",
		SessionNum:  3,
		GeneratedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Canon: &CanonContext{
			Facts: []CanonFact{
				{ID: "f1", Category: "lore", Statement: "Magic is banned", Source: "canon"},
			},
			Entities: []CanonEntity{
				{ID: "npc-1", Name: "Eldrin", Type: EntityTypeNPC, CanonState: EntityStateAlive},
			},
		},
		NarrativeState: &NarrativeContext{
			CurrentSession: 2,
			ActiveQuests: []QuestState{
				{ID: "q1", Name: "Main Quest", Status: "active", SourceAct: "act-1"},
			},
		},
		SessionPrep: &DMContextSessionPrep{
			PreviouslyOn: "Session 2: Found the key.",
			ActiveQuests: []string{"Main Quest"},
		},
		Characters: []CharacterContext{
			{Name: "Aric", Race: "humano", Class: "guerrero", Level: 3, Background: "soldado", Alignment: "LG"},
		},
		Areas: map[string]AreaContext{
			"area_1_1": {ID: "area_1_1", ChapterID: "chapter_1", AreaNumber: 1, Title: "Entrada"},
		},
		NPCs: map[string]NPCContext{
			"Eldrin": {Name: "Eldrin", Description: "Un mago anciano", Motivation: "proteger la torre", DialogueVoice: "Habla en susurros"},
		},
		Bestiary: map[string]MonsterContext{
			"goblin": {
				Name: "Goblin",
				CR:   "1/4",
				AC:   15,
				HP:   7,
				DescriptiveCues: map[string]string{
					"full_hp":  "El goblin está alerta y listo para pelear.",
					"half_hp":  "El goblin sangra pero sigue en pie.",
					"low_hp":   "El goblin se tambalea, apenas consciente.",
					"defeated": "El goblin cae al suelo, inmóvil.",
				},
			},
		},
		Factions: map[string]FactionContext{
			"thieves-guild": {ID: "thieves-guild", Name: "Thieves' Guild", Reputation: -30, Status: "unfriendly", Attitude: "hostile"},
		},
		Quests: []QuestContext{
			{ID: "q1", Title: "Main Quest", Status: "active", Type: "main", Giver: "Eldrin"},
		},
		PDFAvailable: false,
		DMNotes: DMNotesContext{
			Warnings:  []string{"NPC-1 is dead in canon"},
			Reminders: []string{"Award XP for session 2"},
		},
	}

	bytes, err := json.Marshal(original)
	assert.NoError(t, err)

	var decoded DMContextPayload
	err = json.Unmarshal(bytes, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, original.CampaignID, decoded.CampaignID)
	assert.Equal(t, original.SessionNum, decoded.SessionNum)
	assert.Equal(t, original.Canon.Facts[0].Statement, decoded.Canon.Facts[0].Statement)
	assert.Equal(t, original.Characters[0].Name, decoded.Characters[0].Name)
	assert.Equal(t, original.Areas["area_1_1"].Title, decoded.Areas["area_1_1"].Title)
	assert.Equal(t, original.NPCs["Eldrin"].DialogueVoice, decoded.NPCs["Eldrin"].DialogueVoice)
	assert.Equal(t, original.Bestiary["goblin"].CR, decoded.Bestiary["goblin"].CR)
	assert.Equal(t, original.Bestiary["goblin"].DescriptiveCues["half_hp"], decoded.Bestiary["goblin"].DescriptiveCues["half_hp"])
	assert.Equal(t, original.Factions["thieves-guild"].Reputation, decoded.Factions["thieves-guild"].Reputation)
	assert.Equal(t, original.Quests[0].Giver, decoded.Quests[0].Giver)
	assert.Equal(t, original.DMNotes.Warnings[0], decoded.DMNotes.Warnings[0])
}

func TestMonsterContext_HasAllDescriptiveCues(t *testing.T) {
	tests := []struct {
		name    string
		monster MonsterContext
		want    bool
	}{
		{
			name: "all cues present",
			monster: MonsterContext{
				Name: "Goblin",
				DescriptiveCues: map[string]string{
					"full_hp":  "alert",
					"half_hp":  "wounded",
					"low_hp":   "dying",
					"defeated": "dead",
				},
			},
			want: true,
		},
		{
			name: "missing one cue",
			monster: MonsterContext{
				Name: "Goblin",
				DescriptiveCues: map[string]string{
					"full_hp": "alert",
					"half_hp": "wounded",
					"low_hp":  "dying",
				},
			},
			want: false,
		},
		{
			name:    "empty cues",
			monster: MonsterContext{Name: "Goblin"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.monster.HasAllDescriptiveCues()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMonsterContext_EmptyCues(t *testing.T) {
	monster := MonsterContext{
		Name:            "Skeleton",
		CR:              "1/4",
		AC:              13,
		HP:              13,
		DescriptiveCues: map[string]string{},
	}

	assert.False(t, monster.HasAllDescriptiveCues())
	assert.Empty(t, monster.DescriptiveCues)
}

func TestNPCContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		npc     NPCContext
		wantErr bool
	}{
		{
			name:    "valid NPC",
			npc:     NPCContext{Name: "Eldrin", Description: "Mago"},
			wantErr: false,
		},
		{
			name:    "missing name",
			npc:     NPCContext{Description: "Mago"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.npc.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCharacterContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		char    CharacterContext
		wantErr bool
	}{
		{
			name:    "valid character",
			char:    CharacterContext{Name: "Aric", Level: 5},
			wantErr: false,
		},
		{
			name:    "missing name",
			char:    CharacterContext{Level: 5},
			wantErr: true,
		},
		{
			name:    "level too high",
			char:    CharacterContext{Name: "Aric", Level: 25},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.char.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAreaContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		area    AreaContext
		wantErr bool
	}{
		{
			name:    "valid area",
			area:    AreaContext{ID: "area_1", AreaNumber: 5},
			wantErr: false,
		},
		{
			name:    "missing id",
			area:    AreaContext{AreaNumber: 5},
			wantErr: true,
		},
		{
			name:    "area number out of range",
			area:    AreaContext{ID: "area_1", AreaNumber: 20},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.area.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQuestContext_Validate(t *testing.T) {
	tests := []struct {
		name    string
		quest   QuestContext
		wantErr bool
	}{
		{
			name:    "valid quest",
			quest:   QuestContext{ID: "q1", Title: "Main Quest"},
			wantErr: false,
		},
		{
			name:    "missing id",
			quest:   QuestContext{Title: "Main Quest"},
			wantErr: true,
		},
		{
			name:    "missing title",
			quest:   QuestContext{ID: "q1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.quest.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
