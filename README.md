# githooks

A CLI tool that manages Git commit-msg hooks across multiple workspaces. It enforces
[Conventional Commits](https://www.conventionalcommits.org/) format with mandatory Jira issue keys,
ensuring consistent commit messages across your team.

## Table of Contents

- [Commit Message Format](#commit-message-format)
- [Hook Behavior](#hook-behavior)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
  - [Option A: Download Pre-built Binary (Linux/macOS)](#option-a-download-pre-built-binary-linuxmacos)
  - [Option B: Windows Installation](#option-b-windows-installation)
  - [Option C: Build from Source](#option-c-build-from-source)
  - [Option D: Docker (CI Pipelines)](#option-d-docker-ci-pipelines)
  - [Option E: Single Repository (No CLI)](#option-e-single-repository-no-cli)
- [Getting Started](#getting-started)
- [Managing Workspaces](#managing-workspaces)
  - [Adding a Workspace](#adding-a-workspace)
  - [Updating a Workspace](#updating-a-workspace)
  - [Listing Workspaces](#listing-workspaces)
  - [Deleting a Workspace](#deleting-a-workspace)
  - [Uninstalling](#uninstalling)
- [How It Works](#how-it-works)
  - [Automatic Detection](#automatic-detection)
  - [Multiple Jira Projects](#multiple-jira-projects)
  - [File Structure](#file-structure)
  - [Git Configuration](#git-configuration)
- [Troubleshooting](#troubleshooting)
- [Releasing a New Version](#releasing-a-new-version)
- [License](#license)

## Commit Message Format

All commit messages must follow Conventional Commits with a Jira ticket in the scope:

```
<type>(<JIRA-TICKET>): <description>

[optional body]

[optional footer]
```

**Example:**

```
feat(PAY-123): add OAuth2 login for mobile clients

- Implement token refresh flow
- Add secure token storage

BREAKING CHANGE: /api/auth/login now requires client_id parameter
```

### Allowed Types

| Type | Purpose | Version Bump |
|---|---|---|
| `feat` | New feature for the user | Minor (`0.X.0`) |
| `fix` | Bug fix | Patch (`0.0.X`) |
| `perf` | Performance improvement | Patch (`0.0.X`) |
| `revert` | Reverts a previous commit | Patch (`0.0.X`) |
| `build` | Build system or external dependency changes | Patch (`0.0.X`) |
| `chore` | Maintenance tasks (dependency updates, configs) | Patch (`0.0.X`) |
| `docs` | Documentation changes only | — |
| `style` | Formatting, whitespace — no logic change | — |
| `refactor` | Code restructuring without behavior change | — |
| `test` | Adding or correcting tests | — |
| `ci` | CI/CD pipeline configuration | — |

Append `!` before the colon or add a `BREAKING CHANGE:` footer to indicate a breaking change.
Any type with `!` or a breaking change footer triggers a **Major** (`X.0.0`) version bump.

```
feat(PAY-123)!: redesign authentication API
```

> **Semantic Versioning summary:** Given `MAJOR.MINOR.PATCH` — a breaking change increments MAJOR, a new feature increments MINOR, and all other releasable changes increment PATCH. Types marked `—` do not trigger a release on their own.

> **Why this classification?** Only changes that affect the **shipped artifact** trigger a release: `fix` corrects user-facing bugs, `perf` improves user-facing performance, `revert` undoes a change that may have introduced issues, `build` can affect the binary through toolchain updates, and `chore` often updates dependencies that change the compiled binary or Docker image. Types like `docs`, `refactor`, `style`, `test`, and `ci` do not alter the distributed artifact, so they do not warrant a new release.

## Hook Behavior

| Commit Message | Branch | Result |
|---|---|---|
| `add new feature` | any | **Rejected** — missing type prefix |
| `[PAY-123] add feature` | any | **Rejected** — not Conventional Commits format |
| `feat: add feature` | `main` | **Rejected** — missing Jira ticket |
| `feat: add feature` | `feature/PAY-123-login` | **Rewritten** to `feat(PAY-123): add feature` |
| `feat(PAY-123): add feature` | any | **Accepted** |
| `refactor(API-99)!: redesign auth` | any | **Accepted** |
| `Merge branch 'main'` | any | **Accepted** — merge commits skip validation |

When working on a branch that contains a Jira issue key (e.g. `feature/PAY-123-login`),
the hook automatically inserts the ticket into the scope position. Multi-line message bodies
and footers are preserved during rewriting.

## Prerequisites

- **Git** 2.13 or later (required for `includeIf` support)
- **Bash** 4.0 or later (included with Git for Windows)
- **Go** 1.22 or later (only for building from source)

### Supported Platforms

| OS | Architecture | Format |
|---|---|---|
| Linux | amd64, arm64 | tar.gz |
| macOS | amd64, arm64 | tar.gz |
| Windows | amd64, arm64 | zip |

## Installation

### Option A: Download Pre-built Binary (Linux/macOS)

```bash
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.sh | sh
```

The install script automatically detects your operating system and architecture,
downloads the binary, and installs it to `~/.local/bin/`.

> **Note:** If `~/.local/bin/` is not in your PATH, the script will print the
> command to add it. For example:
> ```bash
> echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
> source ~/.zshrc
> ```

### Option B: Windows Installation

**PowerShell (recommended):**

```powershell
irm https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.ps1 | iex
```

**Git Bash / MSYS2:**

```bash
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.sh | sh
mv githooks.exe ~/bin/
```

**Manual download:**

1. Download the latest `githooks-*-windows-amd64.zip` from the [Releases](https://github.com/xiabai84/githooks/releases) page
2. Extract `githooks.exe` to a directory in your PATH

> **Note:** The commit-msg hook script requires Bash. On Windows, this is provided
> by [Git for Windows](https://gitforwindows.org/) (Git Bash), which is available
> by default in any Git installation.

For detailed Windows setup instructions including IDE integration and troubleshooting, see the [Windows Configuration Guide](docs/WINDOWS.md).

### Option C: Build from Source

```bash
git clone https://github.com/xiabai84/githooks.git
cd githooks
go build -o githooks .
```

To embed version information into the binary:

```bash
go build -ldflags "-s -w \
  -X github.com/xiabai84/githooks/buildinfo.version=$(git describe --tags) \
  -X github.com/xiabai84/githooks/buildinfo.gitCommit=$(git rev-parse HEAD)" \
  -o githooks .
```

On Windows, use `go build -o githooks.exe .` instead.

Verify the installation:

```bash
githooks version
```

### Option D: Docker (CI Pipelines)

Use the Docker image to run githooks or `bump-version.py` in CI without local installation:

```bash
# Build the image
docker build -t githooks .

# Run githooks commands
docker run --rm githooks version

# Use bump-version.py inside the container
docker run --rm -v "$(pwd):/repo" -w /repo githooks \
  sh -c "python3 /usr/local/bin/bump-version.py --auto"
```

**GitHub Actions example:**

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

### Option E: Single Repository (No CLI)

If you only need the hook in a single repository without installing the CLI:

```bash
# Install in the current repository
./scripts/install-jira-git-hook

# Install in multiple repositories at once
./scripts/install-jira-git-hook repo1 repo2 repo3

# Restrict to specific Jira projects
./scripts/install-jira-git-hook --projects=PAY,MOB

# Override existing hooks without prompting
./scripts/install-jira-git-hook --yes
```

Full usage:

```
Usage: install-jira-git-hook [ OPTIONS ] [ dir ... ]

Install a Git hook to search for a Jira issue key
in the commit message or branch name.

Options:
  -y, --yes       Override existing commit-msg files
  -p, --projects  Let the hook only accept keys for these Jira projects
                  e.g. --projects=PAY,MOB,INFRA
  -h, --help      Show this help
```

## Getting Started

After installation, set up githooks in three steps:

### 1. Initialize

Run the init command once to create the required folder structure and configuration files:

```bash
githooks init
```

This creates:
- `~/.githooks/commit-msg` — the shared hook script
- `~/.githooks/config/` — configuration directory
- `~/.githooks/config/githooks.json` — workspace registry
- `~/.gitconfig` — created if it does not exist

### 2. Add a Workspace

```bash
githooks add
```

The interactive prompt guides you through:
1. **Workspace name** — a descriptive label for this workspace (e.g. `mobile-app`)
2. **Jira project key** — regex matching allowed ticket prefixes (e.g. `MOB` or `(MOB|PAY)`)
3. **Workspace folder** — the parent directory containing your repositories (e.g. `~/projects/mobile/`)

> **Tip:** The workspace name is just a label to help you identify the workspace — it does not
> need to match your Jira project key. Use something descriptive like `mobile-app`, `backend-api`,
> or `data-pipeline`. The Jira project key (e.g. `MOB`, `PAY`) controls which ticket prefixes are
> accepted in commit messages.

### 3. Start Committing

Any Git repository under the configured workspace folder now enforces the commit message rules automatically. No per-repository setup is needed.

```bash
cd ~/projects/mobile/my-repo
git commit -m "feat(MOB-42): implement dashboard"   # Accepted
git commit -m "quick fix"                             # Rejected
```

## Managing Workspaces

A **workspace** maps a folder on your machine to a set of Jira project keys. Every Git
repository inside that folder automatically gets the commit-msg hook with the configured keys.

Each workspace has three properties:

| Property | Purpose | Example |
|---|---|---|
| **Name** | A human-readable label to identify the workspace | `mobile-app` |
| **Jira project key** | Regex of accepted Jira ticket prefixes | `MOB` or `(MOB\|PAY)` |
| **Folder** | Parent directory containing your Git repositories | `~/projects/mobile/` |

The name is only used for display and file naming — it does not affect Git behavior.
The Jira project key determines which ticket prefixes (e.g. `MOB-123`, `PAY-456`) are
valid in commit messages for repositories under that folder.

### Adding a Workspace

```bash
githooks add
```

Interactively creates a new workspace. After confirmation, githooks:
- Appends the workspace to `~/.githooks/config/githooks.json`
- Creates `~/.githooks/config/gitconfig-<name>` with the hooks path and Jira project key
- Appends an `includeIf` directive to `~/.gitconfig`

### Updating a Workspace

```bash
githooks update
```

Interactively select a workspace to modify. You can change:
- **Name** — renames the workspace and its gitconfig file
- **Jira project key** — updates the accepted ticket patterns (e.g. `MOB` → `(MOB|PAY)`)
- **Folder** — moves the workspace to a different directory path

Each field is pre-filled with the current value — press Enter to keep it unchanged.
After confirming, githooks updates all related config files automatically.

### Listing Workspaces

```bash
githooks list
```

Displays all managed workspaces with their names, folders, and Jira project keys.
The list supports search — start typing to filter by name.
The workspace matching the current directory is pre-selected.

### Deleting a Workspace

```bash
githooks delete
```

Interactively select a workspace to remove. After confirmation, githooks:
- Removes the workspace entry from `githooks.json`
- Deletes the workspace-specific `gitconfig-<name>` file
- Removes the corresponding `includeIf` block from `~/.gitconfig`

### Uninstalling

```bash
githooks uninstall
```

Completely removes all files and configuration managed by githooks. After confirmation, githooks:
- Removes all `includeIf` blocks managed by githooks from `~/.gitconfig`
- Deletes the entire `~/.githooks/` directory (hook script, workspace configs, registry)

## How It Works

### Automatic Detection

githooks leverages Git's [`includeIf`](https://git-scm.com/docs/git-config#_conditional_includes)
directive to automatically apply hooks based on repository location:

```gitconfig
[includeIf "gitdir:~/projects/mobile/"]
    path = .githooks/config/gitconfig-mobile-app
```

Every Git repository under `~/projects/mobile/` — including repositories cloned in the future —
automatically inherits the hook configuration. No per-repository setup is required.

Verify the configuration is active for any repository:

```bash
cd ~/projects/mobile/any-repo
git config --get core.hooksPath       # ~/.githooks
git config --get user.jiraProjects    # MOB
```

### Multiple Jira Projects

A single workspace can accept tickets from multiple Jira projects using a regex:

```
Jira project key RegEx: (MOB|PAY|INFRA)
```

This accepts commits like `feat(MOB-1): ...`, `fix(PAY-42): ...`, or `chore(INFRA-7): ...`.

**Automatic merging:** If you add a new workspace with a folder that already has a workspace
configured, githooks automatically merges the Jira project keys into a single regex pattern.
For example, adding `PAY` to a workspace that already tracks `MOB` produces `(MOB|PAY)` —
no duplicate configuration is created.

```bash
githooks add   # workspace "mobile-app",  key MOB, folder ~/projects/mobile/
githooks add   # workspace "payments",    key PAY, folder ~/projects/mobile/
# Result: single workspace "mobile-app" with key (MOB|PAY)
```

Alternatively, create separate workspaces with more specific folder paths.
Git resolves `includeIf` directives using the longest matching path, so a workspace
at `~/projects/mobile/` takes precedence over a workspace at `~/projects/`.

### File Structure

After initializing and adding workspaces **mobile-app** and **backend-api**, the following
structure is created:

```
~
├── .gitconfig
└── .githooks/
    ├── commit-msg
    └── config/
        ├── gitconfig-mobile-app
        ├── gitconfig-backend-api
        └── githooks.json
```

### Git Configuration

**`~/.gitconfig`** — conditional includes based on repository location:

```gitconfig
[includeIf "gitdir:~/projects/mobile/"]
    path = .githooks/config/gitconfig-mobile-app
[includeIf "gitdir:~/projects/backend/"]
    path = .githooks/config/gitconfig-backend-api
```

**`~/.githooks/config/gitconfig-mobile-app`** — per-workspace settings:

```gitconfig
[core]
    hooksPath=~/.githooks
[user]
    jiraProjects=MOB
```

- `core.hooksPath` points all repositories to the shared hook in `~/.githooks`
- `user.jiraProjects` is a custom Git variable (not part of Git core) that the
  commit-msg hook reads to determine which Jira project keys are valid

## Troubleshooting

**Hook is not triggered**

Ensure the workspace folder path ends with `/` and matches the location of your repositories:

```bash
git config --show-origin --get core.hooksPath
```

If nothing is returned, the `includeIf` directive is not matching. Check:
- The folder path in `~/.gitconfig` matches your repository location
- The path uses `~/` (not an absolute path with your username)

**"Please execute 'githooks init' first"**

Run `githooks init` to create the required directory structure and configuration files.

**Commit rejected but message looks correct**

Check that the type keyword is one of the allowed types (see [Allowed Types](#allowed-types))
and that the Jira ticket follows the format `PROJECT-NUMBER` (e.g. `MOB-123`).

**Branch ticket not auto-inserted**

The branch name must contain a Jira ticket pattern (e.g. `feature/MOB-123-description`).
The hook extracts tickets matching the configured project keys.

## Releasing a New Version

githooks uses [GoReleaser](https://goreleaser.com/) and GitHub Actions to build and publish
releases automatically. A new release is triggered by pushing a Git tag.

### 1. Determine the Version Bump

Follow the [Allowed Types](#allowed-types) table to determine the correct version bump
based on the commits since the last release:

| Change | Bump | Example |
|---|---|---|
| Breaking change (`!` or `BREAKING CHANGE:`) | Major | `1.2.3` → `2.0.0` |
| New feature (`feat`) | Minor | `1.2.3` → `1.3.0` |
| Bug fix, docs, refactor, etc. | Patch | `1.2.3` → `1.2.4` |

You can use the included helper scripts to calculate the next version automatically:

**Single message:**

```bash
# Python
python scripts/bump-version.py 1.2.3 'feat(MOB-123): add new command'
# Output: 1.3.0

# PowerShell
.\scripts\bump-version.ps1 -Version 1.2.3 -Message "feat(MOB-123): add new command"
# Output: 1.3.0
```

**Auto mode (reads git log since last tag):**

```bash
# Python
python scripts/bump-version.py --auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)

# PowerShell
.\scripts\bump-version.ps1 -Auto
# Output: 1.3.0
# stderr: v1.2.0 → v1.3.0 (minor bump, 5 commits: minor: 2, patch: 3)
```

> **CI usage:** The version is printed to **stdout**, the summary to **stderr**.
> In a pipeline, capture just the version:
> ```bash
> VERSION=$(python scripts/bump-version.py --auto)
> # or PowerShell: $VERSION = .\scripts\bump-version.ps1 -Auto
> ```

**Pipe multiple messages via stdin:**

```bash
# Python
git log v1.2.3..HEAD --format=%s | python scripts/bump-version.py 1.2.3
# Output: highest bump across all commits

# PowerShell
git log v1.2.3..HEAD --format=%s | .\scripts\bump-version.ps1 -Version 1.2.3
# Output: highest bump across all commits
```

> **Note:** In zsh/bash, use **single quotes** for commit messages containing `!` to prevent
> shell history expansion (e.g. `'feat(MOB-123)!: ...'` instead of `"feat(MOB-123)!: ..."`).


### 2. Create and Push a Tag

```bash
git tag v1.3.0 -m "v1.3.0 - Short description of the release"
git push origin v1.3.0
```

### 3. Verify the Release

The `Release` workflow runs automatically on any `v*` tag push. It builds binaries for all
platforms (Linux, macOS, Windows × amd64, arm64) and publishes them as a GitHub release.

```bash
# Watch the release workflow
gh run list --repo xiabai84/githooks --limit 3

# View the published release
gh release view v1.3.0
```

The release is available at `https://github.com/xiabai84/githooks/releases/tag/v1.3.0`.
Users running `install.sh` or `install.ps1` will automatically download the latest version.

For a complete guide on using githooks in a team workflow with feature branches,
pull requests, and automated releases, see the [Development Guide](docs/DEVELOPMENT-GUIDE.md).

## Acknowledgements

This project is forked from [stefan-niemeyer/githooks](https://github.com/stefan-niemeyer/githooks).
Thank you to Stefan Niemeyer for the original work that laid the foundation for this tool.

## License

This project is released under the [MIT License](LICENSE).
