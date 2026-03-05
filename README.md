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
  - [Option D: Single Repository (No CLI)](#option-d-single-repository-no-cli)
- [Getting Started](#getting-started)
- [Managing Workspaces](#managing-workspaces)
  - [Adding a Workspace](#adding-a-workspace)
  - [Listing Workspaces](#listing-workspaces)
  - [Deleting a Workspace](#deleting-a-workspace)
- [How It Works](#how-it-works)
  - [Automatic Detection](#automatic-detection)
  - [Multiple Jira Projects](#multiple-jira-projects)
  - [File Structure](#file-structure)
  - [Git Configuration](#git-configuration)
- [Troubleshooting](#troubleshooting)
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
feat(ABC-123): add OAuth2 login for mobile clients

- Implement token refresh flow
- Add secure token storage

BREAKING CHANGE: /api/auth/login now requires client_id parameter
```

### Allowed Types

| Type | Purpose | Version Bump |
|---|---|---|
| `feat` | New feature for the user | Minor (`0.X.0`) |
| `fix` | Bug fix | Patch (`0.0.X`) |
| `docs` | Documentation changes only | Patch (`0.0.X`) |
| `style` | Formatting, whitespace — no logic change | — |
| `refactor` | Code restructuring without behavior change | Patch (`0.0.X`) |
| `perf` | Performance improvement | Patch (`0.0.X`) |
| `test` | Adding or correcting tests | — |
| `build` | Build system or external dependency changes | Patch (`0.0.X`) |
| `ci` | CI/CD pipeline configuration | — |
| `chore` | Maintenance tasks (dependency updates, configs) | Patch (`0.0.X`) |
| `revert` | Reverts a previous commit | Patch (`0.0.X`) |

Append `!` before the colon or add a `BREAKING CHANGE:` footer to indicate a breaking change.
Any type with `!` or a breaking change footer triggers a **Major** (`X.0.0`) version bump.

```
feat(ABC-123)!: redesign authentication API
```

> **Semantic Versioning summary:** Given `MAJOR.MINOR.PATCH` — a breaking change increments MAJOR, a new feature increments MINOR, and all other releasable changes increment PATCH. Types marked `—` do not trigger a release on their own.

> **Why `docs` and `chore` trigger a patch release:** Documentation is bundled into the release archive alongside the binary, so updates to README or guides produce a different artifact. Likewise, `chore` commits often update dependencies or configuration that can affect the shipped result. In contrast, `style`, `test`, and `ci` changes never alter the distributed binary or bundled files, so they do not warrant a new release.

## Hook Behavior

| Commit Message | Branch | Result |
|---|---|---|
| `add new feature` | any | **Rejected** — missing type prefix |
| `[ABC-123] add feature` | any | **Rejected** — not Conventional Commits format |
| `feat: add feature` | `main` | **Rejected** — missing Jira ticket |
| `feat: add feature` | `feature/ABC-123-login` | **Rewritten** to `feat(ABC-123): add feature` |
| `feat(ABC-123): add feature` | any | **Accepted** |
| `refactor(API-99)!: redesign auth` | any | **Accepted** |
| `Merge branch 'main'` | any | **Accepted** — merge commits skip validation |

When working on a branch that contains a Jira issue key (e.g. `feature/ABC-123-login`),
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
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/install.sh | sh
sudo mv githooks /usr/local/bin/
```

The install script automatically detects your operating system and architecture.

### Option B: Windows Installation

**PowerShell (recommended):**

```powershell
irm https://raw.githubusercontent.com/xiabai84/githooks/main/install.ps1 | iex
```

**Git Bash / MSYS2:**

```bash
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/install.sh | sh
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
  -X github.com/stefan-niemeyer/githooks/buildinfo.version=$(git describe --tags) \
  -X github.com/stefan-niemeyer/githooks/buildinfo.gitCommit=$(git rev-parse HEAD)" \
  -o githooks .
```

On Windows, use `go build -o githooks.exe .` instead.

Verify the installation:

```bash
githooks version
```

### Option D: Single Repository (No CLI)

If you only need the hook in a single repository without installing the CLI:

```bash
# Install in the current repository
./install-jira-git-hook

# Install in multiple repositories at once
./install-jira-git-hook repo1 repo2 repo3

# Restrict to specific Jira projects
./install-jira-git-hook --projects=ALPHA,BETA

# Override existing hooks without prompting
./install-jira-git-hook --yes
```

Full usage:

```
Usage: install-jira-git-hook [ OPTIONS ] [ dir ... ]

Install a Git hook to search for a Jira issue key
in the commit message or branch name.

Options:
  -y, --yes       Override existing commit-msg files
  -p, --projects  Let the hook only accept keys for these Jira projects
                  e.g. --projects=DS,MYJIRA,MARS
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
1. **Workspace name** — a label for this workspace (e.g. `alpha`)
2. **Jira project key** — regex matching allowed ticket prefixes (e.g. `ALPHA` or `(ALPHA|BETA)`)
3. **Workspace folder** — the parent directory containing your repositories (e.g. `~/projects/alpha/`)

### 3. Start Committing

Any Git repository under the configured workspace folder now enforces the commit message rules automatically. No per-repository setup is needed.

```bash
cd ~/projects/alpha/my-repo
git commit -m "feat(ALPHA-42): implement dashboard"   # Accepted
git commit -m "quick fix"                              # Rejected
```

## Managing Workspaces

### Adding a Workspace

```bash
githooks add
```

Interactively creates a new workspace. After confirmation, githooks:
- Appends the workspace to `~/.githooks/config/githooks.json`
- Creates `~/.githooks/config/gitconfig-<name>` with the hooks path and Jira project key
- Appends an `includeIf` directive to `~/.gitconfig`

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

## How It Works

### Automatic Detection

githooks leverages Git's [`includeIf`](https://git-scm.com/docs/git-config#_conditional_includes)
directive to automatically apply hooks based on repository location:

```gitconfig
[includeIf "gitdir:~/projects/alpha/"]
    path = .githooks/config/gitconfig-alpha
```

Every Git repository under `~/projects/alpha/` — including repositories cloned in the future —
automatically inherits the hook configuration. No per-repository setup is required.

Verify the configuration is active for any repository:

```bash
cd ~/projects/alpha/any-repo
git config --get core.hooksPath       # ~/.githooks
git config --get user.jiraProjects    # ALPHA
```

### Multiple Jira Projects

A single workspace can accept tickets from multiple Jira projects using a regex:

```
Jira project key RegEx: (ALPHA|BETA|GAMMA)
```

This accepts commits like `feat(ALPHA-1): ...`, `fix(BETA-42): ...`, or `chore(GAMMA-7): ...`.

Alternatively, create separate workspaces with more specific folder paths.
Git resolves `includeIf` directives using the longest matching path, so a workspace
at `~/projects/alpha/` takes precedence over a workspace at `~/projects/`.

### File Structure

After initializing and adding workspaces **Alpha** and **Beta**, the following
structure is created:

```
~
├── .gitconfig
└── .githooks/
    ├── commit-msg
    └── config/
        ├── gitconfig-alpha
        ├── gitconfig-beta
        └── githooks.json
```

### Git Configuration

**`~/.gitconfig`** — conditional includes based on repository location:

```gitconfig
[includeIf "gitdir:~/work/ws-alpha/"]
    path = .githooks/config/gitconfig-alpha
[includeIf "gitdir:~/work/ws-beta/"]
    path = .githooks/config/gitconfig-beta
```

**`~/.githooks/config/gitconfig-alpha`** — per-workspace settings:

```gitconfig
[core]
    hooksPath=~/.githooks
[user]
    jiraProjects=ALPHA
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
and that the Jira ticket follows the format `PROJECT-NUMBER` (e.g. `ABC-123`).

**Branch ticket not auto-inserted**

The branch name must contain a Jira ticket pattern (e.g. `feature/ABC-123-description`).
The hook extracts tickets matching the configured project keys.

## License

This project is released under the [Unlicense](UNLICENSE).
