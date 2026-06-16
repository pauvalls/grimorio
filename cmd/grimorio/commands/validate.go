package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/urfave/cli/v2"
)

// runValidate is the CLI action function for `grimorio validate`.
// It returns a `*cli.Exit` error to control the process exit code:
//   - 0 → success (clean campaign)
//   - 1 → validation found errors or criticals
//   - 2 → usage error (missing arg, invalid scope)
func runValidate(c *cli.Context, engine *services.ValidationEngine) error {
	stdout := c.App.Writer
	stderr := c.App.ErrWriter
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	if c.NArg() < 1 {
		_, _ = fmt.Fprintln(stderr, "campaign name required")
		_, _ = fmt.Fprintln(stderr, "Usage: grimorio validate <campaign> [--scope=structure|wotc|references|all] [--json]")
		return cli.Exit("campaign name required", 2)
	}

	campaignID := c.Args().First()
	scopeFlag := c.String("scope")
	useJSON := c.Bool("json")

	engineScope, postFilter, err := mapCLIScope(scopeFlag)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintln(stderr, "Valid scopes: structure, wotc, references, all")
		return cli.Exit(err.Error(), 2)
	}

	report, err := engine.CheckConsistency(context.Background(), campaignID, engineScope)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Error: validation failed: %v\n", err)
		return cli.Exit(err.Error(), 2)
	}

	// Post-filter for narrow scopes (wotc / references).
	if len(postFilter) > 0 {
		filterReport(report, postFilter)
	}

	// Render.
	if useJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if encErr := enc.Encode(report); encErr != nil {
			_, _ = fmt.Fprintf(stderr, "Error: failed to encode JSON: %v\n", encErr)
			return cli.Exit(encErr.Error(), 2)
		}
	} else {
		renderTextReport(stdout, report)
	}

	// Map report to exit code (0/1).
	if report.Criticals > 0 || report.Errors > 0 {
		return cli.Exit("validation failed", 1)
	}
	return nil
}

// renderTextReport writes a human-readable summary to w.
func renderTextReport(w io.Writer, report *domain.ConsistencyReport) {
	_, _ = fmt.Fprintf(w, "Campaign Validation Report\n")
	_, _ = fmt.Fprintf(w, "==========================\n")
	_, _ = fmt.Fprintf(w, "Campaign: %s\n", report.CampaignID)
	_, _ = fmt.Fprintf(w, "Health: %s\n", report.OverallHealth)
	_, _ = fmt.Fprintln(w)

	if len(report.Issues) == 0 {
		_, _ = fmt.Fprintln(w, "✅ All checks passed")
		return
	}

	for _, issue := range report.Issues {
		marker := "❌"
		if issue.Passed {
			marker = "✅"
		}
		fmt.Fprintf(w, "%s [%s] %s — %s\n", marker, issue.Severity, issue.Rule, issue.Message)
	}

		_, _ = fmt.Fprintln(w)
	fmt.Fprintf(w, "Summary: %d errors, %d warnings, %d criticals (of %d checks)\n",
		report.Errors, report.Warnings, report.Criticals, report.TotalChecks)
}

// matchesAny returns true if name has any of the given prefixes.
func matchesAny(name string, prefixes []string) bool {
	for _, p := range prefixes {
		if len(name) >= len(p) && name[:len(p)] == p {
			return true
		}
	}
	return false
}

// filterReport keeps only issues whose Rule matches one of the prefixes and
// recomputes the aggregate stats (TotalChecks / Passed / Warnings / Errors /
// Criticals / OverallHealth).
func filterReport(report *domain.ConsistencyReport, prefixes []string) {
	filtered := report.Issues[:0]
	for _, issue := range report.Issues {
		if matchesAny(issue.Rule, prefixes) {
			filtered = append(filtered, issue)
		}
	}
	report.Issues = filtered

	report.TotalChecks = len(report.Issues)
	report.Passed, report.Warnings, report.Errors, report.Criticals = 0, 0, 0, 0
	for _, issue := range report.Issues {
		if issue.Passed {
			report.Passed++
			continue
		}
		switch issue.Severity {
		case "critical":
			report.Criticals++
		case "error":
			report.Errors++
		case "warning":
			report.Warnings++
		}
	}

	switch {
	case report.Criticals > 0:
		report.OverallHealth = "critical"
	case report.Errors > 0:
		report.OverallHealth = "poor"
	case report.Warnings > 0:
		report.OverallHealth = "fair"
	case report.TotalChecks > 0:
		report.OverallHealth = "good"
	default:
		report.OverallHealth = "excellent"
	}
}

// mapCLIScope translates a CLI scope string to the engine's ConsistencyScope
// plus an optional rule-name post-filter (for wotc / references).
func mapCLIScope(cliScope string) (domain.ConsistencyScope, []string, error) {
	switch cliScope {
	case "", "all":
		return domain.ConsistencyScopeFull, nil, nil
	case "structure":
		return domain.ConsistencyScopeLoreOnly, nil, nil
	case "wotc":
		return domain.ConsistencyScopeFull, []string{"wotc_"}, nil
	case "references":
		return domain.ConsistencyScopeFull, []string{"integration"}, nil
	default:
		return "", nil, fmt.Errorf("invalid scope %q: valid options are structure, wotc, references, all", cliScope)
	}
}

// NewValidateCommand creates the `grimorio validate` top-level command.
//
// engine is the validation engine (injected for testability).
// baseDir is currently unused at runtime; the engine handles its own baseDir.
func NewValidateCommand(engine *services.ValidationEngine, baseDir string) *cli.Command {
	return &cli.Command{
		Name:      "validate",
		Usage:     "Validate a campaign for consistency, WotC format, and references",
		ArgsUsage: "<campaign>",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "scope",
				Aliases: []string{"s"},
				Usage:   "Validation scope: structure | wotc | references | all (default: all)",
				Value:   "all",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Emit machine-readable JSON (domain.ConsistencyReport shape)",
			},
		},
		Action: func(c *cli.Context) error {
			_ = baseDir // kept for future per-campaign loading
			return runValidate(c, engine)
		},
	}
}

// NewValidateCommandWithEngines is a variant that builds the ValidationEngine
// from filesystem repositories rooted at baseDir. Used by main.go for the
// real CLI surface. Tests should use NewValidateCommand directly with a
// pre-built (often in-memory) engine.
func NewValidateCommandWithEngines(baseDir string) *cli.Command {
	canonRepo := repository.NewFilesystemCanonRepository(baseDir)
	stateRepo := repository.NewFilesystemNarrativeStateRepository(baseDir)
	canonSvc := services.NewCanonService(canonRepo, stateRepo, nil)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)
	engine := services.NewValidationEngine(canonSvc, stateSvc, nil, baseDir)
	return NewValidateCommand(engine, baseDir)
}
