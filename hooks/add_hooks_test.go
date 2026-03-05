package hooks

import (
	"os"
	"testing"

	"github.com/stefan-niemeyer/githooks/config"
)

func TestCheckConfigFiles_AllExist(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	os.WriteFile(config.Default.GitConfigPath, []byte{}, 0644)
	os.WriteFile(config.Default.CommitMsgPath, []byte{}, 0755)
	os.WriteFile(config.Default.GithooksConfigPath, []byte("{}"), 0644)

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
