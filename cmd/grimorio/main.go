package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pauvalls/grimorio/cmd/grimorio/commands"
	"github.com/pauvalls/grimorio/cmd/grimorio/commands/campaign"
	"github.com/pauvalls/grimorio/cmd/grimorio/commands/update"
	"github.com/urfave/cli/v2"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := &cli.App{
		Name:    "grimorio",
		Usage:   "Grimorio TTRPG campaign manager",
		Version: fmt.Sprintf("%s (commit: %s, build date: %s, go: %s)", version, commit, date, runtime.Version()),
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "compiler-version",
				Usage: "Compiler version (1 or 2)",
			},
		},
		// Default action: start MCP server (backward compatible)
		Action: commands.RunMCPServer,
		Commands: []*cli.Command{
			{
				Name:  "mcp",
				Usage: "Start the MCP stdio server",
				Action: func(cCtx *cli.Context) error {
					return commands.RunMCPServer(cCtx)
				},
			},
			commands.NewValidateCommandWithEngines(defaultCampaignsDir()),
	campaign.NewCampaignCommand(),
		func() *cli.Command {
			cmd := update.NewUpdateCommand(version)
			cmd.Subcommands = []*cli.Command{
				update.NewUpdateSkillsCommand(),
				update.NewUpdateAgentsCommand(),
				update.NewUpdateCommandsCommand(),
				update.NewUpdateAllCommand(),
			}
			return cmd
		}(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// defaultCampaignsDir returns the canonical campaigns base directory,
// honouring CAMPAIGN_ROOT for tests / portable installs.
func defaultCampaignsDir() string {
	if root := os.Getenv("CAMPAIGN_ROOT"); root != "" {
		return root
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "campaigns")
	}
	return filepath.Join(home, "campaigns")
}
