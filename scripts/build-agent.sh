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

echo "--- Starting Lighthouse Host Agent Multi-Arch Build v$VERSION ---"

# --- 1. Clean up previous builds ---
echo "Cleaning up old release directory..."
rm -rf $RELEASE_DIR
mkdir -p $RELEASE_DIR

# --- 2. Build for Windows ---
echo "Building for Windows (amd64)..."
WIN_AMD64_DIR="$RELEASE_DIR/windows-amd64"
mkdir -p $WIN_AMD64_DIR
GOOS=windows GOARCH=amd64 go build -o "$WIN_AMD64_DIR/host-agent.exe" $AGENT_SRC_PATH

echo "Building for Windows (arm64)..."
WIN_ARM64_DIR="$RELEASE_DIR/windows-arm64"
mkdir -p $WIN_ARM64_DIR
GOOS=windows GOARCH=arm64 go build -o "$WIN_ARM64_DIR/host-agent.exe" $AGENT_SRC_PATH

# --- 3. Build for Linux ---
echo "Building for Linux (amd64)..."
LINUX_AMD64_DIR="$RELEASE_DIR/linux-amd64"
mkdir -p $LINUX_AMD64_DIR
GOOS=linux GOARCH=amd64 go build -o "$LINUX_AMD64_DIR/host-agent-linux" $AGENT_SRC_PATH

echo "Building for Linux (arm64)..."
LINUX_ARM64_DIR="$RELEASE_DIR/linux-arm64"
mkdir -p $LINUX_ARM64_DIR
GOOS=linux GOARCH=arm64 go build -o "$LINUX_ARM64_DIR/host-agent-linux" $AGENT_SRC_PATH

# --- 4. Build for macOS ---
echo "Building for macOS (amd64)..."
MACOS_AMD64_DIR="$RELEASE_DIR/macos-amd64"
mkdir -p $MACOS_AMD64_DIR
GOOS=darwin GOARCH=amd64 go build -o "$MACOS_AMD64_DIR/host-agent-macos" $AGENT_SRC_PATH

echo "Building for macOS (arm64)..."
MACOS_ARM64_DIR="$RELEASE_DIR/macos-arm64"
mkdir -p $MACOS_ARM64_DIR
GOOS=darwin GOARCH=arm64 go build -o "$MACOS_ARM64_DIR/host-agent-macos" $AGENT_SRC_PATH

echo "All binaries compiled successfully."

# --- 5. Prepare Installer Scripts and READMEs ---
echo "Copying installer scripts and creating READMEs..."

# For Windows
cp ./install-agent.bat "$WIN_AMD64_DIR/"
cp ./install-agent.bat "$WIN_ARM64_DIR/"
echo "To install, right-click 'install-agent.bat' and select 'Run as administrator'." >"$WIN_AMD64_DIR/README.txt"
echo "To install, right-click 'install-agent.bat' and select 'Run as administrator'." >"$WIN_ARM64_DIR/README.txt"

# For Linux
cp ./install-agent.sh "$LINUX_AMD64_DIR/"
cp ./install-agent.sh "$LINUX_ARM64_DIR/"
echo "To install, run the following command in your terminal: sudo ./install-agent.sh" >"$LINUX_AMD64_DIR/README.txt"
echo "To install, run the following command in your terminal: sudo ./install-agent.sh" >"$LINUX_ARM64_DIR/README.txt"

# For macOS
cp ./install-agent.sh "$MACOS_AMD64_DIR/"
cp ./install-agent.sh "$MACOS_ARM64_DIR/"
echo "To install, run the following command in your terminal: sudo ./install-agent.sh" >"$MACOS_AMD64_DIR/README.txt"
echo "To install, run the following command in your terminal: sudo ./install-agent.sh" >"$MACOS_ARM64_DIR/README.txt"

# --- 6. Create Final Archives ---
echo "Creating final release archives..."

# Windows
(cd $WIN_AMD64_DIR/.. && zip -r "lighthouse-agent-$VERSION-windows-amd64.zip" "windows-amd64")
(cd $WIN_ARM64_DIR/.. && zip -r "lighthouse-agent-$VERSION-windows-arm64.zip" "windows-arm64")

# Linux
(cd $LINUX_AMD64_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-linux-amd64.tar.gz" "linux-amd64")
(cd $LINUX_ARM64_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-linux-arm64.tar.gz" "linux-arm64")

# macOS
(cd $MACOS_AMD64_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-macos-amd64.tar.gz" "macos-amd64")
(cd $MACOS_ARM64_DIR/.. && tar -czvf "lighthouse-agent-$VERSION-macos-arm64.tar.gz" "macos-arm64")

# --- 7. Clean up ---
echo "Cleaning up intermediate directories..."
rm -rf $WIN_AMD64_DIR $WIN_ARM64_DIR $LINUX_AMD64_DIR $LINUX_ARM64_DIR $MACOS_AMD64_DIR $MACOS_ARM64_DIR

echo ""
echo "✅ Build complete! Release files are in the '$RELEASE_DIR' directory."
