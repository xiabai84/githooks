#Requires -Version 5.1
<#
.SYNOPSIS
    Bump a semantic version based on a Conventional Commits message.
.DESCRIPTION
    Reads a commit message and determines the version bump according to
    Conventional Commits and Semantic Versioning:

      Major   (X.0.0) - breaking change (! or BREAKING CHANGE footer)
      Minor   (0.X.0) - feat
      Patch   (0.0.X) - fix, docs, refactor, perf, build, chore, revert
      No bump (0.0.0) - style, test, and ci

.PARAMETER Version
    The current semantic version (e.g. 1.0.0).
.PARAMETER Message
    The commit message to analyse. Can also be piped via stdin.
.EXAMPLE
    .\bump-version.ps1 -Version 1.0.0 -Message "feat(ABC-123): add dashboard"
    # Output: 1.1.0
.EXAMPLE
    .\bump-version.ps1 -Version 2.3.1 -Message "fix(ABC-456)!: critical auth fix"
    # Output: 3.0.0
.EXAMPLE
    .\bump-version.ps1 -Version 1.2.3 -Message "fix(ABC-789): null pointer fix"
    # Output: 1.2.4
.EXAMPLE
    .\bump-version.ps1 -Version 1.0.0 -Message "test(ABC-1): add unit tests"
    # Output: 1.0.0 (no bump)
#>

param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^\d+\.\d+\.\d+$')]
    [string]$Version,

    [Parameter(ValueFromPipeline = $true)]
    [string]$Message
)

process {
    if (-not $Message) {
        Write-Error "No commit message provided. Pass -Message or pipe from stdin."
        exit 1
    }

    $commitTypes = "feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
    $pattern = "^($commitTypes)(\([^)]*\))?(!)?: .+"

    if ($Message -notmatch $pattern) {
        Write-Error "Commit message does not follow Conventional Commits format: $Message"
        exit 1
    }

    $type = $Matches[1]
    $bang = $Matches[3]

    # Check for breaking change: ! marker or BREAKING CHANGE footer
    $isBreaking = ($bang -eq "!") -or ($Message -match "BREAKING CHANGE:")

    # Determine bump level
    $bump = $null

    if ($isBreaking) {
        $bump = "major"
    }
    elseif ($type -eq "feat") {
        $bump = "minor"
    }
    elseif ($type -in @("fix", "docs", "refactor", "perf", "build", "chore", "revert")) {
        $bump = "patch"
    }
    else {
        # style, test, ci - no release
        Write-Host $Version
        Write-Host "No version bump for type '$type'." -ForegroundColor Yellow
        exit 0
    }

    # Parse and bump version
    $parts = $Version -split '\.'
    $major = [int]$parts[0]
    $minor = [int]$parts[1]
    $patch = [int]$parts[2]

    switch ($bump) {
        "major" { $major++; $minor = 0; $patch = 0 }
        "minor" { $minor++; $patch = 0 }
        "patch" { $patch++ }
    }

    $newVersion = "$major.$minor.$patch"

    Write-Host $newVersion
}
