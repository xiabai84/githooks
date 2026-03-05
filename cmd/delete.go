package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
	"github.com/xiabai84/githooks/types"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a githooks workspace and its settings.",
	Long:  `Delete a githooks workspace and its settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hooks.CheckConfigFiles(); err != nil {
			return err
		}

		ghConfig, err := hooks.ReadGitHooksConfig()
		if err != nil {
			return err
		}

		empty := types.Workspace{Name: "Quit"}
		workspaces := append(ghConfig.Workspaces, empty)

		prompt1 := promptui.Select{
			Label:     "Delete:",
			Items:     workspaces,
			Templates: workspaceSelectTemplates(),
			Size:      5,
			Searcher:  workspaceSearcher(workspaces),
		}

		i, _, err := prompt1.Run()
		if err != nil {
			return err
		}

		if workspaces[i].Name == "Quit" {
			return nil
		}

		prompt2 := promptui.Prompt{
			Label:     "Do you Really want to delete this workspace",
			IsConfirm: true,
		}
		confirmed, err := prompt2.Run()
		if err != nil {
			fmt.Println("Canceled")
			return nil
		}
		if strings.ToLower(confirmed) != "y" {
			return nil
		}
		return hooks.DeleteSelectedWorkspace(&ghConfig, i)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
