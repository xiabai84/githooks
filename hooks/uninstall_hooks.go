package hooks

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/xiabai84/githooks/config"
)

func Uninstall() error {
	// Remove all includeIf blocks managed by githooks from .gitconfig
	if err := removeGithooksFromGitConfig(); err != nil {
		return err
	}

	// Remove the entire ~/.githooks directory
	if err := os.RemoveAll(config.Default.HookDir); err != nil {
		return fmt.Errorf("removing %s: %w", config.Default.HookDir, err)
	}
	fmt.Println(promptui.IconGood + "  Removed " + config.Default.HookDir)

	fmt.Println(promptui.IconGood + "  githooks has been fully uninstalled")
	return nil
}

func removeGithooksFromGitConfig() error {
	data, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading git config: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	var result []string
	githooksPath := config.GitHooksFolder + "/" + config.GitHooksConfigFolder + "/"

	for i := 0; i < len(lines); i++ {
		// Check if this is an includeIf block followed by a path referencing .githooks/config/
		if strings.Contains(lines[i], "[includeIf") && i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])
			if strings.HasPrefix(nextLine, "path = ") && strings.Contains(nextLine, githooksPath) {
				i++ // skip both the includeIf line and the path line
				continue
			}
		}
		result = append(result, lines[i])
	}

	cleaned := strings.Join(result, "\n")
	if cleaned != content {
		if err := os.WriteFile(config.Default.GitConfigPath, []byte(cleaned), config.ConfigFilePermission); err != nil {
			return fmt.Errorf("writing git config: %w", err)
		}
		fmt.Println(promptui.IconGood + "  Cleaned includeIf blocks from " + config.Default.GitConfigPath)
	}

	return nil
}
