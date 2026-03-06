package hooks

import (
	"os"
	"strings"
	"testing"

	"github.com/xiabai84/githooks/config"
)

func TestUninstall_RemovesHookDir(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create hook files
	if err := os.WriteFile(config.Default.CommitMsgPath, []byte("#!/bin/bash"), 0755); err != nil {
		t.Fatalf("failed to write commit-msg: %v", err)
	}
	if err := os.WriteFile(config.Default.GithooksConfigPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create .gitconfig without includeIf blocks
	if err := os.WriteFile(config.Default.GitConfigPath, []byte("[user]\n    name = Test\n"), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	if _, err := os.Stat(config.Default.HookDir); !os.IsNotExist(err) {
		t.Error("expected ~/.githooks directory to be removed")
	}
}

func TestUninstall_CleansGitConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	gitConfigContent := `[user]
    name = Test
[includeIf "gitdir:~/work/alpha/"]
    path = .githooks/config/gitconfig-alpha
[includeIf "gitdir:~/work/beta/"]
    path = .githooks/config/gitconfig-beta
[core]
    editor = vim
`
	if err := os.WriteFile(config.Default.GitConfigPath, []byte(gitConfigContent), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}

	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}

	data, err := os.ReadFile(config.Default.GitConfigPath)
	if err != nil {
		t.Fatalf("failed to read git config: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "includeIf") {
		t.Error("expected includeIf blocks to be removed from .gitconfig")
	}
	if !strings.Contains(content, "[user]") {
		t.Error("expected non-githooks config to be preserved")
	}
	if !strings.Contains(content, "[core]") {
		t.Error("expected non-githooks config to be preserved")
	}
}

func TestUninstall_NoGitConfig(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// No .gitconfig exists — should not error
	if err := Uninstall(); err != nil {
		t.Fatalf("Uninstall returned error: %v", err)
	}
}
