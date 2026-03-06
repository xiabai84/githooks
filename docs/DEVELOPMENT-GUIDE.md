# Development Guide

A practical guide for teams using githooks to enforce consistent commit messages,
manage branches, and automate releases.

## Table of Contents

- [Commit Code of Conduct](#commit-code-of-conduct)
- [Setup](#setup)
- [Branch Strategy](#branch-strategy)
- [Daily Development Workflow](#daily-development-workflow)
- [Commit Message Best Practices](#commit-message-best-practices)
- [Code Review Guidelines](#code-review-guidelines)
- [Release Workflow](#release-workflow)
- [How bump-version.py Works](#how-bump-versionpy-works)
- [Docker Usage in CI](#docker-usage-in-ci)
- [Hotfix Process](#hotfix-process)
- [Example: Complete Sprint Cycle](#example-complete-sprint-cycle)
- [Summary](#summary)

---

## Commit Code of Conduct

Every team member agrees to the following rules when committing code.

### 1. Every Commit Tells a Story

Each commit should represent **one logical change**. A reader should understand the purpose
of the commit from the first line alone — without opening a diff.

- **Good:** `feat(MOB-123): add password reset email`
- **Bad:** `updates`, `WIP`, `fix stuff`, `Monday work`

### 2. Conventional Commits Are Mandatory

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<JIRA-TICKET>): <description>
```

The githooks commit-msg hook enforces this automatically. There are no exceptions.

### 3. One Ticket Per Branch

Each feature or bugfix branch maps to exactly one Jira ticket. The ticket must appear
in the branch name so the hook can auto-insert it into commit messages.

```
feature/MOB-123-add-login       ✓  one ticket, clear purpose
bugfix/PAY-456-fix-timeout      ✓  one ticket, clear purpose
feature/various-fixes           ✗  no ticket, unclear scope
feature/MOB-123-and-MOB-456     ✗  two tickets, split into two branches
```

### 4. Never Rewrite Shared History

- Never force-push to `main` or shared branches
- Never rebase branches that others are working on
- Use `git revert` instead of `git reset` for published commits

### 5. Keep Commits Small and Focused

- One concern per commit — don't mix formatting with logic changes
- Commit early, commit often — smaller commits are easier to review and revert
- Avoid commits that touch more than 300 lines unless unavoidable (e.g. generated code)

### 6. Breaking Changes Are Announced

Breaking changes must be explicitly marked with `!` or a `BREAKING CHANGE:` footer.
Never sneak a breaking change into a regular `fix` or `feat` commit.

```bash
feat(MOB-123)!: redesign authentication API

BREAKING CHANGE: /api/auth/login now requires client_id parameter
```

### 7. Tests Accompany Code

- New features include tests in the same commit or a follow-up `test` commit
- Bug fixes include a regression test that proves the fix works
- Refactors must not break existing tests

### 8. Clean Up Before Pushing

- Squash WIP commits into meaningful units before opening a PR
- Remove debug statements, commented-out code, and temporary files
- Run the test suite locally before pushing

---

## Setup

### One Time Per Developer

#### 1. Install githooks

```bash
# Linux / macOS
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.ps1 | iex
```

#### 2. Initialize and Add Your Workspace

```bash
githooks init
githooks add
```

The `add` command asks for:
- **Workspace name** — a descriptive label (e.g. `mobile-app`)
- **Jira project key** — regex of accepted ticket prefixes (e.g. `MOB` or `(MOB|PAY)`)
- **Workspace folder** — parent directory containing your repos (e.g. `~/projects/mobile/`)

Every Git repository under that folder is now protected by the commit-msg hook.

#### 3. Place bump-version.py in Your Project

Copy `scripts/bump-version.py` into your project repository so all team
members and the CI pipeline have access to it.

---

## Branch Strategy

We recommend **trunk-based development** with short-lived feature branches.

```
main ─────●─────●─────●─────●─────●──── (always deployable)
           \   /       \   /       \
            ●─●         ●─●        ● ── feature/MOB-100-dashboard
          feature/    bugfix/      (merged within 1-3 days)
          MOB-80      PAY-200
```

### Branch Naming Convention

| Branch | Pattern | Example |
|---|---|---|
| Feature | `feature/<TICKET>-<short-description>` | `feature/MOB-123-add-login` |
| Bugfix | `bugfix/<TICKET>-<short-description>` | `bugfix/PAY-456-fix-timeout` |
| Hotfix | `hotfix/<TICKET>-<short-description>` | `hotfix/PAY-999-sql-injection` |
| Release | `release/<version>` | `release/2.0.0` |

### Rules

1. **`main` is always deployable.** Never push broken code directly to main.
2. **Branches are short-lived.** Merge within 1–3 days. Long-lived branches cause merge conflicts.
3. **Branch from `main`, merge back to `main`.** No branch-to-branch merges.
4. **Delete branches after merge.** Keep the branch list clean.
5. **Use squash merges for features.** This creates one clean commit on `main` per feature.

### When to Use Release Branches

Release branches (`release/X.Y.Z`) are optional. Use them when:
- You need a stabilization period before a release
- Multiple teams need to coordinate a release date
- You must support multiple release versions simultaneously

For most teams, releasing directly from `main` with tags is simpler and recommended.

---

## Daily Development Workflow

### Step 1: Create a Feature Branch

```bash
git checkout main && git pull
git checkout -b feature/MOB-123-add-user-auth
```

### Step 2: Commit Your Changes

The hook validates every commit and auto-inserts the Jira ticket from the branch name:

```bash
# You type:
git commit -m "feat: add OAuth2 login flow"

# Hook rewrites to:
# feat(MOB-123): add OAuth2 login flow
```

You can also include the ticket explicitly:

```bash
git commit -m "feat(MOB-123): add OAuth2 login flow"     # accepted
git commit -m "fix(MOB-123): handle null token"           # accepted
git commit -m "feat(MOB-123)!: redesign auth API"         # accepted (breaking)
```

These are rejected:

```bash
git commit -m "added login"                               # no type prefix
git commit -m "feat: add login"                           # no ticket (on main)
git commit -m "[MOB-123] add login"                       # wrong format
```

### Step 3: Push and Open a Pull Request

```bash
git push -u origin feature/MOB-123-add-user-auth
gh pr create --title 'feat(MOB-123): add OAuth2 user authentication'
```

**PR title must follow Conventional Commits format** — this becomes the merge commit
on `main` when using squash merge.

### Step 4: Code Review and Merge

After approval, use squash merge for a clean history:

```bash
gh pr merge --squash --delete-branch
```

This creates one commit on `main`:
```
feat(MOB-123): add OAuth2 user authentication (#42)
```

---

## Commit Message Best Practices

### Structure

```
<type>(<TICKET>): <short description>     ← subject line (max 72 chars)
                                           ← blank line
<optional body>                            ← explain WHY, not WHAT
                                           ← blank line
<optional footer>                          ← BREAKING CHANGE, references
```

### Subject Line Rules

| Rule | Example |
|---|---|
| Use imperative mood | `add login`, not `added login` or `adds login` |
| Don't capitalize after colon | `feat(MOB-1): add login`, not `feat(MOB-1): Add login` |
| No period at the end | `feat(MOB-1): add login`, not `feat(MOB-1): add login.` |
| Max 72 characters | Keeps `git log --oneline` readable |

### When to Use a Body

Add a body when the **why** isn't obvious from the subject:

```
fix(PAY-456): reject expired tokens before database lookup

Previously, expired tokens were passed to the database layer, causing
unnecessary queries and misleading error messages. This check moves
token validation to the middleware layer.
```

Don't add a body for self-explanatory changes:

```
docs(MOB-789): fix typo in API reference
```

### Type Selection Guide

| If you... | Use |
|---|---|
| Add new functionality for the user | `feat` |
| Fix a bug that affects the user | `fix` |
| Change only documentation | `docs` |
| Restructure code without changing behavior | `refactor` |
| Improve performance | `perf` |
| Add or update tests | `test` |
| Change build scripts or dependencies | `build` |
| Change CI pipeline configuration | `ci` |
| Format code (whitespace, semicolons) | `style` |
| Update dependencies, configs, maintenance | `chore` |
| Revert a previous commit | `revert` |

### Common Mistakes

| Mistake | Fix |
|---|---|
| `feat(MOB-1): feat add login` | Don't repeat the type in the description |
| `fix(MOB-1): fixed bug` | Use imperative: `fix null pointer in auth` |
| `refactor(MOB-1): refactoring` | Be specific: `extract auth into middleware` |
| `chore: misc changes` | Be specific: `update eslint to v9` |
| Mixing concerns in one commit | Split into separate commits |

---

## Code Review Guidelines

### For Authors

1. **Self-review before requesting** — read your own diff as if you're the reviewer
2. **Keep PRs small** — under 400 lines of diff when possible
3. **Write a clear PR description** — summarize what changed and why
4. **Link the Jira ticket** — reviewers need context
5. **Respond to feedback promptly** — don't let PRs go stale

### For Reviewers

1. **Review within 24 hours** — blocking a PR blocks the developer
2. **Check the commit message** — does the PR title follow Conventional Commits?
3. **Focus on correctness and clarity** — not style preferences
4. **Approve when it's good enough** — don't chase perfection
5. **Use "Request Changes" sparingly** — reserve for actual issues, not nitpicks

### Merge Checklist

Before merging, verify:
- [ ] PR title follows `type(TICKET): description` format
- [ ] CI passes (tests, linting, build)
- [ ] Breaking changes are marked with `!` or `BREAKING CHANGE:`
- [ ] No unrelated changes are included
- [ ] Branch is up to date with `main`

---

## Release Workflow

Choose the approach that fits your team. The key difference is **who creates the git tag** —
you manually, or CI automatically.

### Option A: Manual Tag Release

You control exactly when a release happens by creating a git tag yourself.
Best for teams that want explicit control over release timing.

```bash
git checkout main && git pull

# See what changed since last release
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Let bump-version.py determine the next version
python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)

# Tag and push — this triggers the release pipeline
git tag v1.3.0 -m 'v1.3.0 - Add OAuth2 auth, fix session handling'
git push origin v1.3.0
```

**How it works:**
1. `bump-version.py --auto` reads the latest `v*` tag and scans commits since then
2. It picks the highest-priority bump (major > minor > patch)
3. You create the tag and push it
4. CI detects the `v*` tag and builds/publishes the release

### Option B: CI Auto-Release (No Manual Tags)

CI automatically determines the version and creates the tag on every merge to `main`.
Best for teams that want continuous delivery without manual steps.

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
          NEXT=$(python bump-version.py --auto)
          echo "current=${CURRENT#v}" >> "$GITHUB_OUTPUT"
          echo "next=$NEXT" >> "$GITHUB_OUTPUT"

      - name: Create release tag
        if: steps.version.outputs.next != steps.version.outputs.current
        run: |
          git tag "v${{ steps.version.outputs.next }}" -m "v${{ steps.version.outputs.next }}"
          git push origin "v${{ steps.version.outputs.next }}"
```

> **No tag, no release:** If all commits are `style`, `test`, or `ci`, no version bump
> occurs and no tag is created.

### Option C: Release Without Git Tags

For projects that deploy from `main` directly and track the version in a file.

```bash
git checkout main && git pull

# Pass the current version and a commit message directly
python bump-version.py 1.2.3 'feat(MOB-123): add OAuth2 login'
# Output: 1.3.0

# Or scan multiple recent commits via stdin
git log --format=%s -5 | python bump-version.py 1.2.3
# Output: 1.3.0
```

CI pipeline using a VERSION file:

```yaml
# .github/workflows/deploy.yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Get current version
        id: current
        run: echo "version=$(cat VERSION)" >> "$GITHUB_OUTPUT"

      - name: Calculate next version
        id: next
        run: |
          COMMIT_MSG=$(git log -1 --format=%s)
          NEXT=$(python bump-version.py ${{ steps.current.outputs.version }} "$COMMIT_MSG")
          echo "version=$NEXT" >> "$GITHUB_OUTPUT"

      - name: Update version file
        if: steps.next.outputs.version != steps.current.outputs.version
        run: |
          echo "${{ steps.next.outputs.version }}" > VERSION
          git add VERSION
          git commit -m "chore: bump version to ${{ steps.next.outputs.version }}"
          git push

      - name: Deploy
        run: echo "Deploying version ${{ steps.next.outputs.version }}"
```

### Option D: Release from a Release Branch

For teams that batch features into planned releases:

```bash
# Create release branch
git checkout -b release/1.3.0 main

# Final fixes on the release branch
git commit -m 'fix(PAY-789): patch edge case in auth'

# Merge to main and tag
git checkout main
git merge release/1.3.0
git tag v1.3.0 -m 'v1.3.0'
git push origin main --tags
```

### Our Recommendation

For most teams, we recommend **Option A (Manual Tag Release)** because:
- You have full control over when releases happen
- The commit history on `main` is the single source of truth
- `bump-version.py --auto` eliminates guesswork about version numbers
- It works with any CI system that supports tag-triggered builds

Move to **Option B** when your team is confident in test coverage and wants
faster delivery. Move to **Option D** only when you need parallel release stabilization.

---

## How bump-version.py Works

### Three Modes

| Mode | Command | When to Use |
|---|---|---|
| **Auto** | `python bump-version.py --auto` | Projects using git tags |
| **Single** | `python bump-version.py 1.2.3 'feat(...): ...'` | CI per-commit classification |
| **Stdin** | `git log ... \| python bump-version.py 1.2.3` | Batch scanning, no tags |

### Auto Mode (Requires Git Tags)

```bash
python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)
```

If no tags exist, it scans all commits and starts from `0.0.0`.

> **CI usage:** The version is printed to **stdout**, the summary to **stderr**.
> Capture just the version in a pipeline:
> ```bash
> VERSION=$(python bump-version.py --auto)
> ```

### Single Commit Mode (No Tags Needed)

```bash
python bump-version.py 1.2.3 'feat(MOB-123): add dashboard'
# Output: 1.3.0

python bump-version.py 1.2.3 'fix(PAY-456): null pointer'
# Output: 1.2.4

python bump-version.py 1.2.3 'feat(MOB-789)!: redesign auth'
# Output: 2.0.0
```

### Multi-Commit Mode via Stdin (No Tags Needed)

```bash
git log --format=%s -10 | python bump-version.py 1.2.3
# Output: 1.3.0
```

### Bump Priority

When multiple commits are analyzed, the **highest-priority** bump applies:

```
feat(MOB-1): add login           → minor
fix(PAY-2): null pointer         → patch
docs(MOB-3): update README       → patch
feat(PAY-4)!: redesign auth API  → major   ← wins

Result: major bump (1.2.3 → 2.0.0)
```

| Priority | Trigger | Bump |
|---|---|---|
| 1 (highest) | `!` or `BREAKING CHANGE:` | Major |
| 2 | `feat` | Minor |
| 3 | `fix`, `docs`, `refactor`, `perf`, `build`, `chore`, `revert` | Patch |
| — | `style`, `test`, `ci` | None (no release) |

### PowerShell Equivalent

All modes are also available in `bump-version.ps1`:

```powershell
# Auto mode
.\scripts\bump-version.ps1 -Auto

# Single commit
.\scripts\bump-version.ps1 -Version 1.2.3 -Message "feat(MOB-123): add dashboard"

# Stdin pipe
git log --format=%s -10 | .\scripts\bump-version.ps1 -Version 1.2.3
```

---

## Docker Usage in CI

Use the githooks Docker image to run `bump-version.py` without installing Python:

```yaml
jobs:
  version:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - name: Build githooks image
        run: docker build -t githooks .
      - name: Calculate next version
        run: |
          VERSION=$(docker run --rm -v "$(pwd):/repo" -w /repo githooks \
            sh -c "python3 /usr/local/bin/bump-version.py --auto")
          echo "Next version: $VERSION"
```

---

## Hotfix Process

For critical production fixes that cannot wait for the next sprint.

### With Tags

```bash
git checkout -b hotfix/PAY-999-critical-fix main
git commit -m 'fix(PAY-999): patch SQL injection vulnerability'
git push -u origin hotfix/PAY-999-critical-fix

# Fast-track PR — skip full review cycle, get one senior approval
gh pr create --title 'fix(PAY-999): patch SQL injection' && gh pr merge --squash

# Immediate patch release
git checkout main && git pull
python bump-version.py --auto    # → 1.3.1
git tag v1.3.1 -m 'v1.3.1 - Security patch'
git push origin v1.3.1
```

### Without Tags

```bash
git checkout -b hotfix/PAY-999-critical-fix main
git commit -m 'fix(PAY-999): patch SQL injection vulnerability'
git push -u origin hotfix/PAY-999-critical-fix

gh pr create --title 'fix(PAY-999): patch SQL injection' && gh pr merge --squash

# Calculate version manually
git checkout main && git pull
python bump-version.py 1.3.0 'fix(PAY-999): patch SQL injection vulnerability'
# Output: 1.3.1
```

---

## Example: Complete Sprint Cycle

### With Tags

```bash
# === Sprint work ===

# Developer A: new feature
git checkout -b feature/MOB-100-dashboard
git commit -m 'feat: add analytics dashboard'
git commit -m 'feat: add export to CSV'
git push && gh pr create --title 'feat(MOB-100): analytics dashboard'

# Developer B: bug fix
git checkout -b bugfix/PAY-200-login-crash
git commit -m 'fix: handle expired session token'
git push && gh pr create --title 'fix(PAY-200): handle expired session'

# Developer C: refactor
git checkout -b feature/MOB-300-cleanup
git commit -m 'refactor: extract auth middleware'
git commit -m 'docs: update API reference'
git push && gh pr create --title 'refactor(MOB-300): extract auth middleware'

# === All PRs reviewed and merged ===

git checkout main && git pull

python bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 3 commits: minor: 1, patch: 2)

git tag v1.3.0 -m 'v1.3.0 - Analytics dashboard, session fix, auth refactor'
git push origin v1.3.0
# → Release pipeline builds and publishes automatically
```

### Without Tags

```bash
# === Same sprint work, all PRs merged to main ===

git checkout main && git pull

# Current version from VERSION file
CURRENT=1.2.0
git log --format=%s -3 | python bump-version.py $CURRENT
# Output: 1.3.0

echo "1.3.0" > VERSION
git add VERSION
git commit -m 'chore: bump version to 1.3.0'
git push
# → CI deploys from main automatically
```

---

## Summary

| Phase | Tool | What Happens |
|---|---|---|
| Setup | `githooks init` + `githooks add` | Hook installed, workspace configured |
| Develop | `git commit` | Hook enforces Conventional Commits + Jira ticket |
| Review | `git push` + PR | Clean, validated commit history |
| Release (with tags) | `bump-version.py --auto` + `git tag` | Auto-detect version, tag triggers CI |
| Release (without tags) | `bump-version.py <version> <msg>` | Manual version bump, CI deploys from main |

### Quick Reference Card

```bash
# Start a feature
git checkout main && git pull
git checkout -b feature/MOB-123-my-feature

# Commit (hook auto-inserts ticket)
git commit -m "feat: add new feature"

# Push and create PR
git push -u origin feature/MOB-123-my-feature
gh pr create --title 'feat(MOB-123): add new feature'

# After merge — release
git checkout main && git pull
python bump-version.py --auto
git tag v$(python bump-version.py --auto) -m "release"
git push origin --tags
```
