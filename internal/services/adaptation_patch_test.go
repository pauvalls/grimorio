package services

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestAdaptationPatch_GeneratePatch(t *testing.T) {
	ctx := context.Background()
	actRepo := repository.NewMemoryActRepository()
	canonRepo := repository.NewMemoryCanonRepository()

	// Save acts and check for errors
	act1 := &domain.Act{
		CampaignID:        "test-campaign",
		Number:            1,
		Title:             "Act One",
		Content:           "The merchants-guild controls the trade routes.",
		GameMode:          "investigacion",
		ChapterObjectives: []string{"Investigar el gremio de mercaderes", "Identificar al líder"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "mystery",
		RunningGuidance:   "Este capítulo presenta el misterio inicial de la campaña y establece las bases para toda la aventura. Los personajes jugadores deben investigar el gremio de mercaderes y descubrir quién está detrás de las actividades sospechosas que amenazan la estabilidad económica de la ciudad. Hay tres pistas principales que deben encontrar durante su investigación: registros financieros alterados que muestran transacciones inusuales con cuentas en el extranjero, testigos que vieron actividades nocturnas en el almacén del gremio durante las últimas semanas, y un documento comprometedor que vincula directamente a miembros del gremio con operaciones de contrabando. Si los personajes se estancan en algún punto de la investigación, un NPC contacta con información adicional para mantener el ritmo de la aventura. El tono es de misterio con elementos de intriga política que se desarrollan gradualmente. Asegúrate de que cada sesión termine con una revelación importante que motive a los jugadores a continuar la investigación en la siguiente sesión.",
		AssetHandoff:      "Los registros financieros revelan conexión con el Acto 2",
	}
	if err := actRepo.Save(act1); err != nil {
		t.Fatalf("failed to save act 1: %v", err)
	}
	
	act2 := &domain.Act{
		CampaignID:        "test-campaign",
		Number:            2,
		Title:             "Act Two",
		Content:           "The merchants-guild leader is a trusted ally.",
		GameMode:          "intriga",
		ChapterObjectives: []string{"Confrontar al líder del gremio", "Decidir si aliarse o exponerlo"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "political",
		RunningGuidance:   "Este capítulo es de intriga política donde los personajes jugadores deben decidir cómo manejar la corrupción descubierta durante su investigación previa. Hay múltiples caminos disponibles y cada uno tiene consecuencias significativas para el desarrollo de la campaña: pueden aliarse con el líder a cambio de favores y recursos valiosos, exponerlo públicamente arriesgando consecuencias políticas graves que afectarán a toda la ciudad, o eliminarlo silenciosamente manteniendo el secreto pero asumiendo la responsabilidad moral. Cada decisión afecta la reputación con diferentes facciones de la ciudad de manera permanente y duradera. El Gremio reacciona según la elección tomada por los personajes: si se alían obtienen recursos económicos y contactos valiosos en los bajos fondos, si exponen el Gremio se vuelve hostil y busca venganza contra los responsables, si eliminan hay un vacío de poder que otras facciones intentan llenar rápidamente. El tono es tenso con decisiones morales complejas que no tienen respuestas correctas obvias para los jugadores. Las consecuencias de estas decisiones se manifestarán en el siguiente capítulo de manera significativa.",
		AssetHandoff:      "La decisión tomada determina qué facciones ayudan en el Acto 3",
	}
	if err := actRepo.Save(act2); err != nil {
		t.Fatalf("failed to save act 2: %v", err)
	}
	
	// Verify acts were saved
	if _, err := actRepo.List("test-campaign"); err != nil {
		t.Fatalf("failed to list acts: %v", err)
	}

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "merchants-guild", Name: "Merchants Guild", Type: domain.EntityTypeFaction},
		},
	}
	if err := canonRepo.Save("test-campaign", doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	svc := NewAdaptationPatchService(actRepo, canonRepo)

	t.Run("faction coup patch", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "merchants-guild",
			TriggerType: "faction-coup",
			Description: "The merchants-guild has undergone a coup",
			SessionNum:  3,
		}
		patch, err := svc.GeneratePatch(ctx, "test-campaign", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if patch.IsEmpty {
			t.Fatalf("expected non-empty patch")
		}
		if !strings.Contains(patch.MarkdownDiff, "merchants-guild") {
			t.Fatalf("expected markdown diff to reference merchants-guild")
		}
		if len(patch.AffectedActs) != 2 {
			t.Fatalf("expected 2 affected acts, got %d", len(patch.AffectedActs))
		}
	})

	t.Run("no matches returns empty patch", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "nonexistent-faction",
			TriggerType: "faction-coup",
			Description: "A coup in a nonexistent faction",
			SessionNum:  3,
		}
		patch, err := svc.GeneratePatch(ctx, "test-campaign", event)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !patch.IsEmpty {
			t.Fatalf("expected empty patch")
		}
	})

	t.Run("idempotent generation", func(t *testing.T) {
		event := domain.WorldEvent{
			ID:          "merchants-guild",
			TriggerType: "faction-coup",
			Description: "Coup",
			SessionNum:  3,
		}
		patch1, _ := svc.GeneratePatch(ctx, "test-campaign", event)
		patch2, _ := svc.GeneratePatch(ctx, "test-campaign", event)
		if patch1.MarkdownDiff != patch2.MarkdownDiff {
			t.Fatalf("expected idempotent patch generation")
		}
	})
}
