package hooks

import (
	"os"
	"testing"

	"github.com/stefan-niemeyer/githooks/types"
)

func TestGetWorkspaceIndex_MatchesCwd(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd returned error: %v", err)
	}

	workspaces := []types.Workspace{
		{Name: "no-match", Folder: "/nonexistent/path/"},
		{Name: "match", Folder: cwd + "/"},
	}

	idx, err := GetWorkspaceIndex(workspaces)
	if err != nil {
		t.Fatalf("GetWorkspaceIndex returned error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1, got %d", idx)
	}
}

func TestGetWorkspaceIndex_NoMatch(t *testing.T) {
	workspaces := []types.Workspace{
		{Name: "no-match", Folder: "/nonexistent/path/"},
	}

	idx, err := GetWorkspaceIndex(workspaces)
	if err != nil {
		t.Fatalf("GetWorkspaceIndex returned error: %v", err)
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}

func TestGetWorkspaceIndex_LongestMatch(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd returned error: %v", err)
	}

	workspaces := []types.Workspace{
		{Name: "short", Folder: home + "/"},
		{Name: "exact", Folder: cwd + "/"},
	}

	idx, err := GetWorkspaceIndex(workspaces)
	if err != nil {
		t.Fatalf("GetWorkspaceIndex returned error: %v", err)
	}
	if idx != 1 {
		t.Errorf("expected index 1 (longest match), got %d", idx)
	}
}

func TestGetWorkspaceIndex_EmptyList(t *testing.T) {
	idx, err := GetWorkspaceIndex([]types.Workspace{})
	if err != nil {
		t.Fatalf("GetWorkspaceIndex returned error: %v", err)
	}
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}
}
