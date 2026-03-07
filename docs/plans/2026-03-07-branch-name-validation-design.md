# Branch Name Convention Validation

**Date:** 2026-03-07
**Status:** Approved

## Summary

Add branch name convention validation to githooks, enforced both at commit time (bash `commit-msg` hook) and via `githooks check --branch-name`.

## Branch Name Pattern

Valid branch names must match:

```
<prefix>/<TICKET>-<description>
```

- **prefix**: `feat|fix|hotfix|chore|release|bugfix|docs|refactor|test|ci`
- **TICKET**: Jira ticket matching the workspace's `ProjectKeyRE` (e.g. `PROJ-123`)
- **description**: kebab-case (lowercase, hyphens)

**Exempt branches:** `main`, `master`, `develop` skip validation. Detached HEAD is also exempt.

## Approach

Approach 3: Pure bash in hook + Go check command, matching the existing project pattern where commit-msg validation has parallel bash and Go implementations.

## Components

### 1. Bash hook (`config/commit_msg.go`)

Add branch name validation before the existing commit message validation:
- Check if branch is exempt (`main`, `master`, `develop`)
- Validate prefix matches allowed types
- Extract Jira ticket from branch, validate against `$PROJECTS`
- On failure: print error with example, exit 1

### 2. Go validation (`hooks/check_hooks.go`)

New `CheckBranchName(branch, projects string) CheckResult`:
- Same logic as bash: exempt → prefix → ticket
- Reuses existing `extractTicket()` helper

### 3. CLI integration (`cmd/check.go`)

New flag `--branch-name`:
- With argument: `githooks check --branch-name feat/PROJ-123-login`
- Without commit message argument: reads current branch via `git symbolic-ref`

### 4. Tests

Go tests for `CheckBranchName`:
- Valid branch with correct prefix and ticket
- Invalid prefix
- Missing ticket
- Wrong project ticket
- Exempt branches (`main`, `master`, `develop`)
- No project filter (any ticket accepted)

## Error Output

```
ERROR: Branch name must follow convention: <type>/<TICKET>-<description>
  Allowed types: feat, fix, hotfix, chore, release, bugfix, docs, refactor, test, ci
  Example: feat/PROJ-123-add-user-auth

  Current branch: add-login-page
```
