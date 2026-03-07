package hooks

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/xiabai84/githooks/buildinfo"
	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func InitHooks() (types.GitHookConfig, error) {
	ghConfig := types.GitHookConfig{
		Version:    buildinfo.GetBuildInfo().Version,
		Workspaces: []types.Workspace{},
	}

	if err := os.MkdirAll(config.Default.HookDir, config.ExecutableFilePermission); err != nil {
		return ghConfig, fmt.Errorf("creating hook dir: %w", err)
	}
	if err := os.MkdirAll(config.Default.HookConfigDir, config.ExecutableFilePermission); err != nil {
		return ghConfig, fmt.Errorf("creating hook config dir: %w", err)
	}

	if _, err := os.Stat(config.Default.GitConfigPath); err != nil {
		f, err := os.Create(config.Default.GitConfigPath)
		if err != nil {
			return ghConfig, fmt.Errorf("creating git config: %w", err)
		}
		defer f.Close()
		if err := os.Chmod(config.Default.GitConfigPath, config.ConfigFilePermission); err != nil {
			return ghConfig, fmt.Errorf("setting git config permissions: %w", err)
		}
		fmt.Println(promptui.IconGood+"  Created ", config.Default.GitConfigPath)
	}

	_, commitMsgErr := os.Stat(config.Default.CommitMsgPath)
	if err := os.WriteFile(config.Default.CommitMsgPath, []byte(config.CommitMsg), config.ExecutableFilePermission); err != nil {
		return ghConfig, fmt.Errorf("creating commit-msg: %w", err)
	}
	if commitMsgErr != nil {
		fmt.Println(promptui.IconGood+"  Created ", config.Default.CommitMsgPath)
	} else {
		fmt.Println(promptui.IconGood+"  Updated ", config.Default.CommitMsgPath)
	}

	if _, err := os.Stat(config.Default.GithooksConfigPath); err != nil {
		if err := WriteGitHooksConfig(&ghConfig); err != nil {
			return ghConfig, err
		}
		fmt.Println(promptui.IconGood+"  Created ", config.Default.GithooksConfigPath)
	} else {
		// Preserve existing workspaces
		existing, err := ReadGitHooksConfig()
		if err != nil {
			return ghConfig, err
		}
		ghConfig = existing
		ghConfig.Version = buildinfo.GetBuildInfo().Version
		if err := WriteGitHooksConfig(&ghConfig); err != nil {
			return ghConfig, err
		}
		fmt.Println(promptui.IconGood+"  Updated ", config.Default.GithooksConfigPath, "(preserved workspaces)")
	}

	return ghConfig, nil
}
