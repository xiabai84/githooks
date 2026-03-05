package hooks

import (
	"encoding/json"
	"os"
	"runtime"
	"testing"

	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func setupTestConfig(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	origDefault := config.Default
	config.Default = config.NewPaths(tmpDir)
	if err := os.MkdirAll(config.Default.HookConfigDir, 0755); err != nil {
		t.Fatalf("failed to create hook config dir: %v", err)
	}
	return func() { config.Default = origDefault }
}

func TestWriteAndReadGitHooksConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	ghConfig := &types.GitHookConfig{
		Version: "1.0.0",
		Workspaces: []types.Workspace{
			{Name: "test-project", ProjectKeyRE: "TEST", Folder: "~/projects/test/"},
		},
	}

	err := WriteGitHooksConfig(ghConfig)
	if err != nil {
		t.Fatalf("WriteGitHooksConfig returned error: %v", err)
	}

	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}

	if len(readConfig.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(readConfig.Workspaces))
	}
	if readConfig.Workspaces[0].Name != "test-project" {
		t.Errorf("expected workspace name 'test-project', got %q", readConfig.Workspaces[0].Name)
	}
	if readConfig.Workspaces[0].ProjectKeyRE != "TEST" {
		t.Errorf("expected project key 'TEST', got %q", readConfig.Workspaces[0].ProjectKeyRE)
	}
}

func TestWriteGitHooksConfig_FilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions are not supported on Windows")
	}
	cleanup := setupTestConfig(t)
	defer cleanup()

	ghConfig := &types.GitHookConfig{Version: "1.0.0"}
	err := WriteGitHooksConfig(ghConfig)
	if err != nil {
		t.Fatalf("WriteGitHooksConfig returned error: %v", err)
	}

	info, err := os.Stat(config.Default.GithooksConfigPath)
	if err != nil {
		t.Fatalf("os.Stat returned error: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0644 {
		t.Errorf("expected file permission 0644, got %04o", perm)
	}
}

func TestWriteGitHooksConfig_ValidJSON(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	ghConfig := &types.GitHookConfig{
		Version: "1.0.0",
		Workspaces: []types.Workspace{
			{Name: "alpha", ProjectKeyRE: "ALPHA", Folder: "~/alpha/"},
			{Name: "beta", ProjectKeyRE: "(BETA|GAMMA)", Folder: "~/beta/"},
		},
	}
	err := WriteGitHooksConfig(ghConfig)
	if err != nil {
		t.Fatalf("WriteGitHooksConfig returned error: %v", err)
	}

	data, err := os.ReadFile(config.Default.GithooksConfigPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}

	var parsed types.GitHookConfig
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("written file is not valid JSON: %v", err)
	}
	if len(parsed.Workspaces) != 2 {
		t.Errorf("expected 2 workspaces, got %d", len(parsed.Workspaces))
	}
}

func TestReadGitHooksConfig_MissingFile(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	_, err := ReadGitHooksConfig()
	if err == nil {
		t.Error("expected error when config file doesn't exist")
	}
}
