package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all githooks managed files and configuration.",
	Long: `Remove all files and configuration managed by githooks:
  - ~/.githooks/ directory (hook script, config files, workspace configs)
  - includeIf blocks from ~/.gitconfig`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("This will remove all githooks managed files:")
		fmt.Println("  - ~/.githooks/ (hook script, workspace configs)")
		fmt.Println("  - includeIf blocks from ~/.gitconfig")
		fmt.Println()

		prompt := newConfirmPrompt("Are you sure you want to uninstall githooks", false)
		_, err := prompt.Run()
		if err != nil {
			fmt.Println("Canceled")
			return nil
		}

		return hooks.Uninstall()
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
