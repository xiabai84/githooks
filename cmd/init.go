package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stefan-niemeyer/githooks/hooks"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Setup your githooks workspace configuration",
	Long:  `Setup your githooks workspace configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := hooks.InitHooks()
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
