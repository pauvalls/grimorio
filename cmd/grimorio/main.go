package main

import (
	"fmt"
	"os"
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
		campaign.NewCampaignCommand(),
			update.NewUpdateCommand(version),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
