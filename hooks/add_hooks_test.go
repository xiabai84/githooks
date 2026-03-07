package hooks

import (
	"os"
	"testing"

	"github.com/xiabai84/githooks/config"
	"github.com/xiabai84/githooks/types"
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

func TestMergeJiraKeys_SingleKeys(t *testing.T) {
	result := mergeJiraKeys("ALPHA", "BETA")
	if result != "(ALPHA|BETA)" {
		t.Errorf("expected (ALPHA|BETA), got %s", result)
	}
}

func TestMergeJiraKeys_ExistingGroup(t *testing.T) {
	result := mergeJiraKeys("(ALPHA|BETA)", "GAMMA")
	if result != "(ALPHA|BETA|GAMMA)" {
		t.Errorf("expected (ALPHA|BETA|GAMMA), got %s", result)
	}
}

func TestMergeJiraKeys_BothGroups(t *testing.T) {
	result := mergeJiraKeys("(ALPHA|BETA)", "(GAMMA|DELTA)")
	if result != "(ALPHA|BETA|GAMMA|DELTA)" {
		t.Errorf("expected (ALPHA|BETA|GAMMA|DELTA), got %s", result)
	}
}

func TestMergeJiraKeys_Duplicate(t *testing.T) {
	result := mergeJiraKeys("(ALPHA|BETA)", "ALPHA")
	if result != "(ALPHA|BETA)" {
		t.Errorf("expected (ALPHA|BETA), got %s", result)
	}
}

func TestMergeJiraKeys_SingleSame(t *testing.T) {
	result := mergeJiraKeys("ALPHA", "ALPHA")
	if result != "ALPHA" {
		t.Errorf("expected ALPHA, got %s", result)
	}
}

func TestValidateJiraKeyRegex_Valid(t *testing.T) {
	cases := []string{"PROJ", "(ALPHA|BETA)", "(A|B|C)", "MY_PROJECT"}
	for _, key := range cases {
		if err := ValidateJiraKeyRegex(key); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", key, err)
		}
	}
}

func TestValidateJiraKeyRegex_Invalid(t *testing.T) {
	cases := []string{"(ALPHA|", "(?invalid)", "[bad"}
	for _, key := range cases {
		if err := ValidateJiraKeyRegex(key); err == nil {
			t.Errorf("expected %q to be invalid, got no error", key)
		}
	}
}

func TestAddWorkspace_RejectsInvalidRegex(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws := &types.Workspace{Name: "Bad", ProjectKeyRE: "(ALPHA|", Folder: "~/bad/"}
	err := AddWorkspace(ws)
	if err == nil {
		t.Error("expected error for invalid regex, got nil")
	}
}

func TestAddWorkspace_RejectsDuplicateName(t *testing.T) {
	cleanup := setupTestConfig(t)
	defer cleanup()

	if err := os.WriteFile(config.Default.GitConfigPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write git config: %v", err)
	}
	initialConfig := &types.GitHookConfig{Version: "1.0.0", Workspaces: []types.Workspace{}}
	if err := WriteGitHooksConfig(initialConfig); err != nil {
		t.Fatalf("failed to write githooks config: %v", err)
	}

	ws1 := &types.Workspace{Name: "MyProject", ProjectKeyRE: "ALPHA", Folder: "~/alpha/"}
	if err := AddWorkspace(ws1); err != nil {
		t.Fatalf("AddWorkspace(ws1) returned error: %v", err)
	}

	// Different folder, same name — should fail
	ws2 := &types.Workspace{Name: "MyProject", ProjectKeyRE: "BETA", Folder: "~/beta/"}
	err := AddWorkspace(ws2)
	if err == nil {
		t.Error("expected error for duplicate workspace name, got nil")
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
