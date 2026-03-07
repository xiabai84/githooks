package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
)

var checkBranch string
var checkBranchName string

var checkCmd = &cobra.Command{
	Use:   "check [message]",
	Short: "Validate a commit message or branch name",
	Long: `Validates a commit message against Conventional Commits format and Jira ticket rules,
or validates a branch name against naming conventions. Useful for CI pipelines and debugging.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projects := getProjects()

		// Branch name validation mode
		if checkBranchName != "" || cmd.Flags().Changed("branch-name") {
			branch := checkBranchName
			if branch == "" {
				// Read current branch from git
				out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
				if err != nil {
					fmt.Fprintln(os.Stderr, "✘ Could not determine current branch (detached HEAD?)")
					os.Exit(1)
				}
				branch = strings.TrimSpace(string(out))
			}
			result := hooks.CheckBranchName(branch, projects)
			if result.Valid {
				fmt.Println("✔ Valid branch:", result.Message)
			} else {
				fmt.Fprintln(os.Stderr, "✘ Invalid:", result.Error)
				os.Exit(1)
			}
			return
		}

		// Commit message validation mode (requires argument)
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: requires a commit message argument or --branch-name flag")
			os.Exit(1)
		}
		msg := args[0]
		result := hooks.CheckCommitMessage(msg, projects, checkBranch)
		if result.Valid {
			fmt.Println("✔ Valid:", result.Message)
		} else {
			fmt.Fprintln(os.Stderr, "✘ Invalid:", result.Error)
			os.Exit(1)
		}
	},
}

func getProjects() string {
	ghConfig, err := hooks.ReadGitHooksConfig()
	if err != nil || len(ghConfig.Workspaces) == 0 {
		return ""
	}
	idx, _ := hooks.GetWorkspaceIndex(ghConfig.Workspaces)
	return ghConfig.Workspaces[idx].ProjectKeyRE
}

func init() {
	checkCmd.Flags().StringVar(&checkBranch, "branch", "", "Simulate branch name for commit message auto-injection (e.g. feat/PROJ-123-login)")
	checkCmd.Flags().StringVar(&checkBranchName, "branch-name", "", "Validate a branch name (omit value to check current branch)")
	checkCmd.Flag("branch-name").NoOptDefVal = ""
	rootCmd.AddCommand(checkCmd)
}
