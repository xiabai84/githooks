#Requires -Version 5.1
<#
.SYNOPSIS
    Install githooks on Windows.
.DESCRIPTION
    Downloads the latest githooks release for Windows and extracts it
    to the current directory.
.EXAMPLE
    irm https://raw.githubusercontent.com/xiabai84/githooks/main/scripts/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

$repo = "xiabai84/githooks"
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else {
    Write-Error "32-bit systems are not supported."
    exit 1
}

Write-Host "Detecting system: windows-$arch" -ForegroundColor Cyan

# Fetch latest release
$releaseUrl = "https://api.github.com/repos/$repo/releases/latest"
Write-Host "Fetching latest release from $releaseUrl ..."

try {
    $release = Invoke-RestMethod -Uri $releaseUrl -Headers @{ "User-Agent" = "githooks-installer" }
} catch {
    Write-Error "Failed to fetch release information: $_"
    exit 1
}

$pattern = "windows-$arch.zip"
$asset = $release.assets | Where-Object { $_.name -like "*$pattern" } | Select-Object -First 1

if (-not $asset) {
    Write-Error "No release found matching '$pattern'. Check: https://github.com/$repo/releases"
    exit 1
}

$downloadUrl = $asset.browser_download_url
$filename = $asset.name

Write-Host "Downloading $filename ..." -ForegroundColor Cyan

$tempFile = Join-Path $env:TEMP $filename
$tempExtract = Join-Path $env:TEMP "githooks-install"
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
} catch {
    Write-Error "Failed to download: $_"
    exit 1
}

Write-Host "Extracting $filename ..." -ForegroundColor Cyan

if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force }
Expand-Archive -Path $tempFile -DestinationPath $tempExtract -Force
Remove-Item $tempFile -Force

$binaryPath = Join-Path $tempExtract "githooks.exe"
if (Test-Path $binaryPath) {
    $destPath = Get-Location
    Copy-Item $binaryPath -Destination $destPath -Force
    Remove-Item $tempExtract -Recurse -Force
    Write-Host ""
    Write-Host "Installation complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "githooks.exe is in: $destPath"
    Write-Host ""
    Write-Host "To use globally, move it to a directory in your PATH:" -ForegroundColor Yellow
    Write-Host "  Move-Item githooks.exe `$env:USERPROFILE\bin\"
    Write-Host ""
    Write-Host "Note: The commit-msg hook requires Git Bash (included with Git for Windows)."
} else {
    Remove-Item $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
    Write-Error "Extraction succeeded but githooks.exe was not found."
    exit 1
}
