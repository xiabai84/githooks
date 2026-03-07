package hooks

import (
	"strings"
	"testing"
)

func TestCheckCommitMessage_ValidConventionalCommit(t *testing.T) {
	cases := []struct {
		msg      string
		projects string
	}{
		{"feat(PROJ-123): add login", "PROJ"},
		{"fix(PROJ-456): fix crash", "PROJ"},
		{"docs(PROJ-1): update readme", "PROJ"},
		{"chore(PROJ-99): bump deps", "PROJ"},
		{"feat(MOB-1)!: breaking change", "MOB"},
		{"revert(ABC-42): revert bad commit", "ABC"},
		{"feat(PROJ-123): add login", ""},           // no project filter
		{"feat(PROJ-123): add login", "(PROJ|MOB)"}, // multi-project
	}
	for _, tc := range cases {
		result := CheckCommitMessage(tc.msg, tc.projects, "")
		if !result.Valid {
			t.Errorf("expected %q with projects=%q to be valid, got error: %s", tc.msg, tc.projects, result.Error)
		}
	}
}

func TestCheckCommitMessage_InvalidFormat(t *testing.T) {
	result := CheckCommitMessage("not a conventional commit", "", "")
	if result.Valid {
		t.Error("expected non-conventional message to be invalid")
	}
	if result.Error == "" {
		t.Error("expected error message for invalid format")
	}
}

func TestCheckCommitMessage_MissingJiraTicket(t *testing.T) {
	result := CheckCommitMessage("feat: add login without ticket", "PROJ", "")
	if result.Valid {
		t.Error("expected message without Jira ticket to be invalid")
	}
}

func TestCheckCommitMessage_MergeCommitAllowed(t *testing.T) {
	result := CheckCommitMessage("Merge branch 'feature' into main", "PROJ", "")
	if !result.Valid {
		t.Errorf("expected merge commit to be valid, got error: %s", result.Error)
	}
}

func TestCheckCommitMessage_AutoInjectFromBranch(t *testing.T) {
	result := CheckCommitMessage("feat: add login", "PROJ", "feat/PROJ-123-login")
	if !result.Valid {
		t.Errorf("expected auto-inject to make message valid, got error: %s", result.Error)
	}
	if result.Message != "feat(PROJ-123): add login" {
		t.Errorf("expected message to be auto-injected, got %q", result.Message)
	}
}

func TestCheckCommitMessage_WrongProjectTicket(t *testing.T) {
	result := CheckCommitMessage("feat(OTHER-123): add login", "PROJ", "")
	if result.Valid {
		t.Error("expected message with wrong project ticket to be invalid")
	}
}

func TestCheckCommitMessage_NoProjectFilter(t *testing.T) {
	// With no project filter, any ticket format should work
	result := CheckCommitMessage("feat(ANY-999): something", "", "")
	if !result.Valid {
		t.Errorf("expected any ticket to be valid when no project filter, got error: %s", result.Error)
	}
}

func TestCheckBranchName_ValidBranch(t *testing.T) {
	cases := []struct {
		branch   string
		projects string
	}{
		{"feat/PROJ-123-add-login", "PROJ"},
		{"fix/PROJ-456-fix-crash", "PROJ"},
		{"hotfix/PROJ-1-urgent", "PROJ"},
		{"chore/PROJ-99-bump-deps", "PROJ"},
		{"bugfix/PROJ-42-edge-case", "PROJ"},
		{"docs/PROJ-10-update-readme", "PROJ"},
		{"refactor/PROJ-5-cleanup", "PROJ"},
		{"test/PROJ-7-add-tests", "PROJ"},
		{"ci/PROJ-3-pipeline", "PROJ"},
		{"release/PROJ-200-prep", "PROJ"},
		{"feature/PROJ-50-long-form", "PROJ"},
		{"feat/MOB-1-something", "(PROJ|MOB)"},
		{"feat/ANY-999-something", ""},
		{"release/2.0.0", "PROJ"},           // release branches exempt from ticket
		{"release/v1.5.0-rc1", "PROJ"},      // release with version tag
	}
	for _, tc := range cases {
		result := CheckBranchName(tc.branch, tc.projects)
		if !result.Valid {
			t.Errorf("expected branch %q with projects=%q to be valid, got error: %s", tc.branch, tc.projects, result.Error)
		}
	}
}

func TestCheckBranchName_ExemptBranches(t *testing.T) {
	for _, branch := range []string{"main", "master", "develop"} {
		result := CheckBranchName(branch, "PROJ")
		if !result.Valid {
			t.Errorf("expected exempt branch %q to be valid, got error: %s", branch, result.Error)
		}
	}
}

func TestCheckBranchName_InvalidPrefix(t *testing.T) {
	result := CheckBranchName("add-login-page", "PROJ")
	if result.Valid {
		t.Error("expected branch without valid prefix to be invalid")
	}
	if !strings.Contains(result.Error, "must follow convention") {
		t.Errorf("expected error about convention, got: %s", result.Error)
	}
}

func TestCheckBranchName_MissingTicket(t *testing.T) {
	result := CheckBranchName("feat/add-login", "PROJ")
	if result.Valid {
		t.Error("expected branch without ticket to be invalid")
	}
	if !strings.Contains(result.Error, "Jira ticket") {
		t.Errorf("expected error about Jira ticket, got: %s", result.Error)
	}
}

func TestCheckBranchName_WrongProjectTicket(t *testing.T) {
	result := CheckBranchName("feat/OTHER-123-add-login", "PROJ")
	if result.Valid {
		t.Error("expected branch with wrong project ticket to be invalid")
	}
}

func TestCheckBranchName_NoProjectFilter(t *testing.T) {
	result := CheckBranchName("feat/ANYTHING-42-desc", "")
	if !result.Valid {
		t.Errorf("expected any ticket to be valid with no project filter, got error: %s", result.Error)
	}
}
