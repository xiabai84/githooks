# Branch Name Validation Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add branch name convention validation enforced at commit time (bash hook) and via `githooks check --branch-name`.

**Architecture:** Parallel bash + Go implementations matching the existing commit message validation pattern. Go `CheckBranchName()` reuses the existing `extractTicket()` helper. Bash validation is prepended to the existing `commit-msg` hook. CLI gets a `--branch-name` flag on the existing `check` command.

**Tech Stack:** Go (testing, regexp, strings), Bash (commit-msg hook), Cobra CLI

---

### Task 1: Go — CheckBranchName tests

**Files:**
- Modify: `hooks/check_hooks_test.go`

**Step 1: Write failing tests for CheckBranchName**

Add these tests to `hooks/check_hooks_test.go`:

```go
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
		{"feat/MOB-1-something", "(PROJ|MOB)"},
		{"feat/ANY-999-something", ""},
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
```

**Step 2: Add `strings` to the import block if not already present**

The test file import block should include `"strings"` (for `strings.Contains`). Check before adding.

**Step 3: Run tests to verify they fail**

Run: `go test ./hooks/ -run TestCheckBranchName -count=1 -v`
Expected: FAIL — `CheckBranchName` is undefined.

---

### Task 2: Go — Implement CheckBranchName

**Files:**
- Modify: `hooks/check_hooks.go`

**Step 1: Add branch name types and regex**

Add these package-level vars after the existing `convRE` line in `hooks/check_hooks.go`:

```go
var branchPrefixes = "feat|fix|hotfix|chore|release|bugfix|docs|refactor|test|ci"
var branchRE = regexp.MustCompile(`^(` + branchPrefixes + `)/(.+)`)
var exemptBranches = map[string]bool{"main": true, "master": true, "develop": true}
```

**Step 2: Implement CheckBranchName function**

Add after `CheckCommitMessage`:

```go
// CheckBranchName validates a branch name against naming conventions.
// projects is the Jira project key filter (e.g. "PROJ" or "(PROJ|MOB)"), empty means any.
func CheckBranchName(branch, projects string) CheckResult {
	if exemptBranches[branch] {
		return CheckResult{Valid: true, Message: branch}
	}

	if !branchRE.MatchString(branch) {
		return CheckResult{
			Valid: false,
			Error: fmt.Sprintf("Branch name must follow convention: <type>/<TICKET>-<description>\n"+
				"  Allowed types: %s\n"+
				"  Example: feat/PROJ-123-add-user-auth\n\n"+
				"  Current branch: %s", strings.Replace(branchPrefixes, "|", ", ", -1), branch),
		}
	}

	ticket := extractTicket(branch, projects)
	if ticket == "" {
		errMsg := "Branch name must include a Jira ticket."
		if projects != "" {
			errMsg = fmt.Sprintf("Branch name must include a Jira ticket matching '%s'.", projects)
		}
		return CheckResult{
			Valid: false,
			Error: fmt.Sprintf("%s\n  Example: feat/%s-123-add-feature\n\n  Current branch: %s", errMsg, projects, branch),
		}
	}

	return CheckResult{Valid: true, Message: branch}
}
```

**Step 3: Run tests to verify they pass**

Run: `go test ./hooks/ -run TestCheckBranchName -count=1 -v`
Expected: All 6 tests PASS.

**Step 4: Run full test suite**

Run: `go test ./... -count=1`
Expected: All packages PASS.

**Step 5: Commit**

```bash
git add hooks/check_hooks.go hooks/check_hooks_test.go
git commit -m "feat: add CheckBranchName validation with tests"
```

---

### Task 3: CLI — Add --branch-name flag to check command

**Files:**
- Modify: `cmd/check.go`

**Step 1: Update check command to support --branch-name**

Replace the entire `cmd/check.go` with:

```go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xiabai84/githooks/hooks"
)

var checkBranch string
var checkBranchName string

var checkCmd = &cobra.Command{
	Use:   "check [message]",
	Short: "Validate a commit message or branch name",
	Long: `Validates a commit message against Conventional Commits format and Jira ticket rules,
or validates a branch name against naming conventions. Useful for CI pipelines and debugging.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projects := getProjects()

		// Branch name validation mode
		if checkBranchName != "" || cmd.Flags().Changed("branch-name") {
			branch := checkBranchName
			if branch == "" {
				// Read current branch from git
				out, err := exec.Command("git", "symbolic-ref", "--short", "HEAD").Output()
				if err != nil {
					fmt.Fprintln(os.Stderr, "✘ Could not determine current branch (detached HEAD?)")
					os.Exit(1)
				}
				branch = strings.TrimSpace(string(out))
			}
			result := hooks.CheckBranchName(branch, projects)
			if result.Valid {
				fmt.Println("✔ Valid branch:", result.Message)
			} else {
				fmt.Fprintln(os.Stderr, "✘ Invalid:", result.Error)
				os.Exit(1)
			}
			return
		}

		// Commit message validation mode (requires argument)
		if len(args) == 0 {
			fmt.Fprintln(os.Stderr, "Error: requires a commit message argument or --branch-name flag")
			os.Exit(1)
		}
		msg := args[0]
		result := hooks.CheckCommitMessage(msg, projects, checkBranch)
		if result.Valid {
			fmt.Println("✔ Valid:", result.Message)
		} else {
			fmt.Fprintln(os.Stderr, "✘ Invalid:", result.Error)
			os.Exit(1)
		}
	},
}

func getProjects() string {
	ghConfig, err := hooks.ReadGitHooksConfig()
	if err != nil || len(ghConfig.Workspaces) == 0 {
		return ""
	}
	idx, _ := hooks.GetWorkspaceIndex(ghConfig.Workspaces)
	return ghConfig.Workspaces[idx].ProjectKeyRE
}

func init() {
	checkCmd.Flags().StringVar(&checkBranch, "branch", "", "Simulate branch name for commit message auto-injection (e.g. feat/PROJ-123-login)")
	checkCmd.Flags().StringVar(&checkBranchName, "branch-name", "", "Validate a branch name (omit value to check current branch)")
	checkCmd.Flag("branch-name").NoOptDefVal = ""
	rootCmd.AddCommand(checkCmd)
}
```

**Step 2: Build to verify no compilation errors**

Run: `go build ./...`
Expected: Success, no errors.

**Step 3: Run full test suite**

Run: `go test ./... -count=1`
Expected: All packages PASS.

**Step 4: Commit**

```bash
git add cmd/check.go
git commit -m "feat: add --branch-name flag to check command"
```

---

### Task 4: Bash — Add branch validation to commit-msg hook

**Files:**
- Modify: `config/commit_msg.go`

**Step 1: Add branch name validation to the bash hook**

In `config/commit_msg.go`, insert the following bash block **after** the `PROJECTS` resolution (after `fi` on line 21) and **before** `FIRST_LINE=$(head -n 1 "$1")` (line 23). This new block goes between lines 21 and 23:

```bash

# Branch name convention validation
BRANCH_TYPES="feat|fix|hotfix|chore|release|bugfix|docs|refactor|test|ci"
CURRENT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null || echo "")

if [[ -n "$CURRENT_BRANCH" && "$CURRENT_BRANCH" != "main" && "$CURRENT_BRANCH" != "master" && "$CURRENT_BRANCH" != "develop" ]]; then
  BRANCH_RE="^(${BRANCH_TYPES})/.+"
  if ! [[ "$CURRENT_BRANCH" =~ $BRANCH_RE ]]; then
    echo >&2 "ERROR: Branch name must follow convention: <type>/<TICKET>-<description>"
    echo >&2 "  Allowed types: feat, fix, hotfix, chore, release, bugfix, docs, refactor, test, ci"
    echo >&2 "  Example: feat/PROJ-123-add-user-auth"
    echo >&2 ""
    echo >&2 "  Current branch: $CURRENT_BRANCH"
    exit 1
  fi

  # Validate Jira ticket in branch name
  if [ -n "$PROJECTS" ]; then
    BRANCH_TICKET=$(echo "$CURRENT_BRANCH" | grep --ignore-case --extended-regexp --only-matching --regexp="\<${PROJECTS}-[[:digit:]]+\>" | tr '[:lower:]' '[:upper:]')
  else
    BRANCH_TICKET=$(echo "$CURRENT_BRANCH" | grep --extended-regexp --only-matching --regexp='\<[[:alpha:]][[:alnum:]]*-[[:digit:]]+\>' | tr '[:lower:]' '[:upper:]')
  fi

  if [[ -z "$BRANCH_TICKET" ]]; then
    if [ -n "$PROJECTS" ]; then
      echo >&2 "ERROR: Branch name must include a Jira ticket matching '$PROJECTS'."
    else
      echo >&2 "ERROR: Branch name must include a Jira ticket."
    fi
    echo >&2 "  Example: feat/PROJ-123-add-user-auth"
    echo >&2 ""
    echo >&2 "  Current branch: $CURRENT_BRANCH"
    exit 1
  fi
fi

```

**Step 2: Run full test suite**

Run: `go test ./... -count=1`
Expected: All packages PASS. (The bash content is a Go string constant — tests verify the Go-side logic, the bash is tested implicitly at commit time.)

**Step 3: Commit**

```bash
git add config/commit_msg.go
git commit -m "feat: add branch name validation to commit-msg hook"
```

---

### Task 5: Update README documentation

**Files:**
- Modify: `README.md`

**Step 1: Add branch name validation docs**

Find the existing `githooks check` section in README.md (around the "Validating a Commit Message" heading) and add branch name validation examples after it:

```markdown
### Validating a Branch Name

```bash
# Validate a specific branch name
githooks check --branch-name feat/PROJ-123-add-login

# Validate the current branch
githooks check --branch-name
```

Branch names must follow the convention `<type>/<TICKET>-<description>`:
- **Allowed types:** `feat`, `fix`, `hotfix`, `chore`, `release`, `bugfix`, `docs`, `refactor`, `test`, `ci`
- **Ticket:** Must match the workspace's Jira project key
- **Exempt branches:** `main`, `master`, `develop` skip validation

Branch names are also validated automatically at commit time via the `commit-msg` hook.
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add branch name validation to README"
```

---

### Task 6: Final verification

**Step 1: Run full test suite**

Run: `go test ./... -count=1 -race`
Expected: All packages PASS.

**Step 2: Build binary and smoke test**

Run:
```bash
go run . check --branch-name feat/TEST-123-something
go run . check --branch-name main
go run . check --branch-name bad-branch-name
```

Expected:
- First: `✔ Valid branch: feat/TEST-123-something`
- Second: `✔ Valid branch: main`
- Third: `✘ Invalid: Branch name must follow convention...` (exit 1)

**Step 3: Push**

```bash
git push
```
