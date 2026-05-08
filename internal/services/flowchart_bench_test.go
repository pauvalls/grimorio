package services

import (
	"context"
	"fmt"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func BenchmarkFlowchartService_buildNodes(b *testing.B) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	svc := NewFlowchartService(canonRepo, nil)

	// Build a large canon document with many entities and relationships
	const numEntities = 500
	const numRelationships = 1000

	entities := make([]domain.CanonEntity, 0, numEntities)
	for i := 0; i < numEntities; i++ {
		entities = append(entities, domain.CanonEntity{
			ID:         fmt.Sprintf("entity-%04d", i),
			Name:       fmt.Sprintf("Entity %d", i),
			Type:       domain.EntityTypeNPC,
			CanonState: domain.EntityStateAlive,
		})
	}

	relationships := make([]domain.CanonRelationship, 0, numRelationships)
	for i := 0; i < numRelationships; i++ {
		src := i % numEntities
		dst := (i + 1) % numEntities
		relationships = append(relationships, domain.CanonRelationship{
			ID:   fmt.Sprintf("rel-%04d", i),
			From: fmt.Sprintf("entity-%04d", src),
			To:   fmt.Sprintf("entity-%04d", dst),
			Type: domain.RelationshipTypeAlly,
		})
	}

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "bench-campaign",
		Entities:      entities,
		Relationships: relationships,
	}
	_ = canonRepo.Save("bench-campaign", doc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.buildNodes(ctx, "bench-campaign", doc, "decision")
	}
}
