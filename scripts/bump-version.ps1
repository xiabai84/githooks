#Requires -Version 5.1
<#
.SYNOPSIS
    Bump a semantic version based on Conventional Commits.
.DESCRIPTION
    Determines the version bump according to Conventional Commits and
    Semantic Versioning:

      Major   (X.0.0) - breaking change (! or BREAKING CHANGE footer)
      Minor   (0.X.0) - feat
      Patch   (0.0.X) - fix, perf, revert, build, chore
      No bump         - docs, style, refactor, test, ci

    Modes:
      Single message:
        .\bump-version.ps1 -Version 1.0.0 -Message "feat(MOB-123): add dashboard"
        # Output: 1.1.0

      Auto mode (reads git log since last tag):
        .\bump-version.ps1 -Auto
        # Detects current version from latest git tag, scans all commits,
        # and outputs the next version to stdout (summary to stderr).
        # CI usage: $VERSION = .\bump-version.ps1 -Auto

      Pipe multiple messages via stdin:
        git log v1.0.0..HEAD --format=%s | .\bump-version.ps1 -Version 1.0.0
        # Scans all lines and applies the highest-priority bump.

.PARAMETER Version
    The current semantic version (e.g. 1.0.0). Required unless -Auto is used.
.PARAMETER Message
    The commit message to analyse. Can also be piped via stdin.
.PARAMETER Auto
    Auto-detect version from git tags and scan commits since last tag.
.EXAMPLE
    .\bump-version.ps1 -Version 1.0.0 -Message "feat(MOB-123): add dashboard"
    # Output: 1.1.0
.EXAMPLE
    .\bump-version.ps1 -Version 2.3.1 -Message "fix(PAY-456)!: critical auth fix"
    # Output: 3.0.0
.EXAMPLE
    .\bump-version.ps1 -Auto
    # Output: 1.1.0 (auto-detected from git)
.EXAMPLE
    git log v1.0.0..HEAD --format=%s | .\bump-version.ps1 -Version 1.0.0
    # Output: highest bump across all piped commits
#>

param(
    [Parameter()]
    [ValidatePattern('^\d+\.\d+\.\d+$')]
    [string]$Version,

    [Parameter(ValueFromPipeline = $true)]
    [string]$Message,

    [switch]$Auto
)

begin {
    $commitTypes = "feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"
    $pattern = "^($commitTypes)(\([^)]*\))?(!)?: .+"

    $minorTypes = @("feat")
    $patchTypes = @("fix", "perf", "revert", "build", "chore")

    $bumpPriority = @{ "none" = 0; "patch" = 1; "minor" = 2; "major" = 3 }
    $maxBump = "none"
    $messageCount = 0
    $typeCounts = @{}

    function Classify-Message([string]$msg) {
        $trimmed = $msg.Trim()
        if ($trimmed -notmatch $pattern) {
            return "none"
        }
        $type = $Matches[1]
        $bang = $Matches[3]
        $isBreaking = ($bang -eq "!") -or ($trimmed -match "BREAKING CHANGE:")

        if ($isBreaking) { return "major" }
        elseif ($type -in $minorTypes) { return "minor" }
        elseif ($type -in $patchTypes) { return "patch" }
        return "none"
    }

    function Apply-Bump([string]$ver, [string]$bump) {
        $parts = $ver -split '\.'
        $major = [int]$parts[0]
        $minor = [int]$parts[1]
        $patch = [int]$parts[2]

        switch ($bump) {
            "major" { $major++; $minor = 0; $patch = 0 }
            "minor" { $minor++; $patch = 0 }
            "patch" { $patch++ }
        }
        return "$major.$minor.$patch"
    }

    function Get-LastTag {
        try {
            $result = git describe --tags --abbrev=0 --match "v*" 2>$null
            if ($LASTEXITCODE -ne 0 -or -not $result) {
                return "v0.0.0"
            }
            return $result.Trim()
        } catch {
            return "v0.0.0"
        }
    }

    # Handle -Auto mode
    if ($Auto) {
        $lastTag = Get-LastTag
        $Version = $lastTag.TrimStart("v")

        if ($Version -notmatch '^\d+\.\d+\.\d+$') {
            Write-Error "Invalid tag format '$lastTag'. Expected vMAJOR.MINOR.PATCH"
            exit 1
        }

        # If no real tag exists, scan all commits; otherwise scan since last tag
        try {
            if ($lastTag -eq "v0.0.0") {
                $logOutput = git log --format=%s 2>$null
            } else {
                $logOutput = git log "$lastTag..HEAD" --format=%s 2>$null
            }
        } catch {
            $logOutput = $null
        }

        if (-not $logOutput) {
            Write-Host $Version
            Write-Warning "No new commits since last tag."
            exit 0
        }

        foreach ($line in $logOutput) {
            if (-not $line.Trim()) { continue }
            $messageCount++
            $bump = Classify-Message $line
            $typeCounts[$bump] = ($typeCounts[$bump] ?? 0) + 1
            if ($bumpPriority[$bump] -gt $bumpPriority[$maxBump]) {
                $maxBump = $bump
            }
        }

        if ($maxBump -eq "none") {
            Write-Host $Version
            Write-Warning "No release-worthy commits found ($messageCount commits scanned)."
            exit 0
        }

        $newVersion = Apply-Bump $Version $maxBump
        Write-Host $newVersion

        $summary = ($typeCounts.GetEnumerator() | Where-Object { $_.Key -ne "none" } | Sort-Object Key | ForEach-Object { "$($_.Key): $($_.Value)" }) -join ", "
        Write-Warning "$lastTag -> v$newVersion ($maxBump bump, $messageCount commits: $summary)"
        exit 0
    }

    # Validate that Version is provided for non-Auto modes
    if (-not $Version) {
        Write-Host "Usage:" -ForegroundColor Yellow
        Write-Host "  .\bump-version.ps1 -Version <version> -Message <message>   Single commit"
        Write-Host "  .\bump-version.ps1 -Version <version>                      Pipe commits via stdin"
        Write-Host "  .\bump-version.ps1 -Auto                                   Auto-detect from git"
        exit 1
    }

    $pipelineMessages = [System.Collections.ArrayList]@()
}

process {
    if ($Auto) { return }

    if ($Message) {
        [void]$pipelineMessages.Add($Message)
    }
}

end {
    if ($Auto) { return }

    if ($pipelineMessages.Count -eq 0) {
        Write-Error "No commit message provided. Pass -Message or pipe via stdin."
        exit 1
    }

    foreach ($msg in $pipelineMessages) {
        if (-not $msg.Trim()) { continue }
        $messageCount++
        $bump = Classify-Message $msg
        if ($bumpPriority[$bump] -gt $bumpPriority[$maxBump]) {
            $maxBump = $bump
        }
        if ($maxBump -eq "major") { break }
    }

    if ($maxBump -eq "none") {
        Write-Host $Version
        Write-Warning "No version bump ($messageCount commits scanned)."
        exit 0
    }

    $newVersion = Apply-Bump $Version $maxBump
    Write-Host $newVersion
}
