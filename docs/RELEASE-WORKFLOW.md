# Release Workflow with githooks

This document describes how to use githooks in a team development workflow
with feature branches, pull requests, and automated releases from `main`.

## Overview

```
feature/ABC-123-login ──────┐
                             ├── PR & merge ──► main ──► tag v1.2.0 ──► release
feature/ABC-456-dashboard ──┘
```

1. Developers work on **feature branches** with Jira ticket keys
2. The commit-msg hook enforces Conventional Commits on every commit
3. Branches are merged to `main` via pull request
4. A version tag on `main` triggers the release pipeline

## Branch Naming Convention

Include the Jira ticket key in the branch name. The hook can then auto-insert
the ticket into commits that are missing it.

```
feature/ABC-123-add-login
bugfix/ABC-456-fix-null-pointer
hotfix/ABC-789-patch-auth
```

When working on `feature/ABC-123-add-login`, a commit like:

```bash
git commit -m "feat: add OAuth2 flow"
```

is automatically rewritten to:

```bash
feat(ABC-123): add OAuth2 flow
```

## Developer Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/ABC-123-add-login
```

### 2. Make Commits

Write commits following Conventional Commits. The hook validates every commit
and auto-inserts the Jira ticket from the branch name if missing.

```bash
# Ticket auto-inserted from branch name
git commit -m "feat: add login page"
# → feat(ABC-123): add login page

# Ticket already in message — accepted as-is
git commit -m "feat(ABC-123): add token refresh"

# Invalid — rejected by hook
git commit -m "added stuff"
```

Multiple commits on a branch are fine. Each commit should be a logical unit:

```bash
git commit -m "feat: add login form UI"
git commit -m "feat: add OAuth2 token flow"
git commit -m "test: add login integration tests"
git commit -m "docs: add login API documentation"
```

### 3. Push and Create a Pull Request

```bash
git push -u origin feature/ABC-123-add-login
gh pr create --title "feat(ABC-123): add OAuth2 login" --body "..."
```

### 4. Review and Merge

After approval, merge to `main`. Use **squash merge** if you want a single clean
commit on `main`, or **merge commit** to preserve the full history.

**Squash merge** (recommended for clean changelog):

```bash
gh pr merge --squash
# Creates: feat(ABC-123): add OAuth2 login (#42)
```

**Merge commit** (preserves all individual commits):

```bash
gh pr merge --merge
```

## Release Process

### Determine the Next Version

After merging one or more PRs to `main`, determine the version bump from the
commits since the last release:

```bash
# See commits since last tag
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

Apply the highest-priority bump:

| Priority | Trigger | Bump |
|---|---|---|
| 1 | Any commit with `!` or `BREAKING CHANGE:` | **Major** (`X.0.0`) |
| 2 | Any `feat` commit | **Minor** (`0.X.0`) |
| 3 | Only `fix`, `docs`, `refactor`, `perf`, `build`, `chore`, `revert` | **Patch** (`0.0.X`) |
| — | Only `style`, `test`, `ci` | No release needed |

Use the helper scripts to calculate automatically:

```bash
# From the latest commit
git log -1 --format=%s | python bump-version.py 1.2.3

# Or specify the message directly
python bump-version.py 1.2.3 'feat(ABC-123): add login'
# Output: 1.3.0
```

### Tag and Release

```bash
git checkout main
git pull

# Create an annotated tag
git tag v1.3.0 -m "v1.3.0 - Add OAuth2 login and dashboard improvements"
git push origin v1.3.0
```

The GitHub Actions release workflow automatically:
- Builds binaries for all platforms (Linux, macOS, Windows × amd64, arm64)
- Creates a GitHub release with changelog and downloadable assets
- Generates SHA256 checksums

### Verify

```bash
gh release view v1.3.0
```

## Example: Full Release Cycle

```bash
# --- Developer A: feature work ---
git checkout -b feature/ABC-123-add-login
git commit -m "feat: add login page"
git commit -m "feat: add token refresh"
git push -u origin feature/ABC-123-add-login
gh pr create --title "feat(ABC-123): add OAuth2 login"

# --- Developer B: bug fix ---
git checkout -b bugfix/ABC-456-fix-crash
git commit -m "fix: handle null user in session"
git push -u origin bugfix/ABC-456-fix-crash
gh pr create --title "fix(ABC-456): handle null user in session"

# --- After both PRs are merged to main ---
git checkout main && git pull

# Check what changed since last release (v1.2.0)
git log v1.2.0..HEAD --oneline
# abc1234 feat(ABC-123): add OAuth2 login (#42)
# def5678 fix(ABC-456): handle null user in session (#43)

# Highest priority: feat → minor bump
# v1.2.0 → v1.3.0
git tag v1.3.0 -m "v1.3.0 - Add OAuth2 login, fix session crash"
git push origin v1.3.0

# Release is built and published automatically
gh release view v1.3.0
```

## Hotfix Process

For critical fixes that need immediate release:

```bash
# Branch from main
git checkout main
git checkout -b hotfix/ABC-789-security-patch

# Fix and commit
git commit -m "fix(ABC-789): patch authentication bypass"

# Merge directly (or via fast PR)
git checkout main
git merge hotfix/ABC-789-security-patch

# Patch release
git tag v1.3.1 -m "v1.3.1 - Security patch for authentication bypass"
git push origin v1.3.1
```

## CI/CD Integration

The recommended GitHub Actions setup:

| Workflow | Trigger | Purpose |
|---|---|---|
| `CI` (`go-build.yaml`) | Push to `main`, PRs | Build, test, lint on all platforms |
| `Release` (`release.yaml`) | Push `v*` tag | Build binaries and publish GitHub release |

### Optional: Automated Version Tagging

For fully automated releases, add a workflow that tags `main` after merge:

```yaml
name: Auto Release

on:
  push:
    branches: [main]

jobs:
  tag:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Get last tag
        id: last_tag
        run: echo "tag=$(git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)" >> "$GITHUB_OUTPUT"

      - name: Calculate next version
        id: next
        run: |
          msg=$(git log -1 --format=%s)
          next=$(python bump-version.py "${{ steps.last_tag.outputs.tag }}" "$msg" | head -1)
          echo "version=$next" >> "$GITHUB_OUTPUT"

      - name: Create tag
        if: steps.next.outputs.version != steps.last_tag.outputs.tag
        run: |
          git tag "v${{ steps.next.outputs.version }}" -m "v${{ steps.next.outputs.version }}"
          git push origin "v${{ steps.next.outputs.version }}"
```

> **Note:** This auto-tagging workflow is optional. Many teams prefer manual tagging
> to control exactly when releases happen and to write meaningful release notes.

## Summary

| Step | Who | Action |
|---|---|---|
| Develop | Developer | Work on feature branch, commit with Conventional Commits |
| Review | Team | Create PR, review, merge to `main` |
| Release | Maintainer | Tag `main` with version, push tag |
| Publish | CI/CD | Build binaries, create GitHub release (automatic) |
| Install | User | `curl \| sh` always fetches the latest release |
