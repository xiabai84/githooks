package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
	"github.com/xiabai84/githooks/types"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workspaces managed by githooks.",
	Long:  `List all workspaces managed by githooks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := hooks.CheckConfigFiles(); err != nil {
			return err
		}

		ghConfig, err := hooks.ReadGitHooksConfig()
		if err != nil {
			return err
		}

		empty := types.Workspace{Name: "Quit"}
		preselectIdx, err := hooks.GetWorkspaceIndex(ghConfig.Workspaces)
		if err != nil {
			return err
		}

		ghConfig.Workspaces = append(ghConfig.Workspaces, empty)

		prompt := promptui.Select{
			Label:     "Active githooks workspaces:",
			Items:     ghConfig.Workspaces,
			Templates: workspaceSelectTemplates(),
			Size:      5,
			Searcher:  workspaceSearcher(ghConfig.Workspaces),
			CursorPos: preselectIdx,
		}

		_, _, err = prompt.Run()
		return err
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
