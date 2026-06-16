package compiler

import "context"

// Exporter defines the strategy interface for campaign export formats.
type Exporter interface {
	Export(ctx context.Context, campaignDir, title string) (string, error)
	Format() string
}
