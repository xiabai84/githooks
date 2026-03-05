package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/stefan-niemeyer/githooks/buildinfo"
	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

// legacyHook represents the old log-based config format, used only for migration.
type legacyHook struct {
	Project  string
	JiraName string
	WorkDir  string
}

func MigrateGitHooksConfig() error {
	_, err := os.Stat(config.Default.GithooksLogPath)
	if err != nil {
		return nil
	}

	hookArr, err := readFromGitHooksLog()
	if err != nil {
		return fmt.Errorf("reading legacy log: %w", err)
	}

	workspaces := make([]types.Workspace, 0, len(hookArr))
	for _, hook := range hookArr {
		workspaces = append(workspaces, types.Workspace{
			Name:         hook.Project,
			ProjectKeyRE: hook.JiraName,
			Folder:       hook.WorkDir,
		})
	}
	ghConfig := types.GitHookConfig{
		Version:    buildinfo.GetBuildInfo().Version,
		Workspaces: workspaces,
	}
	if err := WriteGitHooksConfig(&ghConfig); err != nil {
		return fmt.Errorf("writing migrated config: %w", err)
	}
	fmt.Printf(promptui.IconGood+" Config file '%s' migrated to '%s'\n", config.Default.GithooksLogPath, config.Default.GithooksConfigPath)

	if err := os.Remove(config.Default.GithooksLogPath); err != nil {
		return fmt.Errorf("removing legacy log: %w", err)
	}
	fmt.Printf(promptui.IconGood+" Old config file '%s' deleted\n", config.Default.GithooksLogPath)
	return nil
}

func WriteGitHooksConfig(ghConfig *types.GitHookConfig) error {
	ghConfig.Version = buildinfo.GetBuildInfo().Version
	configJSON, err := json.Marshal(ghConfig)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(config.Default.GithooksConfigPath, configJSON, config.ConfigFilePermission); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

func readFromGitHooksLog() ([]legacyHook, error) {
	var hookArr []legacyHook

	bytesRead, err := os.ReadFile(config.Default.GithooksLogPath)
	if err != nil {
		return nil, fmt.Errorf("reading log file: %w", err)
	}
	lines := strings.Split(string(bytesRead), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		var hook legacyHook
		if err := json.Unmarshal([]byte(line), &hook); err != nil {
			return nil, fmt.Errorf("parsing log entry: %w", err)
		}
		hookArr = append(hookArr, hook)
	}

	return hookArr, nil
}

func ReadGitHooksConfig() (types.GitHookConfig, error) {
	bytesRead, err := os.ReadFile(config.Default.GithooksConfigPath)
	if err != nil {
		return types.GitHookConfig{}, fmt.Errorf("reading config file: %w", err)
	}
	ghConfig := types.GitHookConfig{}
	if err := json.Unmarshal(bytesRead, &ghConfig); err != nil {
		return types.GitHookConfig{}, fmt.Errorf("parsing config file: %w", err)
	}
	return ghConfig, nil
}
