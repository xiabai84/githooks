package hooks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func DeleteSelectedWorkspace(ghConfig *types.GitHookConfig, idx int) error {
	removedWorkspace := ghConfig.Workspaces[idx].Name
	if err := overwriteGitConfig(&ghConfig.Workspaces[idx]); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GitConfigPath, "(removed includeIf block)")
	ghConfig.Workspaces = append(ghConfig.Workspaces[:idx], ghConfig.Workspaces[idx+1:]...)
	if err := WriteGitHooksConfig(ghConfig); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Modified", config.Default.GithooksConfigPath, "(removed workspace entry)")
	wsConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-"+strings.ToLower(removedWorkspace))
	if err := deleteWorkspaceGitConfig(removedWorkspace); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Deleted ", wsConfigPath)
	fmt.Println(promptui.IconGood+"  Removed workspace", removedWorkspace)
	return nil
}

func overwriteGitConfig(workspace *types.Workspace) error {
	bytesRead, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		return fmt.Errorf("reading git config: %w", err)
	}

	lines := strings.Split(string(bytesRead), "\n")
	var result []string
	wsConfigSuffix := config.GitHooksConfigPrefix + "-" + strings.ToLower(workspace.Name)

	for i := 0; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		// Match includeIf line with the workspace's folder
		if strings.Contains(trimmed, "[includeIf") && strings.Contains(trimmed, "gitdir:"+workspace.Folder) {
			// Check if next line is the path line for this workspace
			if i+1 < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(nextTrimmed, "path") && strings.Contains(nextTrimmed, wsConfigSuffix) {
					i++ // skip both lines
					continue
				}
			}
		}
		result = append(result, lines[i])
	}

	newContent := strings.Join(result, "\n")
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(newContent), config.ConfigFilePermission); err != nil {
		return fmt.Errorf("writing git config: %w", err)
	}
	return nil
}

func deleteWorkspaceGitConfig(wsName string) error {
	configPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-"+strings.ToLower(wsName))
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("removing workspace config %s: %w", configPath, err)
	}
	return nil
}
