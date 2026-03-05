package cmd

import (
	"testing"

	"github.com/stefan-niemeyer/githooks/config"
	"github.com/stefan-niemeyer/githooks/types"
)

func TestWorkspaceSelectTemplates_ReturnsNonNil(t *testing.T) {
	tmpl := workspaceSelectTemplates()
	if tmpl == nil {
		t.Fatal("expected non-nil templates")
	}
	if tmpl.Active == "" {
		t.Error("expected non-empty Active template")
	}
	if tmpl.Inactive == "" {
		t.Error("expected non-empty Inactive template")
	}
	if tmpl.Selected == "" {
		t.Error("expected non-empty Selected template")
	}
	if tmpl.Details != config.DetailTmpl {
		t.Error("expected Details to match config.DetailTmpl")
	}
}

func TestWorkspaceSearcher_FindsByName(t *testing.T) {
	workspaces := []types.Workspace{
		{Name: "Alpha Project"},
		{Name: "Beta Service"},
		{Name: "Gamma API"},
	}
	searcher := workspaceSearcher(workspaces)

	tests := []struct {
		input string
		index int
		want  bool
	}{
		{"alpha", 0, true},
		{"ALPHA", 0, true},
		{"Alpha", 0, true},
		{"beta", 1, true},
		{"gamma", 2, true},
		{"api", 2, true},
		{"delta", 0, false},
		{"delta", 1, false},
		{"", 0, true},   // empty input matches everything
	}

	for _, tt := range tests {
		t.Run(tt.input+"_idx"+string(rune('0'+tt.index)), func(t *testing.T) {
			got := searcher(tt.input, tt.index)
			if got != tt.want {
				t.Errorf("searcher(%q, %d) = %v, want %v", tt.input, tt.index, got, tt.want)
			}
		})
	}
}

func TestWorkspaceSearcher_IgnoresSpaces(t *testing.T) {
	workspaces := []types.Workspace{
		{Name: "My Project"},
	}
	searcher := workspaceSearcher(workspaces)

	if !searcher("myproject", 0) {
		t.Error("expected searcher to match 'myproject' against 'My Project' (ignoring spaces)")
	}
	if !searcher("my project", 0) {
		t.Error("expected searcher to match 'my project' against 'My Project'")
	}
}
