package hooks

import (
	"os"
	"testing"

	"github.com/stefan-niemeyer/githooks/config"
)

func TestMigrateGitHooksConfig_NoLogFile(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// No legacy log file exists — migration should be a no-op
	err := MigrateGitHooksConfig()
	if err != nil {
		t.Fatalf("MigrateGitHooksConfig returned error: %v", err)
	}
}

func TestMigrateGitHooksConfig_MigratesLegacyLog(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	// Create a legacy log file with two entries
	legacyContent := `{"Project":"alpha","JiraName":"ALPHA","WorkDir":"~/alpha/"}
{"Project":"beta","JiraName":"BETA","WorkDir":"~/beta/"}
`
	if err := os.WriteFile(config.Default.GithooksLogPath, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("failed to write legacy log: %v", err)
	}

	err := MigrateGitHooksConfig()
	if err != nil {
		t.Fatalf("MigrateGitHooksConfig returned error: %v", err)
	}

	// Legacy log should be deleted
	if _, err := os.Stat(config.Default.GithooksLogPath); !os.IsNotExist(err) {
		t.Error("expected legacy log file to be deleted after migration")
	}

	// New config should contain migrated workspaces
	ghConfig, err := ReadGitHooksConfig()
	if err != nil {
		t.Fatalf("ReadGitHooksConfig returned error: %v", err)
	}
	if len(ghConfig.Workspaces) != 2 {
		t.Fatalf("expected 2 migrated workspaces, got %d", len(ghConfig.Workspaces))
	}
	if ghConfig.Workspaces[0].Name != "alpha" {
		t.Errorf("expected first workspace name 'alpha', got %q", ghConfig.Workspaces[0].Name)
	}
	if ghConfig.Workspaces[1].ProjectKeyRE != "BETA" {
		t.Errorf("expected second workspace key 'BETA', got %q", ghConfig.Workspaces[1].ProjectKeyRE)
	}
}

func TestMigrateGitHooksConfig_InvalidJSON(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	if err := os.WriteFile(config.Default.GithooksLogPath, []byte("not valid json\n"), 0644); err != nil {
		t.Fatalf("failed to write legacy log: %v", err)
	}

	err := MigrateGitHooksConfig()
	if err == nil {
		t.Error("expected error for invalid JSON in legacy log")
	}
}
