package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func TestUpdateWorkspace_PrintsFileOperations_ChangeJiraKeys(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "Alpha", ProjectKeyRE: "(ALPHA|BETA)", Folder: "~/work/alpha/"}

	var err error
	output := captureStdout(t, func() {
		err = UpdateWorkspace(ghConfig, 0, updated)
	})

	if err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	// Should mention modifying githooks.json
	if !strings.Contains(output, "Modified") || !strings.Contains(output, config.Default.GithooksConfigPath) {
		t.Errorf("expected output to mention modifying %s, got:\n%s", config.Default.GithooksConfigPath, output)
	}

	// Should mention creating/modifying workspace gitconfig
	wsConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpha")
	if !strings.Contains(output, wsConfigPath) {
		t.Errorf("expected output to mention %s, got:\n%s", wsConfigPath, output)
	}

	// Should mention updated workspace
	if !strings.Contains(output, "Updated workspace") {
		t.Errorf("expected output to mention updating workspace, got:\n%s", output)
	}
}

func TestUpdateWorkspace_PrintsFileOperations_ChangeFolder(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "Alpha", ProjectKeyRE: "ALPHA", Folder: "~/work/new-alpha/"}

	var err error
	output := captureStdout(t, func() {
		err = UpdateWorkspace(ghConfig, 0, updated)
	})

	if err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	// Should mention modifying .gitconfig (removed old + added new includeIf)
	if !strings.Contains(output, "Modified") || !strings.Contains(output, config.Default.GitConfigPath) {
		t.Errorf("expected output to mention modifying %s, got:\n%s", config.Default.GitConfigPath, output)
	}
}

func TestUpdateWorkspace_PrintsFileOperations_ChangeName(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "AlphaRenamed", ProjectKeyRE: "ALPHA", Folder: "~/work/alpha/"}

	var err error
	output := captureStdout(t, func() {
		err = UpdateWorkspace(ghConfig, 0, updated)
	})

	if err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	// Should mention deleting old workspace gitconfig
	oldPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpha")
	if !strings.Contains(output, "Deleted") || !strings.Contains(output, oldPath) {
		t.Errorf("expected output to mention deleting %s, got:\n%s", oldPath, output)
	}

	// Should mention creating new workspace gitconfig
	newPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpharenamed")
	if !strings.Contains(output, "Created") || !strings.Contains(output, newPath) {
		t.Errorf("expected output to mention creating %s, got:\n%s", newPath, output)
	}
}

func setupUpdateTest(t *testing.T) (*types.GitHookConfig, func()) {
	t.Helper()
	cleanup := setupTestConfig(t)

	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws := &types.Workspace{Name: "Alpha", ProjectKeyRE: "ALPHA", Folder: "~/work/alpha/"}
	if err := AddWorkspace(ws); err != nil {
		t.Fatalf("AddWorkspace returned error: %v", err)
	}

	ghConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}

	return &ghConfig, cleanup
}

func TestUpdateWorkspace_ChangeJiraKeys(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "Alpha", ProjectKeyRE: "(ALPHA|BETA)", Folder: "~/work/alpha/"}
	if err := UpdateWorkspace(ghConfig, 0, updated); err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if readConfig.Workspaces[0].ProjectKeyRE != "(ALPHA|BETA)" {
		t.Errorf("expected (ALPHA|BETA), got %q", readConfig.Workspaces[0].ProjectKeyRE)
	}

	// Verify gitconfig file has updated keys
	wsConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpha")
	content, err := os.ReadFile(wsConfigPath)
	if err != nil {
		t.Fatalf("failed to read workspace config: %v", err)
	}
	if !strings.Contains(string(content), "jiraProjects=(ALPHA|BETA)") {
		t.Errorf("expected jiraProjects=(ALPHA|BETA) in config, got %q", string(content))
	}
}

func TestUpdateWorkspace_ChangeFolder(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "Alpha", ProjectKeyRE: "ALPHA", Folder: "~/work/new-alpha/"}
	if err := UpdateWorkspace(ghConfig, 0, updated); err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if readConfig.Workspaces[0].Folder != "~/work/new-alpha/" {
		t.Errorf("expected ~/work/new-alpha/, got %q", readConfig.Workspaces[0].Folder)
	}

	// Verify .gitconfig has the new folder path
	gitConfig, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("failed to read git config: %v", err)
	}
	if !strings.Contains(string(gitConfig), "~/work/new-alpha/") {
		t.Error("expected .gitconfig to contain new folder path")
	}
	if strings.Contains(string(gitConfig), "~/work/alpha/") {
		t.Error("expected .gitconfig to NOT contain old folder path")
	}
}

func TestUpdateWorkspace_ChangeName(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "AlphaRenamed", ProjectKeyRE: "ALPHA", Folder: "~/work/alpha/"}
	if err := UpdateWorkspace(ghConfig, 0, updated); err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if readConfig.Workspaces[0].Name != "AlphaRenamed" {
		t.Errorf("expected AlphaRenamed, got %q", readConfig.Workspaces[0].Name)
	}

	// New gitconfig file should exist
	newPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpharenamed")
	if _, err := os.Stat(newPath); err != nil {
		t.Errorf("expected new config file at %s", newPath)
	}

	// Old gitconfig file should be removed
	oldPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-alpha")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old config file %s to be deleted", oldPath)
	}
}

func TestUpdateWorkspace_ChangeAll(t *testing.T) {
	ghConfig, cleanup := setupUpdateTest(t)
	defer cleanup()

	updated := &types.Workspace{Name: "Beta", ProjectKeyRE: "(BETA|GAMMA)", Folder: "~/work/beta/"}
	if err := UpdateWorkspace(ghConfig, 0, updated); err != nil {
		t.Fatalf("UpdateWorkspace returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	ws := readConfig.Workspaces[0]
	if ws.Name != "Beta" || ws.ProjectKeyRE != "(BETA|GAMMA)" || ws.Folder != "~/work/beta/" {
		t.Errorf("unexpected workspace: %+v", ws)
	}

	// Verify gitconfig file
	wsConfigPath := filepath.Join(config.Default.HookConfigDir, config.GitHooksConfigPrefix+"-beta")
	content, err := os.ReadFile(wsConfigPath)
	if err != nil {
		t.Fatalf("failed to read workspace config: %v", err)
	}
	if !strings.Contains(string(content), "jiraProjects=(BETA|GAMMA)") {
		t.Errorf("expected jiraProjects=(BETA|GAMMA), got %q", string(content))
	}

	// Verify .gitconfig
	gitConfig, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("failed to read git config: %v", err)
	}
	if !strings.Contains(string(gitConfig), "~/work/beta/") {
		t.Error("expected .gitconfig to contain ~/work/beta/")
	}
	if strings.Contains(string(gitConfig), "~/work/alpha/") {
		t.Error("expected .gitconfig to NOT contain ~/work/alpha/")
	}
}
