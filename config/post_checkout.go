package config

var PostCheckout = `#!/usr/bin/env bash

# post-checkout hook: warn on invalid branch names after branch switch
# This hook runs after git checkout/switch. It only warns (does not block).
# The third argument ($3) is 1 for a branch checkout, 0 for a file checkout.

# Only run on branch checkout, not file checkout
if [[ "$3" != "1" ]]; then
  exit 0
fi

CURRENT_BRANCH=$(git symbolic-ref --short HEAD 2>/dev/null || echo "")

# Skip if no branch (detached HEAD) or exempt branch
if [[ -z "$CURRENT_BRANCH" || "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" || "$CURRENT_BRANCH" == "develop" ]]; then
  exit 0
fi

BRANCH_TYPES="feat|feature|fix|hotfix|chore|release|bugfix|docs|refactor|test|ci"
BRANCH_RE="^(${BRANCH_TYPES})/.+"

if ! [[ "$CURRENT_BRANCH" =~ $BRANCH_RE ]]; then
  echo ""
  echo "⚠ WARNING: Branch name does not follow convention: <type>/<TICKET>-<description>"
  echo "  Allowed types: feat, feature, fix, hotfix, chore, release, bugfix, docs, refactor, test, ci"
  echo "  Example: feat/PROJ-123-add-user-auth"
  echo ""
  echo "  Current branch: $CURRENT_BRANCH"
  echo ""
  exit 0
fi

# Release branches are exempt from ticket requirement
if [[ "$CURRENT_BRANCH" =~ ^release/ ]]; then
  exit 0
fi

# Check for Jira ticket in branch name
PROJECTS=""
if [[ "$PROJECTS" == "" ]]; then
  PROJECTS=$(git config --get user.jiraProjects 2>/dev/null || echo "")
fi

if [ -n "$PROJECTS" ]; then
  BRANCH_TICKET=$(echo "$CURRENT_BRANCH" | grep --ignore-case --extended-regexp --only-matching --regexp="\<${PROJECTS}-[[:digit:]]+\>" | tr '[:lower:]' '[:upper:]')
else
  BRANCH_TICKET=$(echo "$CURRENT_BRANCH" | grep --extended-regexp --only-matching --regexp='\<[[:alpha:]][[:alnum:]]*-[[:digit:]]+\>' | tr '[:lower:]' '[:upper:]')
fi

if [[ -z "$BRANCH_TICKET" ]]; then
  echo ""
  if [ -n "$PROJECTS" ]; then
    echo "⚠ WARNING: Branch name should include a Jira ticket matching '$PROJECTS'."
  else
    echo "⚠ WARNING: Branch name should include a Jira ticket."
  fi
  echo "  Example: feat/PROJ-123-add-user-auth"
  echo ""
  echo "  Current branch: $CURRENT_BRANCH"
  echo ""
fi

exit 0
`
