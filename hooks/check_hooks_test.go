package hooks

import (
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
