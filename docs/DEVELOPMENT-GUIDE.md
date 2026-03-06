# Development Guide: Using githooks for Commit Management and Releases

This guide explains how development teams use githooks together with `bump-version.py`
to manage commit messages, branch workflows, and software releases.

## How the Pieces Fit Together

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                          Developer Workstation                               │
│                                                                              │
│  githooks init  ──►  commit-msg hook installed                               │
│  githooks add   ──►  workspace configured (folder + Jira keys)               │
│                                                                              │
│  Every git commit  ──►  hook validates Conventional Commits format            │
│                    ──►  hook auto-inserts Jira ticket from branch name        │
└──────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                              Git Repository                                  │
│                                                                              │
│  feature/ABC-123-*  ──►  PR ──►  merge to main                              │
│  bugfix/ABC-456-*   ──►  PR ──►  merge to main                              │
│                                                                              │
│  bump-version.py --auto  ──►  determines next version from commit history    │
│  git tag v1.2.0          ──►  triggers release pipeline                      │
└──────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌──────────────────────────────────────────────────────────────────────────────┐
│                            CI/CD Pipeline                                    │
│                                                                              │
│  v* tag push  ──►  build artifacts  ──►  publish release                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Setup (One Time Per Developer)

### 1. Install githooks

```bash
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.sh | sh
```

### 2. Initialize and Add Your Workspace

```bash
githooks init
githooks add
```

The `add` command asks for:
- **Workspace name** — e.g. `my-team`
- **Jira project key** — e.g. `ABC` or `(ABC|DEF)` for multiple projects
- **Workspace folder** — e.g. `~/projects/my-team/`

Every Git repository under that folder is now protected by the commit-msg hook.

### 3. Place bump-version.py in Your Project

Copy `scripts/bump-version.py` into the root of your project repository so all team
members and the CI pipeline have access to it.

## Daily Development Workflow

### Step 1: Create a Feature Branch

Always include the Jira ticket in the branch name:

```bash
git checkout -b feature/ABC-123-add-user-auth
```

### Step 2: Commit Your Changes

Write commits following Conventional Commits. The hook does two things:

1. **Validates** — rejects messages that don't match `type(scope): description`
2. **Auto-inserts** — adds the Jira ticket from the branch name if missing

```bash
# You type this:
git commit -m "feat: add OAuth2 login flow"

# Hook rewrites it to:
# feat(ABC-123): add OAuth2 login flow
```

You can also include the ticket yourself:

```bash
git commit -m "feat(ABC-123): add OAuth2 login flow"     # accepted
git commit -m "fix(ABC-123): handle null token"           # accepted
git commit -m "feat(ABC-123)!: redesign auth API"         # accepted (breaking)
```

These are rejected:

```bash
git commit -m "added login"                               # no type prefix
git commit -m "feat: add login"                           # no ticket (unless on a ticket branch)
git commit -m "[ABC-123] add login"                       # wrong format
```

### Step 3: Push and Open a Pull Request

```bash
git push -u origin feature/ABC-123-add-user-auth
```

Create a PR with a Conventional Commits title:

```bash
gh pr create --title 'feat(ABC-123): add OAuth2 user authentication'
```

### Step 4: Merge to Main

After code review, merge the PR. Recommended: **squash merge** for a clean
single commit on `main`:

```bash
gh pr merge --squash
```

This creates one clean commit on `main`:
```
feat(ABC-123): add OAuth2 user authentication (#42)
```

## Release Workflow

### Option A: Manual Release (Recommended for Most Teams)

After merging one or more PRs to `main`:

```bash
git checkout main && git pull

# See what changed since last release
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Let bump-version.py determine the next version automatically
python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)

# Tag and release
git tag v1.3.0 -m 'v1.3.0 - Add OAuth2 auth, fix session handling'
git push origin v1.3.0
```

The `--auto` flag:
1. Reads the latest `v*` tag as the current version
2. Scans all commits since that tag
3. Picks the **highest-priority** bump (major > minor > patch)
4. Prints the next version

### Option B: CI-Driven Automatic Release

Add this to your project's CI pipeline to release on every merge to `main`:

```yaml
# .github/workflows/auto-release.yaml
name: Auto Release

on:
  push:
    branches: [main]

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-python@v5
        with:
          python-version: '3.x'

      - name: Calculate next version
        id: version
        run: |
          CURRENT=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
          NEXT=$(python bump-version.py --auto 2>/dev/null)
          echo "current=$CURRENT" >> "$GITHUB_OUTPUT"
          echo "next=$NEXT" >> "$GITHUB_OUTPUT"

      - name: Create release tag
        if: steps.version.outputs.next != steps.version.outputs.current
        run: |
          git tag "v${{ steps.version.outputs.next }}" -m "v${{ steps.version.outputs.next }}"
          git push origin "v${{ steps.version.outputs.next }}"
```

This triggers your existing release workflow (GoReleaser, Docker build, npm publish, etc.)
whenever the version changes.

### Option C: Release from a Release Branch

For teams that batch features into planned releases:

```bash
# Create release branch
git checkout -b release/1.3.0 main

# Final fixes on the release branch
git commit -m 'fix(ABC-789): patch edge case in auth'

# Merge to main and tag
git checkout main
git merge release/1.3.0
git tag v1.3.0 -m 'v1.3.0'
git push origin main --tags
```

## How bump-version.py Works

### Single Commit Mode

Useful in CI to classify one commit at a time:

```bash
python bump-version.py 1.2.3 'feat(ABC-123): add dashboard'
# Output: 1.3.0
```

### Multi-Commit Mode (Stdin)

Pipe multiple commit messages. The highest-priority bump wins:

```bash
git log v1.2.0..HEAD --format=%s | python bump-version.py 1.2.0
# Output: 1.3.0
```

### Auto Mode

Reads everything from git — no arguments needed:

```bash
python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)
```

> **CI usage:** The version is printed to **stdout**, the summary to **stderr**.
> Capture just the version in a pipeline:
> ```bash
> VERSION=$(python bump-version.py --auto)
> ```

### Bump Priority

When multiple commits are analyzed, the **highest-priority** bump applies:

```
feat(ABC-1): add login           → minor
fix(ABC-2): null pointer         → patch
docs(ABC-3): update README       → patch
feat(ABC-4)!: redesign auth API  → major   ← wins

Result: major bump (1.2.3 → 2.0.0)
```

The full priority table:

| Priority | Trigger | Bump | Commits |
|---|---|---|---|
| 1 (highest) | `!` or `BREAKING CHANGE:` | Major | Any type |
| 2 | `feat` | Minor | — |
| 3 | `fix`, `docs`, `refactor`, `perf`, `build`, `chore`, `revert` | Patch | — |
| — | `style`, `test`, `ci` | None | No release |

## Hotfix Process

For critical production fixes:

```bash
git checkout -b hotfix/ABC-999-critical-fix main
git commit -m 'fix(ABC-999): patch SQL injection vulnerability'
git push -u origin hotfix/ABC-999-critical-fix

# Fast-track PR merge
gh pr create --title 'fix(ABC-999): patch SQL injection' && gh pr merge --squash

# Immediate patch release
git checkout main && git pull
python bump-version.py --auto    # → 1.3.1
git tag v1.3.1 -m 'v1.3.1 - Security patch'
git push origin v1.3.1
```

## Example: Complete Sprint Cycle

```bash
# === Sprint work ===

# Developer A: new feature
git checkout -b feature/ABC-100-dashboard
git commit -m 'feat: add analytics dashboard'
git commit -m 'feat: add export to CSV'
git push && gh pr create --title 'feat(ABC-100): analytics dashboard'

# Developer B: bug fixes
git checkout -b bugfix/ABC-200-login-crash
git commit -m 'fix: handle expired session token'
git push && gh pr create --title 'fix(ABC-200): handle expired session'

# Developer C: docs + refactor
git checkout -b feature/ABC-300-cleanup
git commit -m 'refactor: extract auth middleware'
git commit -m 'docs: update API reference'
git push && gh pr create --title 'refactor(ABC-300): extract auth middleware'

# === All PRs reviewed and merged to main ===

git checkout main && git pull
git log v1.2.0..HEAD --format=%s
# feat(ABC-100): analytics dashboard (#10)
# fix(ABC-200): handle expired session (#11)
# refactor(ABC-300): extract auth middleware (#12)

python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 3 commits: minor: 1, patch: 2)

git tag v1.3.0 -m 'v1.3.0 - Analytics dashboard, session fix, auth refactor'
git push origin v1.3.0
# → Release pipeline builds and publishes automatically
```

## Summary

| Phase | Tool | What Happens |
|---|---|---|
| Setup | `githooks init` + `githooks add` | Hook installed, workspace configured |
| Develop | `git commit` | Hook enforces Conventional Commits + Jira ticket |
| Review | `git push` + PR | Clean, validated commit history |
| Release | `bump-version.py --auto` | Scans commits, calculates next version |
| Publish | `git tag` + `git push` | CI builds and publishes release |
