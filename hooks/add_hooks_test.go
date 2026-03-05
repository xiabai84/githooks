package hooks

import (
	"os"
	"testing"

	"github.com/xiabai84/githooks/config"
)

func TestCheckConfigFiles_AllExist(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	if err := os.WriteFile(config.Default.GitConfigPath, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	if err := os.WriteFile(config.Default.CommitMsgPath, []byte{}, 0755); err != nil {
		t.Fatalf("failed to write commit-msg: %v", err)
	}
	if err := os.WriteFile(config.Default.GithooksConfigPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	err := CheckConfigFiles()
	if err != nil {
		t.Errorf("expected no error when all files exist, got: %v", err)
	}
}

func TestCheckConfigFiles_MissingFile(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	err := CheckConfigFiles()
	if err == nil {
		t.Error("expected error when config file is missing")
	}
}
