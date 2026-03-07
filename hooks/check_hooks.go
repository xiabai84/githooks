package hooks

import (
	"fmt"
	"regexp"
	"strings"
)

// CheckResult holds the result of a commit message validation.
type CheckResult struct {
	Valid   bool
	Message string // the (possibly auto-injected) commit message
	Error   string
}

var commitTypes = "feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
var convRE = regexp.MustCompile(`^(` + commitTypes + `)(\(([^)]*)\))?(!)?: (.+)`)

// CheckCommitMessage validates a commit message against Conventional Commits + Jira rules.
// projects is the Jira project key filter (e.g. "PROJ" or "(PROJ|MOB)"), empty means any.
// branch is the git branch name for auto-injection simulation, empty to skip.
func CheckCommitMessage(msg, projects, branch string) CheckResult {
	firstLine := strings.SplitN(msg, "\n", 2)[0]

	// Allow merge commits
	if strings.HasPrefix(firstLine, "Merge ") {
		return CheckResult{Valid: true, Message: firstLine}
	}

	// Validate Conventional Commits format
	matches := convRE.FindStringSubmatch(firstLine)
	if matches == nil {
		return CheckResult{
			Valid: false,
			Error: "Commit message must follow Conventional Commits format: <type>(<scope>): <description>",
		}
	}

	commitType := matches[1]
	bang := matches[4]
	desc := matches[5]

	// Extract Jira ticket from message
	msgTicket := extractTicket(firstLine, projects)

	// Try auto-inject from branch
	if branch != "" {
		branchTicket := extractTicket(branch, projects)
		if branchTicket != "" && msgTicket != branchTicket {
			// Auto-inject branch ticket as scope
			newMsg := fmt.Sprintf("%s(%s)%s: %s", commitType, branchTicket, bang, desc)
			return CheckResult{Valid: true, Message: newMsg}
		}
	}

	// Ensure Jira ticket present
	if msgTicket == "" {
		errMsg := "Commit message must include a Jira ticket."
		if projects != "" {
			errMsg = fmt.Sprintf("Commit message must include a Jira ticket matching '%s'.", projects)
		}
		return CheckResult{Valid: false, Error: errMsg}
	}

	return CheckResult{Valid: true, Message: firstLine}
}

// extractTicket finds the first Jira ticket in text matching the project filter.
func extractTicket(text, projects string) string {
	text = strings.ToUpper(text)
	var re *regexp.Regexp
	if projects != "" {
		// Strip outer parens for regex building
		p := projects
		if strings.HasPrefix(p, "(") && strings.HasSuffix(p, ")") {
			p = p[1 : len(p)-1]
		}
		re = regexp.MustCompile(`\b(` + p + `)-\d+\b`)
	} else {
		re = regexp.MustCompile(`\b[A-Z][A-Z0-9]*-\d+\b`)
	}
	match := re.FindString(text)
	return match
}
