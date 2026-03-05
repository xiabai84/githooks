package config

import (
	"path/filepath"
	"testing"
)

func TestNewPaths_BuildsCorrectPaths(t *testing.T) {
	p := NewPaths("/home/testuser")

	if p.HomeDir != "/home/testuser" {
		t.Errorf("HomeDir = %q, want /home/testuser", p.HomeDir)
	}
	if p.HookDir != filepath.Join("/home/testuser", ".githooks") {
		t.Errorf("HookDir = %q, want %q", p.HookDir, filepath.Join("/home/testuser", ".githooks"))
	}
	if p.HookConfigDir != filepath.Join("/home/testuser", ".githooks", "config") {
		t.Errorf("HookConfigDir = %q, want %q", p.HookConfigDir, filepath.Join("/home/testuser", ".githooks", "config"))
	}
	if p.GithooksLogPath != filepath.Join("/home/testuser", ".githooks", "config", "githooks.log") {
		t.Errorf("GithooksLogPath = %q, want %q", p.GithooksLogPath, filepath.Join("/home/testuser", ".githooks", "config", "githooks.log"))
	}
	if p.GithooksConfigPath != filepath.Join("/home/testuser", ".githooks", "config", "githooks.json") {
		t.Errorf("GithooksConfigPath = %q, want %q", p.GithooksConfigPath, filepath.Join("/home/testuser", ".githooks", "config", "githooks.json"))
	}
	if p.CommitMsgPath != filepath.Join("/home/testuser", ".githooks", "commit-msg") {
		t.Errorf("CommitMsgPath = %q, want %q", p.CommitMsgPath, filepath.Join("/home/testuser", ".githooks", "commit-msg"))
	}
	if p.GitConfigPath != filepath.Join("/home/testuser", ".gitconfig") {
		t.Errorf("GitConfigPath = %q, want %q", p.GitConfigPath, filepath.Join("/home/testuser", ".gitconfig"))
	}
}

func TestNewPaths_DifferentHomeDirs(t *testing.T) {
	tests := []struct {
		homeDir         string
		wantHookDir     string
		wantGitConfig   string
	}{
		{"/home/alice", filepath.Join("/home/alice", ".githooks"), filepath.Join("/home/alice", ".gitconfig")},
		{"/Users/bob", filepath.Join("/Users/bob", ".githooks"), filepath.Join("/Users/bob", ".gitconfig")},
		{"/tmp/test", filepath.Join("/tmp/test", ".githooks"), filepath.Join("/tmp/test", ".gitconfig")},
	}

	for _, tt := range tests {
		t.Run(tt.homeDir, func(t *testing.T) {
			p := NewPaths(tt.homeDir)
			if p.HookDir != tt.wantHookDir {
				t.Errorf("HookDir = %q, want %q", p.HookDir, tt.wantHookDir)
			}
			if p.GitConfigPath != tt.wantGitConfig {
				t.Errorf("GitConfigPath = %q, want %q", p.GitConfigPath, tt.wantGitConfig)
			}
		})
	}
}

func TestNewPaths_NestedStructure(t *testing.T) {
	p := NewPaths("/home/user")

	// HookConfigDir should be inside HookDir
	if filepath.Dir(p.HookConfigDir) != p.HookDir {
		t.Errorf("HookConfigDir parent = %q, want %q", filepath.Dir(p.HookConfigDir), p.HookDir)
	}

	// CommitMsgPath should be inside HookDir
	if filepath.Dir(p.CommitMsgPath) != p.HookDir {
		t.Errorf("CommitMsgPath parent = %q, want %q", filepath.Dir(p.CommitMsgPath), p.HookDir)
	}

	// GithooksConfigPath should be inside HookConfigDir
	if filepath.Dir(p.GithooksConfigPath) != p.HookConfigDir {
		t.Errorf("GithooksConfigPath parent = %q, want %q", filepath.Dir(p.GithooksConfigPath), p.HookConfigDir)
	}

	// GithooksLogPath should be inside HookConfigDir
	if filepath.Dir(p.GithooksLogPath) != p.HookConfigDir {
		t.Errorf("GithooksLogPath parent = %q, want %q", filepath.Dir(p.GithooksLogPath), p.HookConfigDir)
	}
}

func TestDefaultPaths_Initialized(t *testing.T) {
	if Default.HomeDir == "" {
		t.Error("Default.HomeDir should not be empty")
	}
	if Default.HookDir == "" {
		t.Error("Default.HookDir should not be empty")
	}
	if Default.GitConfigPath == "" {
		t.Error("Default.GitConfigPath should not be empty")
	}
}
