package hooks

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func DeleteSelectedWorkspace(ghConfig *types.GitHookConfig, idx int) error {
	removedWorkspace := ghConfig.Workspaces[idx].Name
	if err := overwriteGitConfig(&ghConfig.Workspaces[idx]); err != nil {
		return err
	}
	ghConfig.Workspaces = append(ghConfig.Workspaces[:idx], ghConfig.Workspaces[idx+1:]...)
	if err := WriteGitHooksConfig(ghConfig); err != nil {
		return err
	}
	if err := deleteWorkspaceGitConfig(removedWorkspace); err != nil {
		return err
	}
	fmt.Println(promptui.IconGood+"  Removed workspace", removedWorkspace)
	return nil
}

func overwriteGitConfig(workspace *types.Workspace) error {
	bytesRead, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		return fmt.Errorf("reading git config: %w", err)
	}
	var partToReplace bytes.Buffer
	tmpl, err := template.New("original").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(config.GitConfigPatch)
	if err != nil {
		return fmt.Errorf("parsing git config patch template: %w", err)
	}
	if err := tmpl.Execute(&partToReplace, workspace); err != nil {
		return fmt.Errorf("executing git config patch template: %w", err)
	}
	newGitConfigContent := strings.Replace(string(bytesRead), partToReplace.String(), "", -1)
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(newGitConfigContent), config.ConfigFilePermission); err != nil {
		return fmt.Errorf("writing git config: %w", err)
	}
	return nil
}

func deleteWorkspaceGitConfig(wsName string) error {
	configPath := config.Default.HookConfigDir + "/" + config.GitHooksConfigPrefix + "-" + strings.ToLower(wsName)
	if err := os.Remove(configPath); err != nil {
		return fmt.Errorf("removing workspace config %s: %w", configPath, err)
	}
	return nil
}
