#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# The main path for your host-agent source code
AGENT_SRC_PATH="../services/host-agent/cmd/agent/main.go"
# The version of this release
VERSION="1.0.0"
# The output directory for all release files
RELEASE_DIR="../releases/host-agent/v$VERSION"

echo "--- Starting Lighthouse Host Agent Build v$VERSION ---"

# --- 1. Clean up previous builds ---
echo "Cleaning up old release directory.."
rm -rf $RELEASE_DIR
mkdir -p $RELEASE_DIR

# --- 2. Build for Windows ---
echo "Building for Windows (amd64)..."
WIN_DIR="$RELEASE_DIR/windows-amd64"
mkdir -p $WIN_DIR
GOOS=windows GOARCH=amd64 go build -o "$WIN_DIR/host-agent.exe" $AGENT_SRC_PATH

# --- 3. Build for Linux ---
echo "Building for Linux (amd64)..."
LINUX_DIR="$RELEASE_DIR/linux-amd64"
mkdir -p $LINUX_DIR
GOOS=linux GOARCH=amd64 go build -o "$LINUX_DIR/host-agent-linux" $AGENT_SRC_PATH

# --- 4. Build for macOS ---
echo "Building for macOS (amd64)..."
MACOS_DIR="$RELEASE_DIR/macos-amd64"
mkdir -p $MACOS_DIR
GOOS=darwin GOARCH=amd64 go build -o "$MACOS_DIR/host-agent-macos" $AGENT_SRC_PATH

echo "All binaries compiled successfully."

# --- 5. Prepare Installer Scripts and READMEs ---
echo "Copying installer scripts and creating READMEs..."

# For Windows
cp ./install-agent.bat "$WIN_DIR/"
echo "To install, right-click 'install.bat' and select 'Run as administrator'." >"$WIN_DIR/README.txt"

# For Linux
cp ./install-agent.sh "$LINUX_DIR/"
echo "To install, run the following command in your terminal: sudo ./install.sh" >"$LINUX_DIR/README.txt"

# For macOS
cp ./install-agent.sh "$MACOS_DIR/"
echo "To install, run the following command in your terminal: sudo ./install.sh" >"$MACOS_DIR/README.txt"
# Note: The Linux and macOS installers can be the same script.

# --- 6. Create Final Archives ---
echo "Creating final release archives..."

# Windows (.zip)
(cd $WIN_DIR/.. && zip -r "lighthouse-agent-$VERSION-windows-amd64.zip" "windows-amd64")

# Linux (.tar.gz)
(cd $LINUX_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-linux-amd64.tar.gz" "linux-amd64")

# macOS (.tar.gz)
(cd $MACOS_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-macos-amd64.tar.gz" "macos-amd64")

# Clean up the intermediate directories
rm -rf $WIN_DIR $LINUX_DIR $MACOS_DIR

echo ""
echo "✅ Build complete! Release files are in the '$RELEASE_DIR' directory."
