package hooks

import (
	"fmt"

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
		// Add new includeIf block
		if err := updateGitConfigFile(updated); err != nil {
			return err
		}
	}

	// If name changed, remove old workspace gitconfig file
	if old.Name != updated.Name {
		// Ignore error if old file doesn't exist
		_ = deleteWorkspaceGitConfig(old.Name)
	}

	// Update workspace in config
	ghConfig.Workspaces[idx] = *updated

	if err := WriteGitHooksConfig(ghConfig); err != nil {
		return err
	}

	// Rewrite workspace gitconfig file with new values
	if err := createWorkspaceGitConfig(updated); err != nil {
		return err
	}

	// If folder unchanged but Jira keys changed, no gitconfig update needed
	// (workspace gitconfig already rewritten above)

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

// RenameWorkspaceGitConfig handles renaming the gitconfig file in .gitconfig
// when the workspace name changes but the folder stays the same.
func RenameWorkspaceGitConfig(old, updated *types.Workspace) error {
	if old.Name == updated.Name {
		return nil
	}

	// Remove old includeIf and add new one
	if err := overwriteGitConfig(old); err != nil {
		return err
	}

	// Re-add with the new name (folder is the same, so gitdir matches)
	tmpWs := *updated
	return updateGitConfigFile(&tmpWs)
}

// GetWorkspaceGitConfigPath returns the path to a workspace's gitconfig file.
func GetWorkspaceGitConfigPath(wsName string) string {
	return config.Default.HookConfigDir + "/" + config.GitHooksConfigPrefix + "-" + wsName
}
