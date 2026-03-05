package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func TestAddWorkspace_CreatesAllArtifacts(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create prerequisite: .gitconfig and initial config
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws := &types.Workspace{
		Name:         "TestProject",
		ProjectKeyRE: "TEST",
		Folder:       "~/projects/test/",
	}
	err := AddWorkspace(ws)
	if err != nil {
		t.Fatalf("AddWorkspace returned error: %v", err)
	}

	// Verify workspace was added to config
	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if len(readConfig.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(readConfig.Workspaces))
	}
	if readConfig.Workspaces[0].Name != "TestProject" {
		t.Errorf("expected workspace name 'TestProject', got %q", readConfig.Workspaces[0].Name)
	}

	// Verify workspace-specific git config was created
	wsConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-testproject")
	wsContent, err := os.ReadFile(wsConfigPath)
	if err != nil {
		t.Fatalf("expected workspace config file to exist at %s", wsConfigPath)
	}
	if !strings.Contains(string(wsContent), "jiraProjects=TEST") {
		t.Errorf("workspace config should contain jiraProjects=TEST, got %q", string(wsContent))
	}
	if !strings.Contains(string(wsContent), "hooksPath=~/.githooks") {
		t.Errorf("workspace config should contain hooksPath, got %q", string(wsContent))
	}

	// Verify .gitconfig was updated with includeIf
	gitConfig, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(gitConfig), "includeIf") {
		t.Error("expected .gitconfig to contain includeIf directive")
	}
	if !strings.Contains(string(gitConfig), "~/projects/test/") {
		t.Error("expected .gitconfig to contain workspace folder path")
	}
}

func TestAddWorkspace_MultipleWorkspaces(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws1 := &types.Workspace{Name: "Alpha", ProjectKeyRE: "ALPHA", Folder: "~/alpha/"}
	ws2 := &types.Workspace{Name: "Beta", ProjectKeyRE: "BETA", Folder: "~/beta/"}

	if err := AddWorkspace(ws1); err != nil {
		t.Fatalf("AddWorkspace(Alpha) returned error: %v", err)
	}
	if err := AddWorkspace(ws2); err != nil {
		t.Fatalf("AddWorkspace(Beta) returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if len(readConfig.Workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(readConfig.Workspaces))
	}
}

func TestAddWorkspace_MissingGitConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Write githooks.json but NOT .gitconfig
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws := &types.Workspace{Name: "Fail", ProjectKeyRE: "FAIL", Folder: "~/fail/"}
	err := AddWorkspace(ws)
	if err == nil {
		t.Error("expected error when .gitconfig doesn't exist")
	}
}
