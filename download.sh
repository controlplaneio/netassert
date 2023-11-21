#!/bin/bash
set -euo pipefail

USER='controlplaneio'
REPO='netassert'
BINARY='netassert'
PWD=$(pwd)
LATEST=$(curl --silent "https://api.github.com/repos/$USER/$REPO/releases/latest" | grep '"tag_name":' | cut -d'"' -f4)
echo "Found latest release: $LATEST"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
echo "OS: $OS"
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
  ARCH="amd64"
fi
echo "ARCH: $ARCH"
FILE="${BINARY}_${LATEST}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/controlplaneio/${REPO}/releases/download/${LATEST}/${FILE}"
CHECKSUM_URL="https://github.com/controlplaneio/${REPO}/releases/download/${LATEST}/checksums-sha256.txt"
echo "[+] Downloading latest checksums from ${CHECKSUM_URL}"
if ! curl -sfLo "checksums.txt" "$CHECKSUM_URL"; then
  echo "Failed to download checksums"
  exit 1
fi
echo "[+] Downloading latest tarball from ${DOWNLOAD_URL}"
if ! curl -sfLO "$DOWNLOAD_URL"; then
  echo "Failed to download tarball"
  exit 1
fi
echo "[+] Verifying checksums"
if ! sha256sum -c checksums.txt --ignore-missing; then
  echo "[+] Checksum verification failed"
  exit 1
fi
echo "[+] Downloaded file verified successfully"
## unzip the tarball
echo "[+] Unzipping the downloaded tarball in directory ${PWD}"
if ! tar -xzf "${FILE}"; then
  echo "[+] Failed to unzip the downloaded tarball"
  exit 1
fi
echo "[+] Downloaded file unzipped successfully"
if [[ ! -f "${BINARY}" ]]; then
  echo "[+] ${BINARY} file was not found in the current path"
  exit 1
fi
echo "[+] You can now run netassert from ${PWD}/${BINARY}"