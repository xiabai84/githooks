package cmd

import (
	"github.com/stefan-niemeyer/githooks/hooks"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "githooks",
	Short: "githooks helps developers with setting name conventions of a Git commit message",
	Long:  `githooks prevents developer to enter commit messages, which don't contain predefined Jira issue keys.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// cobra already prints the error
		return
	}
}

func init() {
	if err := hooks.MigrateGitHooksConfig(); err != nil {
		cobra.CheckErr(err)
	}
}
