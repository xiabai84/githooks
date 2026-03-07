package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
)

var checkBranch string

var checkCmd = &cobra.Command{
	Use:   "check <message>",
	Short: "Validate a commit message without committing",
	Long:  `Validates a commit message against Conventional Commits format and Jira ticket rules. Useful for CI pipelines and debugging.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		msg := args[0]

		// Try to determine Jira project filter from current workspace
		projects := ""
		ghConfig, err := hooks.ReadGitHooksConfig()
		if err == nil && len(ghConfig.Workspaces) > 0 {
			idx, _ := hooks.GetWorkspaceIndex(ghConfig.Workspaces)
			projects = ghConfig.Workspaces[idx].ProjectKeyRE
		}

		result := hooks.CheckCommitMessage(msg, projects, checkBranch)
		if result.Valid {
			fmt.Println("✔ Valid:", result.Message)
		} else {
			fmt.Fprintln(os.Stderr, "✘ Invalid:", result.Error)
			os.Exit(1)
		}
	},
}

func init() {
	checkCmd.Flags().StringVar(&checkBranch, "branch", "", "Simulate branch name for auto-injection (e.g. feat/PROJ-123-login)")
	rootCmd.AddCommand(checkCmd)
}
