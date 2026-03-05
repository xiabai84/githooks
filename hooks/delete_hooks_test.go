package hooks

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func setupDeleteTest(t *testing.T) (*types.GitHookConfig, func()) {
	t.Helper()
	cleanup := setupTestConfig(t)

	// Create a .gitconfig with includeIf entries for two workspaces
	ws1 := types.Workspace{Name: "Alpha", ProjectKeyRE: "ALPHA", Folder: "~/work/alpha/"}
	ws2 := types.Workspace{Name: "Beta", ProjectKeyRE: "BETA", Folder: "~/work/beta/"}

	ghConfig := &types.GitHookConfig{
		Version:    "1.0.0",
		Workspaces: []types.Workspace{ws1, ws2},
	}

	// Write the JSON config
	err := WriteGitHooksConfig(ghConfig)
	if err != nil {
		t.Fatalf("WriteGitHooksConfig returned error: %v", err)
	}

	// Create workspace git config files
	for _, ws := range ghConfig.Workspaces {
		configPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-"+strings.ToLower(ws.Name))
		if err := os.WriteFile(configPath, []byte("[core]\n    hooksPath=~/.githooks\n"), 0644); err != nil {
			t.Fatalf("failed to write workspace config: %v", err)
		}
	}

	// Build .gitconfig content with includeIf blocks
	var gitConfigContent strings.Builder
	gitConfigContent.WriteString("[user]\n    name = Test\n")
	for _, ws := range ghConfig.Workspaces {
		gitConfigContent.WriteString("[includeIf \"gitdir:" + ws.Folder + "\"]\n")
		gitConfigContent.WriteString("    path = " + config.GitHooksFolder + "/" + config.GitHooksConfigFolder + "/" + config.GitHooksConfigPrefix + "-" + strings.ToLower(ws.Name) + "\n")
	}
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(gitConfigContent.String()), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	return ghConfig, cleanup
}

func TestDeleteSelectedWorkspace_RemovesWorkspace(t *testing.T) {
	ghConfig, cleanup := setupDeleteTest(t)
	defer cleanup()

	err := DeleteSelectedWorkspace(ghConfig, 0) // delete Alpha
	if err != nil {
		t.Fatalf("DeleteSelectedWorkspace returned error: %v", err)
	}

	// Config should now have only Beta
	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if len(readConfig.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace remaining, got %d", len(readConfig.Workspaces))
	}
	if readConfig.Workspaces[0].Name != "Beta" {
		t.Errorf("expected remaining workspace 'Beta', got %q", readConfig.Workspaces[0].Name)
	}
}

func TestDeleteSelectedWorkspace_RemovesGitConfigFile(t *testing.T) {
	ghConfig, cleanup := setupDeleteTest(t)
	defer cleanup()

	alphaConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpha")

	// Verify it exists before delete
	if _, err := os.Stat(alphaConfigPath); err != nil {
		t.Fatalf("expected %s to exist before delete", alphaConfigPath)
	}

	err := DeleteSelectedWorkspace(ghConfig, 0)
	if err != nil {
		t.Fatalf("DeleteSelectedWorkspace returned error: %v", err)
	}

	// Verify workspace config file was deleted
	if _, err := os.Stat(alphaConfigPath); !os.IsNotExist(err) {
		t.Errorf("expected %s to be deleted", alphaConfigPath)
	}
}

func TestDeleteSelectedWorkspace_UpdatesGitConfig(t *testing.T) {
	ghConfig, cleanup := setupDeleteTest(t)
	defer cleanup()

	err := DeleteSelectedWorkspace(ghConfig, 0) // delete Alpha
	if err != nil {
		t.Fatalf("DeleteSelectedWorkspace returned error: %v", err)
	}

	gitConfigContent, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	content := string(gitConfigContent)

	if strings.Contains(content, "alpha") {
		t.Error("expected .gitconfig to no longer contain 'alpha' includeIf block")
	}
	if !strings.Contains(content, "beta") {
		t.Error("expected .gitconfig to still contain 'beta' includeIf block")
	}
}

func TestDeleteSelectedWorkspace_GitConfigPermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions are not supported on Windows")
	}
	ghConfig, cleanup := setupDeleteTest(t)
	defer cleanup()

	err := DeleteSelectedWorkspace(ghConfig, 0)
	if err != nil {
		t.Fatalf("DeleteSelectedWorkspace returned error: %v", err)
	}

	info, err := os.Stat(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("os.Stat returned error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0644 {
		t.Errorf("expected .gitconfig permission 0644, got %04o", perm)
	}
}
