#!/bin/bash

OWNER="chalkan3"
REPO="sloth-runner"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR=$(mktemp -d)

# Function to get the latest release tag
get_latest_release() {
  curl --silent "https://api.github.com/repos/$OWNER/$REPO/releases/latest" | \
    grep '"tag_name":' | \
    sed -E 's/.*"([^"]+)".*/\1/'
}

LATEST_TAG=$(get_latest_release)

if [ -z "$LATEST_TAG" ]; then
  echo "Error: Could not fetch latest release tag."
  exit 1
fi

echo "Latest release tag: $LATEST_TAG"

# Determine OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  "x86_64")
    ARCH="amd64"
    ;;
  "arm64" | "aarch64")
    ARCH="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Construct artifact name
ARTIFACT_NAME="${REPO}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$OWNER/$REPO/releases/download/$LATEST_TAG/$ARTIFACT_NAME"

echo "Downloading from: $DOWNLOAD_URL"

# Download the artifact
curl -L "$DOWNLOAD_URL" -o "$TEMP_DIR/$ARTIFACT_NAME"

if [ $? -ne 0 ]; then
  echo "Error: Failed to download artifact."
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Extract the artifact
tar -xzf "$TEMP_DIR/$ARTIFACT_NAME" -C "$TEMP_DIR"

if [ $? -ne 0 ]; then
  echo "Error: Failed to extract artifact."
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Find the executable (assuming it's named 'sloth-runner' inside the extracted directory)
EXECUTABLE_PATH=$(find "$TEMP_DIR" -name "sloth-runner" -type f)

if [ -z "$EXECUTABLE_PATH" ]; then
  echo "Error: Executable 'sloth-runner' not found in the extracted archive."
  rm -rf "$TEMP_DIR"
  exit 1
fi

# Move the executable to the install directory
sudo mv "$EXECUTABLE_PATH" "$INSTALL_DIR/"

if [ $? -ne 0 ]; then
  echo "Error: Failed to move executable to $INSTALL_DIR. Please ensure you have sudo privileges."
  rm -rf "$TEMP_DIR"
  exit 1
fi

echo "Successfully installed sloth-runner to $INSTALL_DIR/sloth-runner"

# Clean up
rm -rf "$TEMP_DIR"
