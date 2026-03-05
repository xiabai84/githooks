#!/usr/bin/env python3
"""Bump a semantic version based on a Conventional Commits message.

Determines the version bump according to Conventional Commits and
Semantic Versioning:

  Major (X.0.0) - breaking change (! or BREAKING CHANGE footer)
  Minor (0.X.0) - feat
  Patch (0.0.X) - fix, docs, refactor, perf, build, chore, revert
  No bump (0.0.0) - style, test, and ci

Usage:
  python bump-version.py 1.0.0 "feat(ABC-123): add dashboard"
  # Output: 1.1.0

  python bump-version.py 2.3.1 "fix(ABC-456)!: critical auth fix"
  # Output: 3.0.0

  git log -1 --format=%s | python bump-version.py 1.2.3
  # Reads commit message from stdin
"""

import re
import sys

COMMIT_TYPES = r"feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
CONV_RE = re.compile(rf"^({COMMIT_TYPES})(\([^)]*\))?(!)?: .+")

MINOR_TYPES = {"feat"}
PATCH_TYPES = {"fix", "docs", "refactor", "perf", "build", "chore", "revert"}
NO_RELEASE_TYPES = {"style", "test", "ci"}


def bump_version(version: str, message: str) -> str:
    match = CONV_RE.match(message)
    if not match:
        print(f"Error: commit message does not follow Conventional Commits format: {message}", file=sys.stderr)
        sys.exit(1)

    commit_type = match.group(1)
    bang = match.group(3)
    is_breaking = bang == "!" or "BREAKING CHANGE:" in message

    if is_breaking:
        bump = "major"
    elif commit_type in MINOR_TYPES:
        bump = "minor"
    elif commit_type in PATCH_TYPES:
        bump = "patch"
    elif commit_type in NO_RELEASE_TYPES:
        print(version)
        print(f"No version bump for type '{commit_type}'.", file=sys.stderr)
        sys.exit(0)
    else:
        print(f"Error: unknown commit type '{commit_type}'.", file=sys.stderr)
        sys.exit(1)

    major, minor, patch = (int(p) for p in version.split("."))

    if bump == "major":
        major, minor, patch = major + 1, 0, 0
    elif bump == "minor":
        minor, patch = minor + 1, 0
    elif bump == "patch":
        patch += 1

    print(f"{major}.{minor}.{patch}")


def main():
    if len(sys.argv) < 2:
        print("Usage: bump-version.py <version> [message]", file=sys.stderr)
        print('  echo "feat: ..." | bump-version.py 1.0.0', file=sys.stderr)
        sys.exit(1)

    version = sys.argv[1]
    if not re.match(r"^\d+\.\d+\.\d+$", version):
        print(f"Error: invalid version format '{version}'. Expected MAJOR.MINOR.PATCH", file=sys.stderr)
        sys.exit(1)

    if len(sys.argv) >= 3:
        message = sys.argv[2]
    elif not sys.stdin.isatty():
        message = sys.stdin.readline().strip()
    else:
        print("Error: no commit message provided. Pass as argument or pipe via stdin.", file=sys.stderr)
        sys.exit(1)

    if not message:
        print("Error: empty commit message.", file=sys.stderr)
        sys.exit(1)

    bump_version(version, message)


if __name__ == "__main__":
    main()
