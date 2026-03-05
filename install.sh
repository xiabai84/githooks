#!/bin/sh

set -e

REPO="stefan-niemeyer/githooks"
BINARY="githooks"
OSTYPE=""
ARCH=""
FORMAT=""
DOWNLOAD_URL=""

detect_os() {
  case "$(uname -s)" in
  Linux*)
    OSTYPE="linux"
    FORMAT="tar.gz"
    ;;
  Darwin*)
    OSTYPE="darwin"
    FORMAT="tar.gz"
    ;;
  MINGW* | MSYS* | CYGWIN* | Windows_NT*)
    OSTYPE="windows"
    FORMAT="zip"
    ;;
  *)
    echo "Error: Unsupported operating system '$(uname -s)'."
    echo "Supported: Linux, macOS, Windows (Git Bash / MSYS2 / WSL)"
    exit 1
    ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
  x86_64 | amd64)
    ARCH="amd64"
    ;;
  armv8* | aarch64* | arm64*)
    ARCH="arm64"
    ;;
  *)
    echo "Error: Unsupported architecture '$(uname -m)'."
    echo "Supported: amd64 (x86_64), arm64 (aarch64)"
    exit 1
    ;;
  esac
}

fetch_download_url() {
  DOWNLOAD_URL=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | \
    grep -Ee "browser_download_url.*${OSTYPE}-${ARCH}\.${FORMAT}" | \
    sed -Ee 's/^ *"browser_download_url": *"(.*)"/\1/g')

  if [ -z "$DOWNLOAD_URL" ]; then
    echo "Error: Could not find a release for ${OSTYPE}-${ARCH}.${FORMAT}"
    echo "Check available releases at: https://github.com/${REPO}/releases"
    exit 1
  fi
}

download_and_extract() {
  filename="${DOWNLOAD_URL##*/}"
  echo "Downloading ${BINARY} from ${DOWNLOAD_URL} ..."

  trap 'rm -f "$filename"' EXIT

  if ! curl -fsLO "$DOWNLOAD_URL"; then
    echo ""
    echo "Error: Failed to download ${filename}."
    echo "Please verify the release exists at: https://github.com/${REPO}/releases"
    exit 1
  fi

  echo "Extracting ${filename} ..."

  case "$FORMAT" in
  tar.gz)
    if ! command -v tar >/dev/null 2>&1; then
      echo "Error: 'tar' is required but not found. Please install tar and try again."
      exit 1
    fi
    tar -xzf "${filename}"
    ;;
  zip)
    if command -v unzip >/dev/null 2>&1; then
      unzip -o "${filename}"
    elif command -v powershell >/dev/null 2>&1; then
      powershell -Command "Expand-Archive -Force '${filename}' '.'"
    else
      echo "Error: 'unzip' or 'powershell' is required to extract .zip files."
      exit 1
    fi
    ;;
  esac

  echo ""
  echo "Installation complete!"
  echo ""

  if [ "$OSTYPE" = "windows" ]; then
    echo "Move githooks.exe to a directory in your PATH, for example:"
    echo "  move githooks.exe C:\\Users\\%USERNAME%\\bin\\"
    echo ""
    echo "Note: The commit-msg hook requires Git Bash (included with Git for Windows)."
  else
    echo "Move githooks to a directory in your PATH, for example:"
    echo "  sudo mv githooks /usr/local/bin/"
  fi
}

detect_os
detect_arch
fetch_download_url
download_and_extract
