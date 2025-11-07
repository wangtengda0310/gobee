#!/usr/bin/env bash
# Downloads the Dolt binary into testdata/mysql1/ for local testing.
#
# dolt.exe is intentionally excluded from version control because it exceeds
# GitHub's 100 MB file size limit. Run this script once before running the
# dumper tests that depend on it.
#
# Usage:
#   ./download_dolt.sh              # download the pinned version
#   DOLT_VERSION=2.2.3 ./download_dolt.sh   # download a specific version
#
# Prerequisites: curl and (unzip OR tar). On Windows, tar handles .zip files.

set -euo pipefail

DOLT_VERSION="${DOLT_VERSION:-1.58.0}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_DIR="${SCRIPT_DIR}/mysql1"
TARGET="${TARGET_DIR}/dolt.exe"

OS="$(uname -s)"
ARCH="$(uname -m)"
case "${OS}" in
  MINGW*|MSYS*|CYGWIN*) PLATFORM="windows-amd64"; EXT="zip" ;;
  Darwin)
    case "${ARCH}" in
      arm64|aarch64) PLATFORM="darwin-arm64" ;;
      *) PLATFORM="darwin-amd64" ;;
    esac
    EXT="tar.gz"
    ;;
  Linux)
    case "${ARCH}" in
      arm64|aarch64) PLATFORM="linux-arm64" ;;
      *) PLATFORM="linux-amd64" ;;
    esac
    EXT="tar.gz"
    ;;
  *) echo "Unsupported OS: ${OS}" >&2; exit 1 ;;
esac

FILENAME="dolt-${PLATFORM}.${EXT}"
URL="https://github.com/dolthub/dolt/releases/download/v${DOLT_VERSION}/${FILENAME}"

mkdir -p "${TARGET_DIR}"

if [ -x "${TARGET}" ]; then
  echo "dolt already exists at ${TARGET} (remove it to re-download)."
  exit 0
fi

echo "Downloading Dolt v${DOLT_VERSION} -> ${URL}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT
ARCHIVE="${TMP_DIR}/${FILENAME}"

curl -fL --retry 3 -o "${ARCHIVE}" "${URL}"

echo "Extracting..."
case "${EXT}" in
  zip)
    if command -v unzip >/dev/null 2>&1; then
      unzip -o "${ARCHIVE}" -d "${TMP_DIR}" >/dev/null
    else
      tar -xf "${ARCHIVE}" -C "${TMP_DIR}"
    fi
    ;;
  tar.gz|tgz)
    tar -xzf "${ARCHIVE}" -C "${TMP_DIR}"
    ;;
esac

# The archive layout is dolt-<platform>/bin/dolt[.exe]
BIN_PATH="$(find "${TMP_DIR}" -type f \( -name 'dolt' -o -name 'dolt.exe' \) -path '*/bin/*' | head -n1)"
if [ -z "${BIN_PATH}" ]; then
  echo "Could not locate dolt binary in the archive." >&2
  exit 1
fi

cp -f "${BIN_PATH}" "${TARGET}"
chmod +x "${TARGET}" 2>/dev/null || true

echo "Done: ${TARGET}"
"${TARGET}" version
