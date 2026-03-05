package hooks

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/xiabai84/githooks/config"
)

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
