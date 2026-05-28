package campaign

import (
	"github.com/urfave/cli/v2"
)

// NewCampaignCommand returns the parent "campaign" CLI command with all subcommands.
func NewCampaignCommand() *cli.Command {
	return &cli.Command{
		Name:  "campaign",
		Usage: "Campaign management commands",
		Subcommands: []*cli.Command{
			NewListCommand(),
			NewDiffCommand(),
			NewExportCommand(),
			NewImportCommand(),
			NewRegisterCommand(),
		},
	}
}
