package config

var CommitMsg = `#!/usr/bin/env bash

# Conventional Commit types
COMMIT_TYPES="feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"

# List of Jira projects whose keys will be accepted by the hook.
# The installation procedure creates a RegExp of a comma separated list of project keys.
# PROJECTS="(DS|MYJIRA|MARS)"               # Accept issue keys like DS-17 or MARS-6
# PROJECTS="(MYJIRA)"                       # Accept only issue keys that start with MYJIRA, like MYJIRA-1966
# PROJECTS="MYJIRA"                         # Same as "(MYJIRA)"
# PROJECTS=""                               # Accept every issue key matching [[:alpha:]][[:alnum:]]*-[[:digit:]]+

PROJECTS=""

if [[ "$PROJECTS" == "" ]]; then
  # find if and where user.jiraProjects is set with
  # git config --show-origin user.jiraProjects
  PROJECTS=$(git config --get user.jiraProjects)
fi

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

FIRST_LINE=$(head -n 1 "$1")

# Allow merge commits without validation
if [[ "$FIRST_LINE" =~ ^Merge ]]; then
  exit 0
fi

# Validate Conventional Commits format: type(scope)!: description
CONV_RE="^(${COMMIT_TYPES})(\(([^)]*)\))?(!)?: (.+)"
if ! [[ "$FIRST_LINE" =~ $CONV_RE ]]; then
  echo >&2 "ERROR: Commit message must follow Conventional Commits format:"
  echo >&2 "  <type>(<scope>): <description>"
  echo >&2 ""
  echo >&2 "  Allowed types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert"
  echo >&2 "  Example: feat(ABC-123): add user authentication"
  echo >&2 ""
  echo >&2 "  Tip: In zsh, use single quotes for messages containing '!' to avoid history expansion:"
  echo >&2 "    git commit -m 'feat(ABC-123)!: breaking change'"
  exit 1
fi

TYPE="${BASH_REMATCH[1]}"
SCOPE="${BASH_REMATCH[3]}"
BANG="${BASH_REMATCH[4]}"
DESC="${BASH_REMATCH[5]}"

# Parse Jira ticket from branch name
parse_git_branch() {
  local branch
  branch=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --abbrev-ref HEAD 2>/dev/null)
  if [ -z "$branch" ]; then
    return
  fi
  if [ -n "$PROJECTS" ]; then
    echo "$branch" | \
        grep --ignore-case --extended-regexp --only-matching --regexp="\<${PROJECTS}-[[:digit:]]+\>" | \
        tr '[:lower:]' '[:upper:]'
  else
    echo "$branch" | \
        grep --extended-regexp --only-matching --regexp='\<[[:alpha:]][[:alnum:]]*-[[:digit:]]+\>' | \
        tr '[:lower:]' '[:upper:]'
  fi
}

# Extract Jira ticket from commit message first line
parse_ticket_from_message() {
  if [ -n "$PROJECTS" ]; then
    echo "$FIRST_LINE" | \
        grep --ignore-case --extended-regexp --only-matching --regexp="\<${PROJECTS}-[[:digit:]]+\>" | \
        tr '[:lower:]' '[:upper:]'
  else
    echo "$FIRST_LINE" | \
        grep --extended-regexp --only-matching --regexp='\<[[:alpha:]][[:alnum:]]*-[[:digit:]]+\>' | \
        tr '[:lower:]' '[:upper:]'
  fi
}

BRANCH_TICKET=$(parse_git_branch)
MSG_TICKET=$(parse_ticket_from_message)

# If branch has a ticket not in the message, auto-insert it as scope
if [[ -n "$BRANCH_TICKET" && ! "$MSG_TICKET" =~ $BRANCH_TICKET ]]; then
  NEW_FIRST_LINE="${TYPE}(${BRANCH_TICKET})${BANG}: ${DESC}"
  REST=$(tail -n +2 "$1")
  echo "New commit message: ${NEW_FIRST_LINE}"
  if [[ -n "$REST" ]]; then
    printf "%s\n%s" "$NEW_FIRST_LINE" "$REST" > "$1"
  else
    echo "$NEW_FIRST_LINE" > "$1"
  fi
  exit 0
fi

# Ensure a Jira ticket is present
if [[ -z "$MSG_TICKET" ]]; then
  if [ -n "$PROJECTS" ]; then
    echo >&2 "ERROR: Commit message must include a Jira ticket matching '$PROJECTS'."
  else
    echo >&2 "ERROR: Commit message must include a Jira ticket."
  fi
  echo >&2 "  Example: feat(ABC-123): add user authentication"
  echo >&2 ""
  echo >&2 "  Tip: In zsh, use single quotes for messages containing '!' to avoid history expansion:"
  echo >&2 "    git commit -m 'feat(ABC-123)!: breaking change'"
  exit 1
fi
`
