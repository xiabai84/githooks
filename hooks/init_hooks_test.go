package hooks

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
)

func TestInitHooks_PrintsFileOperations_FirstRun(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	var err error
	output := captureStdout(t, func() {
		_, err = InitHooks()
	})

	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	// Should mention creating .gitconfig
	if !strings.Contains(output, "Created") || !strings.Contains(output, config.Default.GitConfigPath) {
		t.Errorf("expected output to mention creating %s, got:\n%s", config.Default.GitConfigPath, output)
	}

	// Should mention creating commit-msg (first run)
	if !strings.Contains(output, "Created") || !strings.Contains(output, config.Default.CommitMsgPath) {
		t.Errorf("expected output to mention creating %s, got:\n%s", config.Default.CommitMsgPath, output)
	}

	// Should mention creating githooks.json
	if !strings.Contains(output, "Created") || !strings.Contains(output, config.Default.GithooksConfigPath) {
		t.Errorf("expected output to mention creating %s, got:\n%s", config.Default.GithooksConfigPath, output)
	}
}

func TestInitHooks_PrintsFileOperations_SecondRun(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// First init
	captureStdout(t, func() {
		_, _ = InitHooks()
	})

	// Second init
	var err error
	output := captureStdout(t, func() {
		_, err = InitHooks()
	})

	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	// Should mention updating commit-msg
	if !strings.Contains(output, "Updated") || !strings.Contains(output, config.Default.CommitMsgPath) {
		t.Errorf("expected output to mention updating %s, got:\n%s", config.Default.CommitMsgPath, output)
	}

	// Should mention updating githooks.json (preserved workspaces)
	if !strings.Contains(output, "Updated") || !strings.Contains(output, config.Default.GithooksConfigPath) {
		t.Errorf("expected output to mention updating %s, got:\n%s", config.Default.GithooksConfigPath, output)
	}

	// .gitconfig already exists, should NOT say Created for .gitconfig
	// (it should not appear as Created on second run)
}

func TestInitHooks_CreatesAllFiles(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	ghConfig, err := InitHooks()
	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	if ghConfig.Version == "" {
		t.Error("expected non-empty version in config")
	}
	if len(ghConfig.Workspaces) != 0 {
		t.Errorf("expected 0 workspaces, got %d", len(ghConfig.Workspaces))
	}

	// Verify all files and directories were created
	paths := []string{
		config.Default.HookDir,
		config.Default.HookConfigDir,
		config.Default.GitConfigPath,
		config.Default.CommitMsgPath,
		config.Default.GithooksConfigPath,
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist, got error: %v", p, err)
		}
	}
}

func TestInitHooks_GitConfigPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions are not supported on Windows")
	}
	cleanup := setupTestConfig(t)
	defer cleanup()

	_, err := InitHooks()
	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	info, err := os.Stat(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("os.Stat returned error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0644 {
		t.Errorf("expected .gitconfig permission 0644, got %04o", perm)
	}
}

func TestInitHooks_CommitMsgIsExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file permissions are not supported on Windows")
	}
	cleanup := setupTestConfig(t)
	defer cleanup()

	_, err := InitHooks()
	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	info, err := os.Stat(config.Default.CommitMsgPath)
	if err != nil {
		t.Fatalf("os.Stat returned error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0755 {
		t.Errorf("expected commit-msg permission 0755, got %04o", perm)
	}
}

func TestInitHooks_CommitMsgContainsBashShebang(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	_, err := InitHooks()
	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	content, err := os.ReadFile(config.Default.CommitMsgPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.HasPrefix(string(content), "#!/usr/bin/env bash") {
		t.Error("commit-msg should start with bash shebang")
	}
}

func TestInitHooks_DoesNotOverwriteExistingFiles(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create git config with custom content before init
	if err := os.WriteFile(config.Default.GitConfigPath, []byte("custom content"), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	_, err := InitHooks()
	if err != nil {
		t.Fatalf("InitHooks returned error: %v", err)
	}

	// Existing file should not be overwritten
	content, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) != "custom content" {
		t.Errorf("existing .gitconfig was overwritten, got %q", string(content))
	}
}

func TestInitHooks_Idempotent(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	_, err := InitHooks()
	if err != nil {
		t.Fatalf("first InitHooks returned error: %v", err)
	}

	_, err = InitHooks()
	if err != nil {
		t.Fatalf("second InitHooks returned error: %v", err)
	}
}

func TestInitHooks_PreservesExistingWorkspaces(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// First init creates empty config
	_, err := InitHooks()
	if err != nil {
		t.Fatalf("first InitHooks returned error: %v", err)
	}

	// Add a workspace
	ws := &types.Workspace{Name: "TestWS", ProjectKeyRE: "TEST", Folder: "~/work/test/"}
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	if err := AddWorkspace(ws); err != nil {
		t.Fatalf("AddWorkspace returned error: %v", err)
	}

	// Run init again
	ghConfig, err := InitHooks()
	if err != nil {
		t.Fatalf("second InitHooks returned error: %v", err)
	}

	// Workspaces must be preserved
	if len(ghConfig.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace, got %d", len(ghConfig.Workspaces))
	}
	if ghConfig.Workspaces[0].Name != "TestWS" {
		t.Errorf("expected workspace name TestWS, got %q", ghConfig.Workspaces[0].Name)
	}
	if ghConfig.Workspaces[0].ProjectKeyRE != "TEST" {
		t.Errorf("expected project key TEST, got %q", ghConfig.Workspaces[0].ProjectKeyRE)
	}

	// Verify githooks.json on disk
	readConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if len(readConfig.Workspaces) != 1 {
		t.Fatalf("expected 1 workspace on disk, got %d", len(readConfig.Workspaces))
	}
}

func TestInitHooks_AlwaysUpdatesCommitMsg(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// First init
	_, err := InitHooks()
	if err != nil {
		t.Fatalf("first InitHooks returned error: %v", err)
	}

	// Overwrite commit-msg with stale content
	if err := os.WriteFile(config.Default.CommitMsgPath, []byte("old hook"), 0755); err != nil {
		t.Fatalf("failed to write commit-msg: %v", err)
	}

	// Run init again
	_, err = InitHooks()
	if err != nil {
		t.Fatalf("second InitHooks returned error: %v", err)
	}

	// commit-msg must be updated (not the stale content)
	content, err := os.ReadFile(config.Default.CommitMsgPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(content) == "old hook" {
		t.Error("expected commit-msg to be updated, but it still has stale content")
	}
	if !strings.HasPrefix(string(content), "#!/usr/bin/env bash") {
		t.Error("expected commit-msg to start with bash shebang")
	}
}
