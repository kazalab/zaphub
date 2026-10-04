package cli

import (
	"github.com/spf13/cobra"

	"github.com/kazalab/zaphub/internal/config"
)

func NewCommand() *cobra.Command {
	var logLevel string

	cmd := &cobra.Command{
		Use:   "zaphub",
		Short: "Zaphub API server",
		Long:  "Zaphub - French TV EPG, cinema releases and YouTube feeds API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon(logLevel, args)
		},
	}

	cmd.Flags().StringVar(&logLevel, "log-level", config.Load().LogLevel, "log level (debug, info, warn, error)")

	cmd.AddCommand(newVersionCommand())

	return cmd
}

func Execute() error {
	return NewCommand().Execute()
}
