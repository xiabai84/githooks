package hooks

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func UpdateWorkspace(ghConfig *types.GitHookConfig, idx int, updated *types.Workspace) error {
	old := ghConfig.Workspaces[idx]

	// If folder changed, remove old includeIf block from .gitconfig
	if old.Folder != updated.Folder {
		if err := overwriteGitConfig(&old); err != nil {
			return err
		}
		fmt.Println(promptui.IconGood+"  Modified", config.Default.GitConfigPath, "(removed old includeIf block)")
		// Add new includeIf block
		if err := updateGitConfigFile(updated); err != nil {
			return err
		}
	}

	// If name changed, remove old workspace gitconfig file
	if old.Name != updated.Name {
		oldConfigPath := config.Default.HookConfigDir + "/" + config.GitHooksConfigPrefix + "-" + strings.ToLower(old.Name)
		if err := deleteWorkspaceGitConfig(old.Name); err == nil {
			fmt.Println(promptui.IconGood+"  Deleted ", oldConfigPath)
		}
	}

	// Update workspace in config
	ghConfig.Workspaces[idx] = *updated

	if err := WriteGitHooksConfig(ghConfig); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GithooksConfigPath, "(updated workspace entry)")

	// Rewrite workspace gitconfig file with new values
	if err := createWorkspaceGitConfig(updated); err != nil {
		return err
	}

	fmt.Println(promptui.IconGood + "  Updated workspace " + updated.Name)
	printChanges(&old, updated)
	return nil
}

func printChanges(old, updated *types.Workspace) {
	if old.Name != updated.Name {
		fmt.Printf("  Name:         %s → %s\n", old.Name, updated.Name)
	}
	if old.ProjectKeyRE != updated.ProjectKeyRE {
		fmt.Printf("  Jira keys:    %s → %s\n", old.ProjectKeyRE, updated.ProjectKeyRE)
	}
	if old.Folder != updated.Folder {
		fmt.Printf("  Folder:       %s → %s\n", old.Folder, updated.Folder)
	}
}