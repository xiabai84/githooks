#!/usr/bin/env python3
"""Bump a semantic version based on Conventional Commits.

Determines the version bump according to Conventional Commits and
Semantic Versioning:

  Major (X.0.0) - breaking change (! or BREAKING CHANGE footer)
  Minor (0.X.0) - feat
  Patch (0.0.X) - fix, perf, revert, build
  No bump       - docs, style, refactor, test, ci, chore

Modes:
  Single message:
    python bump-version.py 1.0.0 'feat(ABC-123): add dashboard'
    # Output: 1.1.0

  Auto mode (reads git log since last tag):
    python bump-version.py --auto
    # Detects current version from latest git tag, scans all commits,
    # and outputs the next version to stdout (summary to stderr).
    # CI usage: VERSION=$(python bump-version.py --auto)

  Pipe multiple messages via stdin:
    git log v1.0.0..HEAD --format=%s | python bump-version.py 1.0.0
    # Scans all lines and applies the highest-priority bump.
"""

import re
import subprocess
import sys

COMMIT_TYPES = "feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
CONV_RE = re.compile(r"^(" + COMMIT_TYPES + r")(\([^)]*\))?(!)?: .+")

BUMP_PRIORITY = {"major": 3, "minor": 2, "patch": 1, "none": 0}
MINOR_TYPES = {"feat"}
PATCH_TYPES = {"fix", "perf", "revert", "build"}
NO_RELEASE_TYPES = {"docs", "style", "refactor", "test", "ci", "chore"}


def classify_message(message: str) -> str:
    """Classify a commit message and return the bump level."""
    match = CONV_RE.match(message.strip())
    if not match:
        return "none"

    commit_type = match.group(1)
    bang = match.group(3)
    is_breaking = bang == "!" or "BREAKING CHANGE:" in message

    if is_breaking:
        return "major"
    elif commit_type in MINOR_TYPES:
        return "minor"
    elif commit_type in PATCH_TYPES:
        return "patch"
    return "none"


def highest_bump(messages: list[str]) -> str:
    """Return the highest-priority bump across all messages."""
    max_bump = "none"
    for msg in messages:
        bump = classify_message(msg)
        if BUMP_PRIORITY[bump] > BUMP_PRIORITY[max_bump]:
            max_bump = bump
        if max_bump == "major":
            break  # can't go higher
    return max_bump


def apply_bump(version: str, bump: str) -> str:
    """Apply a bump level to a semantic version string."""
    major, minor, patch = (int(p) for p in version.split("."))

    if bump == "major":
        major, minor, patch = major + 1, 0, 0
    elif bump == "minor":
        minor, patch = minor + 1, 0
    elif bump == "patch":
        patch += 1

    return f"{major}.{minor}.{patch}"


def git_command(*args: str) -> str:
    """Run a git command and return stdout."""
    result = subprocess.run(
        ["git"] + list(args),
        capture_output=True, text=True
    )
    if result.returncode != 0:
        print(f"Error: git {' '.join(args)} failed: {result.stderr.strip()}", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def get_last_tag() -> str:
    """Get the most recent version tag."""
    result = subprocess.run(
        ["git", "describe", "--tags", "--abbrev=0", "--match", "v*"],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        return "v0.0.0"
    return result.stdout.strip()


def auto_bump():
    """Detect version from git tags, scan commits, and print next version."""
    last_tag = get_last_tag()
    version = last_tag.lstrip("v")

    if not re.match(r"^\d+\.\d+\.\d+$", version):
        print(f"Error: invalid tag format '{last_tag}'. Expected vMAJOR.MINOR.PATCH", file=sys.stderr)
        sys.exit(1)

    # If no real tag exists, scan all commits; otherwise scan since last tag
    if last_tag == "v0.0.0":
        log_output = git_command("log", "--format=%s")
    else:
        log_output = git_command("log", f"{last_tag}..HEAD", "--format=%s")

    if not log_output:
        print(version)
        print("No new commits since last tag.", file=sys.stderr)
        sys.exit(0)

    messages = log_output.split("\n")
    bump = highest_bump(messages)

    if bump == "none":
        print(version)
        print(f"No release-worthy commits found ({len(messages)} commits scanned).", file=sys.stderr)
        sys.exit(0)

    new_version = apply_bump(version, bump)
    print(new_version)

    # Summary to stderr
    type_counts: dict[str, int] = {}
    for msg in messages:
        b = classify_message(msg)
        type_counts[b] = type_counts.get(b, 0) + 1
    summary = ", ".join(f"{k}: {v}" for k, v in sorted(type_counts.items()) if k != "none")
    print(f"{last_tag} → v{new_version} ({bump} bump, {len(messages)} commits: {summary})", file=sys.stderr)


def main():
    # Auto mode
    if len(sys.argv) >= 2 and sys.argv[1] == "--auto":
        auto_bump()
        return

    # Manual mode
    if len(sys.argv) < 2:
        print("Usage:", file=sys.stderr)
        print("  bump-version.py <version> [message]     Single commit", file=sys.stderr)
        print("  bump-version.py <version>               Pipe commits via stdin", file=sys.stderr)
        print("  bump-version.py --auto                  Auto-detect from git", file=sys.stderr)
        sys.exit(1)

    version = sys.argv[1]
    if not re.match(r"^\d+\.\d+\.\d+$", version):
        print(f"Error: invalid version format '{version}'. Expected MAJOR.MINOR.PATCH", file=sys.stderr)
        sys.exit(1)

    # Single message as argument
    if len(sys.argv) >= 3:
        messages = [sys.argv[2]]
    # Multiple messages from stdin
    elif not sys.stdin.isatty():
        messages = [line.strip() for line in sys.stdin if line.strip()]
    else:
        print("Error: no commit message provided. Pass as argument or pipe via stdin.", file=sys.stderr)
        sys.exit(1)

    if not messages:
        print("Error: no commit messages provided.", file=sys.stderr)
        sys.exit(1)

    bump = highest_bump(messages)

    if bump == "none":
        print(version)
        print(f"No version bump ({len(messages)} commits scanned).", file=sys.stderr)
        sys.exit(0)

    new_version = apply_bump(version, bump)
    print(new_version)


if __name__ == "__main__":
    main()
