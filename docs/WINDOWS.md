# Windows Configuration Guide

This guide walks you through setting up githooks on Windows, including installing the CLI,
configuring the commit-msg hook, and handling Windows-specific path conventions.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Step 1: Install Git for Windows](#step-1-install-git-for-windows)
- [Step 2: Install githooks](#step-2-install-githooks)
- [Step 3: Initialize githooks](#step-3-initialize-githooks)
- [Step 4: Add a Workspace](#step-4-add-a-workspace)
- [Step 5: Verify the Setup](#step-5-verify-the-setup)
- [Path Conventions on Windows](#path-conventions-on-windows)
- [Using githooks with IDEs](#using-githooks-with-ides)
  - [Visual Studio Code](#visual-studio-code)
  - [JetBrains IDEs (IntelliJ, Rider, etc.)](#jetbrains-ides-intellij-rider-etc)
  - [Visual Studio](#visual-studio)
- [Using githooks with Windows Terminal](#using-githooks-with-windows-terminal)
- [Troubleshooting Windows-specific Issues](#troubleshooting-windows-specific-issues)
- [Uninstalling](#uninstalling)

## Prerequisites

| Requirement | Minimum Version | How to Check |
|---|---|---|
| Git for Windows | 2.13+ | `git --version` |
| Windows | 10 / Server 2016+ | `winver` |

Git for Windows includes **Git Bash**, which provides the Bash shell required to run the commit-msg hook.
No additional Bash installation is needed.

## Step 1: Install Git for Windows

If Git is not already installed, download it from [gitforwindows.org](https://gitforwindows.org/).

During installation, ensure these options are selected:

- **Use Git from the Windows Command Prompt** (or "Git from the command line and also from 3rd-party software")
- **Use bundled OpenSSH**
- **Checkout as-is, commit Unix-style line endings** (recommended)

Verify the installation:

```powershell
git --version
# git version 2.47.1.windows.1

# Verify Git Bash is available
& "C:\Program Files\Git\bin\bash.exe" --version
# GNU bash, version 5.2.37(1)-release
```

## Step 2: Install githooks

### Option A: PowerShell (recommended)

Open PowerShell and run:

```powershell
# Download and extract the latest release
irm https://raw.githubusercontent.com/xiabai84/githooks/main/install.ps1 | iex

# Move to a directory in your PATH
New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\bin" | Out-Null
Move-Item -Force githooks.exe "$env:USERPROFILE\bin\githooks.exe"
```

Add `%USERPROFILE%\bin` to your PATH if not already done:

```powershell
# Add to PATH permanently (current user)
$binPath = "$env:USERPROFILE\bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$binPath*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$binPath", "User")
    Write-Host "Added $binPath to PATH. Restart your terminal for changes to take effect."
}
```

### Option B: Git Bash

Open Git Bash and run:

```bash
curl -sfL https://raw.githubusercontent.com/xiabai84/githooks/main/install.sh | sh

# Create ~/bin if needed and move the binary there
mkdir -p ~/bin
mv githooks.exe ~/bin/
```

Git Bash automatically includes `~/bin` in the PATH.

### Option C: Manual Installation

1. Go to the [Releases](https://github.com/xiabai84/githooks/releases) page
2. Download `githooks-<version>-windows-amd64.zip`
3. Extract `githooks.exe`
4. Move it to a directory in your PATH (e.g. `C:\Users\<you>\bin\`)

Verify the installation in any terminal:

```powershell
githooks version
```

## Step 3: Initialize githooks

Open **Git Bash** (required for the interactive prompts) and run:

```bash
githooks init
```

This creates:

```
C:\Users\<you>\
├── .gitconfig
└── .githooks\
    ├── commit-msg
    └── config\
        └── githooks.json
```

> **Note:** On Windows, `~` resolves to `C:\Users\<your-username>`.
> Git Bash and Git for Windows handle this translation automatically.

## Step 4: Add a Workspace

Still in Git Bash:

```bash
githooks add
```

When prompted, enter:

1. **Workspace name**: e.g. `myproject`
2. **Jira project key RegEx**: e.g. `MYPROJ` or `(ALPHA|BETA)`
3. **Workspace folder**: use Unix-style paths with `~/`

### Windows Path Format

When entering the workspace folder path, **always use forward slashes and `~/`**:

| Input | Correct? |
|---|---|
| `~/projects/myproject/` | Yes |
| `~/repos/` | Yes |
| `C:\Users\me\projects\` | **No** — use `~/projects/` instead |
| `C:/Users/me/projects/` | **No** — use `~/projects/` instead |

Git's `includeIf` directive on Windows understands `~/` and forward slashes.

### Example Session

```
$ githooks add
? Enter your workspace name: myproject
✔ myproject
? Enter your Jira project key RegEx (MYPROJECT): MYPROJECT
✔ MYPROJECT
? Enter path to your workspace (~/repos/myproject/): ~/repos/myproject/
✔ ~/repos/myproject/
========================== ~/.gitconfig ==========================
[includeIf "gitdir:~/repos/myproject/"]
    path = .githooks/config/gitconfig-myproject
========================== ~/.githooks/config/gitconfig-myproject ==========================
[core]
    hooksPath=~/.githooks
[user]
    jiraProjects=MYPROJECT
? Input was correct (y/N): y
```

## Step 5: Verify the Setup

Navigate to any Git repository under your workspace folder:

```bash
cd ~/repos/myproject/some-repo

# Verify the hook configuration is active
git config --get core.hooksPath
# C:/Users/<you>/.githooks

git config --get user.jiraProjects
# MYPROJECT
```

Test the hook with a commit:

```bash
# This should be accepted
git commit --allow-empty -m "feat(MYPROJECT-1): test commit"

# This should be rejected (no conventional commit format)
git commit --allow-empty -m "test commit"
# ERROR: Commit message must follow Conventional Commits format...

# This should be rejected (no Jira ticket)
git commit --allow-empty -m "feat: test commit"
# ERROR: Commit message must include a Jira ticket...
```

## Path Conventions on Windows

Git for Windows translates paths between Windows and Unix formats. Here's how
githooks paths map:

| githooks config | Windows filesystem |
|---|---|
| `~/.githooks/` | `C:\Users\<you>\.githooks\` |
| `~/.gitconfig` | `C:\Users\<you>\.gitconfig` |
| `~/.githooks/commit-msg` | `C:\Users\<you>\.githooks\commit-msg` |
| `~/repos/myproject/` | `C:\Users\<you>\repos\myproject\` |

The `commit-msg` hook script is a Bash script. Git for Windows automatically invokes
it through Git Bash's bundled `bash.exe`, regardless of which terminal or IDE you use
to run `git commit`. You do not need to configure anything for this — it works
out of the box.

## Using githooks with IDEs

### Visual Studio Code

VS Code uses the Git installation found in your PATH. The commit-msg hook runs
automatically when you commit from the Source Control panel or the integrated terminal.

No additional configuration is needed. To verify:

1. Open the integrated terminal (`` Ctrl+` ``)
2. Run `git config --get core.hooksPath`
3. It should return your `.githooks` path

If VS Code cannot find Git, set the path in settings:

```json
{
    "git.path": "C:\\Program Files\\Git\\cmd\\git.exe"
}
```

### JetBrains IDEs (IntelliJ, Rider, etc.)

JetBrains IDEs respect Git hooks by default. The commit-msg hook runs when you
commit from the IDE's commit dialog.

If hooks are not running, check:
1. **Settings** > **Version Control** > **Git**
2. Ensure **"Run Git hooks"** is checked (enabled by default)
3. Verify the Git executable path points to your Git for Windows installation

### Visual Studio

Visual Studio 2019+ respects Git hooks when using the built-in Git integration.

If hooks are not triggered:
1. Go to **Tools** > **Options** > **Source Control** > **Git Global Settings**
2. Ensure it uses the system Git installation, not a bundled one
3. Verify with `git config --get core.hooksPath` in the Developer PowerShell

## Using githooks with Windows Terminal

You can run `githooks` from any terminal on Windows:

**Windows Terminal (PowerShell):**
```powershell
githooks.exe list
```

**Windows Terminal (Git Bash profile):**
```bash
githooks list
```

**Command Prompt:**
```cmd
githooks.exe list
```

For `githooks init` and `githooks add`, use Git Bash since the interactive
prompts work best in a Unix-compatible terminal.

## Troubleshooting Windows-specific Issues

### "bash: githooks: command not found" in Git Bash

The directory containing `githooks.exe` is not in your PATH.

```bash
# Check where githooks is located
which githooks.exe || find ~/bin -name "githooks.exe" 2>/dev/null

# Add ~/bin to PATH in your ~/.bashrc
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Hook not running — "cannot spawn .githooks/commit-msg"

Git cannot find `bash.exe` to execute the hook script. Fix:

```bash
# Verify bash is available to Git
where bash

# If bash is not found, check the Git for Windows installation
ls "C:/Program Files/Git/bin/bash.exe"
```

This usually means Git was installed without adding it to the system PATH.
Reinstall Git for Windows and select **"Git from the command line and also from 3rd-party software"**.

### "permission denied" when running the hook

Windows does not use Unix file permissions, but Git tracks the executable bit.
Ensure the commit-msg script is marked as executable in Git:

```bash
# Check if the file is executable
ls -la ~/.githooks/commit-msg
# Should show -rwxr-xr-x

# If not executable, fix it:
chmod +x ~/.githooks/commit-msg
```

### Line ending issues (hook fails with cryptic errors)

If the commit-msg hook was checked out with Windows line endings (`\r\n`),
Bash cannot parse it. Fix:

```bash
# Check for Windows line endings
file ~/.githooks/commit-msg
# Should say "Bourne-Again shell script" NOT "with CRLF line terminators"

# Fix line endings
sed -i 's/\r$//' ~/.githooks/commit-msg
```

To prevent this globally, ensure Git is configured to handle line endings:

```bash
git config --global core.autocrlf input
```

You can also add a `.gitattributes` file to force Unix line endings for hook scripts:

```
commit-msg text eol=lf
```

### `includeIf` not matching on Windows

Git on Windows requires forward slashes in `gitdir` paths. If your `.gitconfig`
contains backslashes, the `includeIf` will not match:

```gitconfig
# WRONG — will not match on Windows
[includeIf "gitdir:C:\\Users\\me\\repos\\"]

# CORRECT
[includeIf "gitdir:~/repos/"]

# ALSO CORRECT
[includeIf "gitdir:C:/Users/me/repos/"]
```

Running `githooks add` always generates the correct format using `~/`.

### Antivirus blocking githooks.exe

Some antivirus software may flag `githooks.exe` as unknown. If you get execution
policy or "blocked by antivirus" errors:

1. Check **Windows Security** > **Virus & threat protection** > **Protection history**
2. Allow the file if it was downloaded from the official GitHub releases
3. Or build from source to avoid unsigned binary issues:
   ```powershell
   git clone https://github.com/xiabai84/githooks.git
   cd githooks
   go build -o githooks.exe .
   ```

## Uninstalling

To remove githooks from your Windows system:

**1. Remove the binary:**

```powershell
Remove-Item "$env:USERPROFILE\bin\githooks.exe" -Force
```

**2. Remove the hook directory and config:**

```powershell
Remove-Item "$env:USERPROFILE\.githooks" -Recurse -Force
```

**3. Clean up `.gitconfig`:**

Open `~/.gitconfig` in a text editor and remove all `[includeIf ...]` blocks
that reference `.githooks/config/gitconfig-*`.
