package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemActRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	// Create campaign directory first
	campaignDir := filepath.Join(tmpDir, "campaign-1")
	if err := os.MkdirAll(filepath.Join(campaignDir, "areas"), 0755); err != nil {
		t.Fatal(err)
	}

	act := &domain.Act{
		CampaignID:        "campaign-1",
		Number:            1,
		Title:             "The Beginning",
		Content:           "Once upon a time...",
		GameMode:          "investigacion",
		ChapterObjectives: []string{"Descubrir la identidad del traidor", "Recuperar el artefacto robado"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "mystery",
		RunningGuidance:   "Este capítulo introduce a los personajes jugadores en el misterio central de la campaña. Comienza con una escena social en la que un NPC aliado presenta el hook principal de la aventura. La investigación debe llevar a los personajes a través de tres ubicaciones clave: primero el distrito comercial donde pueden hablar con mercaderes y recolectar rumores, luego los archivos del gremio donde pueden encontrar documentos importantes, y finalmente la taberna del puerto donde un contacto les proporciona información crucial. Cada ubicación debe revelar una pista diferente que avance la trama principal. Si los personajes se estancan en algún punto, usa un encuentro aleatorio o haz que un NPC contacte con información adicional. El ritmo debe ser moderado, permitiendo tiempo para exploración y roleo, pero con tensión creciente a medida que se acerca el final del capítulo. Asegúrate de que cada sesión termine con un cliffhanger o revelación importante que motive a los jugadores a continuar.",
		AssetHandoff:      "La carta encontrada revela la ubicación del almacén en el Acto 2",
	}

	// Test Save
	if err := repo.Save(act); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Read
	read, err := repo.Read("campaign-1", 1)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Number != 1 {
		t.Errorf("Read() number = %d, want 1", read.Number)
	}

	// Test Read not found
	_, err = repo.Read("campaign-1", 999)
	if err == nil {
		t.Error("Read() should error for nonexistent act")
	}

	// Test List
	list, err := repo.List("campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List() len = %d, want 1", len(list))
	}

	// Test Save with header prefix
	act2 := &domain.Act{
		CampaignID:        "campaign-1",
		Number:            2,
		Title:             "The Middle",
		Content:           "# The Middle\n\nContent here",
		GameMode:          "sandbox_urbano",
		GameModeSecondary: "intriga",
		ChapterObjectives: []string{"Establecer contacto con la facción rebelde", "Evitar la guerra entre gremios"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "political",
		RunningGuidance:   "Este capítulo es de sandbox urbano donde los personajes jugadores tienen libertad completa para explorar la ciudad e interactuar con sus habitantes. Hay cuatro facciones principales con las que pueden interactuar: el Gremio de Mercaderes controla el comercio y tiene recursos económicos, la Guardia de la Ciudad mantiene el orden pero está corrupta, los Contrabandistas del Puerto operan en las sombras y tienen información valiosa, y los Magos del Círculo poseen conocimiento arcano. Cada facción tiene sus propios objetivos, recursos y contactos. Los personajes pueden completar los objetivos en cualquier orden, pero deben asegurarse de que al menos dos facciones queden en posición favorable al final del capítulo para mantener el equilibrio de poder. Si los personajes ignoran por completo una facción, esa facción tomará medidas drásticas que se manifestarán como consecuencias en el siguiente capítulo. El tono es de intriga política con oportunidades variadas para combate táctico, operaciones de sigilo, y negociación compleja. Las decisiones que tomen los personajes aquí tendrán repercusiones duraderas.",
		AssetHandoff:      "El sello del gremio obtenido permite acceso al distrito noble en el Acto 3",
	}
	if err := repo.Save(act2); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", 1); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 1 {
		t.Errorf("List() after delete len = %d, want 1", len(list))
	}
}

func TestFilesystemActRepository_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	invalid := &domain.Act{CampaignID: "", Number: 0}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid act")
	}
}

func TestFilesystemActRepository_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	// Create campaign directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "campaign-1", "areas"), 0755); err != nil {
		t.Fatal(err)
	}

	list, err := repo.List("campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List() len = %d, want 0", len(list))
	}
}
