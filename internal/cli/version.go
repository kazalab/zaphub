package cli

import (
	"github.com/juli3nk/go-utils/version"
	"github.com/spf13/cobra"
)

var (
	appVersion   = "unknown-version"
	appCommit    = "unknown-commit"
	appBuildDate = "unknown-builddate"
)

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version and exit",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.New(appVersion, appCommit, appBuildDate)
			v.Show()
		},
	}

	return cmd
}
