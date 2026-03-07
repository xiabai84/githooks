package cmd

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
	"github.com/xiabai84/githooks/types"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing workspace (Jira keys, folder, or name).",
	Long:  `Interactively update an existing workspace's Jira project keys, workspace folder, or name.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hooks.CheckConfigFiles(); err != nil {
			return err
		}

		ghConfig, err := hooks.ReadGitHooksConfig()
		if err != nil {
			return err
		}

		if len(ghConfig.Workspaces) == 0 {
			fmt.Println("No workspaces configured. Use 'githooks add' to create one.")
			return nil
		}

		quit := types.Workspace{Name: "Quit"}
		workspaces := append(ghConfig.Workspaces, quit)

		selectPrompt := promptui.Select{
			Label:     "Select workspace to update:",
			Items:     workspaces,
			Templates: workspaceSelectTemplates(),
			Size:      5,
			Searcher:  workspaceSearcher(workspaces),
		}

		i, _, err := selectPrompt.Run()
		if err != nil {
			return err
		}

		if workspaces[i].Name == "Quit" {
			return nil
		}

		current := ghConfig.Workspaces[i]

		// Prompt for new name
		newName, err := types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please provide a name for the workspace.",
			Label:    fmt.Sprintf("Workspace name (%s):", current.Name),
		}, current.Name)
		if err != nil {
			return err
		}

		// Prompt for new Jira keys
		newKeys, err := types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please provide a Jira project key RegEx.",
			Label:    fmt.Sprintf("Jira project key RegEx (%s):", current.ProjectKeyRE),
		}, current.ProjectKeyRE)
		if err != nil {
			return err
		}

		// Prompt for new folder
		defaultFolder := current.Folder
		newFolder, err := types.GetPromptInput(types.Dialog{
			ErrorMsg: "Please enter a path to your workspace.",
			Label:    fmt.Sprintf("Workspace folder (%s):", defaultFolder),
		}, defaultFolder)
		if err != nil {
			return err
		}
		if !strings.HasSuffix(newFolder, "/") {
			newFolder += "/"
		}

		updated := types.Workspace{
			Name:         newName,
			ProjectKeyRE: strings.ToUpper(newKeys),
			Folder:       newFolder,
		}

		// Check if anything changed
		if current.Name == updated.Name &&
			current.ProjectKeyRE == updated.ProjectKeyRE &&
			current.Folder == updated.Folder {
			fmt.Println("No changes made.")
			return nil
		}

		// Preview changes
		fmt.Println()
		fmt.Println("Changes:")
		if current.Name != updated.Name {
			fmt.Printf("  Name:         %s → %s\n", current.Name, updated.Name)
		}
		if current.ProjectKeyRE != updated.ProjectKeyRE {
			fmt.Printf("  Jira keys:    %s → %s\n", current.ProjectKeyRE, updated.ProjectKeyRE)
		}
		if current.Folder != updated.Folder {
			fmt.Printf("  Folder:       %s → %s\n", current.Folder, updated.Folder)
		}
		fmt.Println()

		confirmPrompt := promptui.Prompt{
			Label:     "Apply changes",
			IsConfirm: true,
			Default:   "y",
		}
		_, err = confirmPrompt.Run()
		if err != nil {
			fmt.Println("Canceled")
			return nil
		}

		return hooks.UpdateWorkspace(&ghConfig, i, &updated)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
