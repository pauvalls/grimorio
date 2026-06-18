package consolidation

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// CampaignConsolidator orchestrates all consolidation analyzers and applies safe fixes.
type CampaignConsolidator struct {
	baseDir    string
	reader     CampaignReader
	applicator FixApplicator
}

// NewCampaignConsolidator creates an orchestrator with a filesystem reader.
func NewCampaignConsolidator(baseDir string) *CampaignConsolidator {
	return &CampaignConsolidator{
		baseDir:    baseDir,
		reader:     &filesystemReader{baseDir: baseDir},
		applicator: &defaultApplicator{},
	}
}

// NewCampaignConsolidatorWithReader creates an orchestrator with a custom reader.
func NewCampaignConsolidatorWithReader(baseDir string, reader CampaignReader) *CampaignConsolidator {
	if reader == nil {
		reader = &filesystemReader{baseDir: baseDir}
	}
	return &CampaignConsolidator{
		baseDir:    baseDir,
		reader:     reader,
		applicator: &defaultApplicator{},
	}
}

// SetApplicator sets a custom fix applicator (mainly for tests).
func (c *CampaignConsolidator) SetApplicator(a FixApplicator) {
	c.applicator = a
}

// Detect runs all analyzers in read-only mode.
func (c *CampaignConsolidator) Detect(ctx context.Context, campaignID string) (*domain.ConsolidationReport, error) {
	return c.Consolidate(ctx, campaignID, domain.ConsolidationOptions{AutoFix: false})
}

// ReadFiles returns the campaign markdown files without running any analyzer.
// Useful for callers that want to dispatch to a specific analyzer (e.g. the
// ValidationEngine and CampaignHealthCheck integration layer).
func (c *CampaignConsolidator) ReadFiles(ctx context.Context, campaignID string) ([]CampaignFile, error) {
	return c.reader.ReadCampaign(ctx, campaignID)
}

// Consolidate runs analyzers and applies safe fixes when AutoFix is enabled.
func (c *CampaignConsolidator) Consolidate(ctx context.Context, campaignID string, opts domain.ConsolidationOptions) (*domain.ConsolidationReport, error) {
	files, err := c.reader.ReadCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to read campaign: %w", err)
	}

	report := &domain.ConsolidationReport{
		CampaignID: campaignID,
		Metadata: map[string]interface{}{
			"auto_fix": opts.AutoFix,
			"read_at":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	analyzers := c.buildAnalyzers(campaignID)
	for _, a := range analyzers {
		res, err := a.Analyze(ctx, files)
		if err != nil {
			report.ChecksRun = append(report.ChecksRun, domain.ConsolidationCheck{
				Rule:     a.Name(),
				Passed:   false,
				Severity: "error",
				Message:  fmt.Sprintf("Analyzer %s failed: %v", a.Name(), err),
			})
			continue
		}
		report.ChecksRun = append(report.ChecksRun, c.toDomainCheck(a.Name(), res))
		for _, issue := range res.Issues {
			report.RemainingIssues = append(report.RemainingIssues, c.toDomainIssue(issue))
		}
		for _, q := range res.Questions {
			report.NeedsHuman = append(report.NeedsHuman, c.toDomainQuestion(q))
		}
	}

	if opts.AutoFix {
		if err := c.applySafeFixes(ctx, campaignID, report, files, opts); err != nil {
			return nil, fmt.Errorf("failed to apply safe fixes: %w", err)
		}
	}

	fresh, err := c.VerifyFreshness(ctx, campaignID)
	if err == nil {
		report.Freshness = fresh
	}

	return report, nil
}

// ResolveAmbiguity applies a user decision for a known ambiguity.
func (c *CampaignConsolidator) ResolveAmbiguity(ctx context.Context, campaignID string, qid string, decision string) error {
	files, err := c.reader.ReadCampaign(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to read campaign: %w", err)
	}

	if strings.HasPrefix(qid, "entity-") {
		resolver := NewEntityResolver(0.85)
		if err := c.applyEntityDecision(ctx, campaignID, resolver, files, decision); err != nil {
			return fmt.Errorf("failed to apply entity decision: %w", err)
		}
	}

	return nil
}

// RegenerateIndex writes an INDEX.md with breadcrumbs and verified links.
func (c *CampaignConsolidator) RegenerateIndex(ctx context.Context, campaignID string) error {
	files, err := c.reader.ReadCampaign(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to read campaign: %w", err)
	}

	campaignDir := filepath.Join(c.baseDir, campaignID)
	indexPath := filepath.Join(campaignDir, "INDEX.md")

	var sources []string
	for _, f := range files {
		base := filepath.Base(f.RelPath)
		if strings.EqualFold(base, "INDEX.md") || strings.EqualFold(base, "index.md") {
			continue
		}
		if strings.HasSuffix(f.RelPath, ".md") {
			sources = append(sources, f.RelPath)
		}
	}
	sort.Strings(sources)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("# %s — Campaign Index\n\n", campaignID))
	b.WriteString("## Breadcrumbs\n\n")
	b.WriteString(fmt.Sprintf("Campaigns / %s / INDEX\n\n", campaignID))
	b.WriteString("## Verified Links\n\n")
	for _, src := range sources {
		b.WriteString(fmt.Sprintf("- [%s](%s)\n", src, src))
	}

	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		return fmt.Errorf("failed to create campaign directory: %w", err)
	}
	if err := os.WriteFile(indexPath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("failed to write INDEX.md: %w", err)
	}
	return nil
}

// VerifyFreshness compares generated files against source files.
func (c *CampaignConsolidator) VerifyFreshness(ctx context.Context, campaignID string) (*domain.FreshnessReport, error) {
	files, err := c.reader.ReadCampaign(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to read campaign: %w", err)
	}

	campaignDir := filepath.Join(c.baseDir, campaignID)
	report := &domain.FreshnessReport{CampaignID: campaignID}

	var sourceNewest time.Time
	for _, f := range files {
		base := filepath.Base(f.RelPath)
		if strings.EqualFold(base, "campaign.md") || strings.EqualFold(base, "INDEX.md") || strings.EqualFold(base, "index.md") {
			continue
		}
		if f.ModTime.After(sourceNewest) {
			sourceNewest = f.ModTime
		}
	}
	report.SourcesNewest = sourceNewest

	campaignMD := filepath.Join(campaignDir, "campaign.md")
	if info, err := os.Stat(campaignMD); err == nil {
		report.CampaignMDTime = info.ModTime()
		report.CampaignMDStale = info.ModTime().Before(sourceNewest)
	}

	indexMD := filepath.Join(campaignDir, "INDEX.md")
	if info, err := os.Stat(indexMD); err == nil {
		report.IndexMDTime = info.ModTime()
		report.IndexStale = info.ModTime().Before(sourceNewest)
	}

	switch {
	case report.CampaignMDStale && report.IndexStale:
		report.Message = "campaign.md and INDEX.md are stale relative to sources"
	case report.CampaignMDStale:
		report.Message = "campaign.md is stale relative to sources"
	case report.IndexStale:
		report.Message = "INDEX.md is stale relative to sources"
	default:
		report.Message = "Generated files are up to date"
	}

	return report, nil
}

func (c *CampaignConsolidator) buildAnalyzers(campaignID string) []Analyzer {
	return []Analyzer{
		NewEntityResolver(0.85),
		NewLoreCoherence(),
		NewStatBlockResolver(),
		NewEventCanonizer(),
		NewFileConsolidator(),
		NewMapReferenceChecker(filepath.Join(c.baseDir, campaignID)),
	}
}

func (c *CampaignConsolidator) toDomainCheck(name string, res *AnalysisResult) domain.ConsolidationCheck {
	sev := res.Severity
	if sev == "" {
		sev = "info"
	}
	return domain.ConsolidationCheck{
		Rule:      name,
		Passed:    res.Passed,
		Severity:  sev,
		Message:   res.Message,
		Locations: res.Locations,
	}
}

func (c *CampaignConsolidator) toDomainIssue(issue domainIssue) domain.ConsolidationIssue {
	return domain.ConsolidationIssue{
		Rule:       issue.Rule,
		Severity:   issue.Severity,
		Message:    issue.Message,
		Locations:  issue.Locations,
		Suggestion: issue.Suggestion,
	}
}

func (c *CampaignConsolidator) toDomainQuestion(q domainQuestion) domain.AmbiguityQuestion {
	return domain.AmbiguityQuestion{
		ID:       q.ID,
		Rule:     q.Rule,
		Question: q.Question,
		Options:  q.Options,
		Context:  q.Context,
	}
}

func (c *CampaignConsolidator) applySafeFixes(ctx context.Context, campaignID string, report *domain.ConsolidationReport, files []CampaignFile, opts domain.ConsolidationOptions) error {
	campaignDir := filepath.Join(c.baseDir, campaignID)
	backupDir := c.backupDir(opts)
	if backupDir != "" {
		if err := c.writeBackup(campaignDir, backupDir, files); err != nil {
			return fmt.Errorf("failed to write backup: %w", err)
		}
	}

	// Entity renames above threshold.
	resolver := NewEntityResolver(0.85)
	canonicalMap := resolver.CanonicalMap(files)
	for variant, canonical := range canonicalMap {
		fixed, err := c.renameEntityInFiles(files, variant, canonical)
		if err != nil {
			return err
		}
		if fixed > 0 {
			report.FixesApplied = append(report.FixesApplied, domain.ConsolidationFix{
				Rule:   "entity_name_uniqueness",
				Target: variant,
				Before: variant,
				After:  canonical,
			})
		}
	}

	// Duplicate file deletion.
	removed, err := RemoveDuplicateFiles(files)
	if err != nil {
		return err
	}
	for _, path := range removed {
		report.FixesApplied = append(report.FixesApplied, domain.ConsolidationFix{
			Rule:   "duplicate_file",
			Target: path,
			Before: path,
			After:  "removed",
		})
	}

	return nil
}

func (c *CampaignConsolidator) renameEntityInFiles(files []CampaignFile, variant, canonical string) (int, error) {
	fixed := 0
	for i := range files {
		content := files[i].Content
		newContent := strings.ReplaceAll(content, variant, canonical)
		if newContent == content {
			continue
		}
		if err := os.WriteFile(files[i].Path, []byte(newContent), 0644); err != nil {
			return fixed, fmt.Errorf("failed to rename entity in %s: %w", files[i].Path, err)
		}
		files[i].Content = newContent
		fixed++
	}
	return fixed, nil
}

func (c *CampaignConsolidator) applyEntityDecision(ctx context.Context, campaignID string, resolver *EntityResolver, files []CampaignFile, decision string) error {
	for variant, canonical := range resolver.CanonicalMap(files) {
		if decision == canonical {
			_, err := c.renameEntityInFiles(files, variant, canonical)
			return err
		}
	}
	return nil
}

func (c *CampaignConsolidator) backupDir(opts domain.ConsolidationOptions) string {
	if opts.BackupDir != "" {
		return opts.BackupDir
	}
	return filepath.Join(".consolidation", "backups", time.Now().UTC().Format("20060102-150405"))
}

func (c *CampaignConsolidator) writeBackup(campaignDir, backupDir string, files []CampaignFile) error {
	fullBackup := filepath.Join(campaignDir, backupDir)
	if err := os.MkdirAll(fullBackup, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}
	for _, f := range files {
		dest := filepath.Join(fullBackup, f.RelPath)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return fmt.Errorf("failed to create backup parent: %w", err)
		}
		if err := copyFile(f.Path, dest); err != nil {
			return fmt.Errorf("failed to backup %s: %w", f.RelPath, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// filesystemReader reads campaign markdown files from disk.
type filesystemReader struct {
	baseDir string
}

func (r *filesystemReader) ReadCampaign(ctx context.Context, campaignID string) ([]CampaignFile, error) {
	campaignDir := filepath.Join(r.baseDir, campaignID)
	info, err := os.Stat(campaignDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat campaign directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("campaign path is not a directory: %s", campaignDir)
	}

	var files []CampaignFile
	err = filepath.WalkDir(campaignDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			rel, _ := filepath.Rel(campaignDir, path)
			if rel == ".consolidation" || rel == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			return nil
		}
		rel, _ := filepath.Rel(campaignDir, path)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info %s: %w", path, err)
		}
		files = append(files, CampaignFile{
			Path:    path,
			RelPath: rel,
			Content: string(content),
			ModTime: info.ModTime(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// defaultApplicator applies fixes when no custom applicator is provided.
type defaultApplicator struct{}

func (a *defaultApplicator) ApplyFix(ctx context.Context, campaignID string, fix domainFix, files []CampaignFile) ([]CampaignFile, error) {
	return files, nil
}
