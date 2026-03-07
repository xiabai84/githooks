package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
	"github.com/xiabai84/githooks/types"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new workspace with githooks",
	Long:  `Add a new workspace with githooks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hooks.CheckConfigFiles(); err != nil {
			return err
		}

		projName, err := types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please provide a name for the workspace.",
			Label:    "Enter your workspace name:",
		}, "")
		if err != nil {
			return err
		}

		jiraName := strings.ToUpper(projName)
		jiraName, err = types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please provide a Jira project key RegEx to track, e.g. ALPHA or (ALPHA|BETA)",
			Label:    fmt.Sprintf("Enter your Jira project key RegEx (%s):", jiraName),
		}, jiraName)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		homeDir, err := os.UserHomeDir()
		if err == nil && len(homeDir) != 0 {
			cwd = strings.Replace(cwd, homeDir, "~", 1)
		}
		if !strings.HasSuffix(cwd, "/") {
			cwd += "/"
		}
		workDir, err := types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please enter a path to your workspace.",
			Label:    fmt.Sprintf("Enter path to your workspace (%s):", cwd),
		}, cwd)
		if err != nil {
			return err
		}

		if !strings.HasSuffix(workDir, "/") {
			workDir += "/"
		}

		newWorkspace := types.Workspace{
			Name:         projName,
			ProjectKeyRE: strings.ToUpper(jiraName),
			Folder:       workDir,
		}
		if err := hooks.PreviewConfig(&newWorkspace); err != nil {
			return err
		}

		prompt := promptui.Prompt{
			Label:     "Input was correct",
			IsConfirm: true,
			Default:   "y",
		}

		_, err = prompt.Run()
		if err != nil {
			fmt.Println(promptui.IconBad + " Canceled adding of a new githooks workspace.")
			return nil
		}

		return hooks.AddWorkspace(&newWorkspace)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
