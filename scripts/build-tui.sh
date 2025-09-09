#!/bin/bash
set -e

VERSION="v1.0.0" # Change this when releasing new versions
OUTPUT_DIR="releases/$VERSION"

echo "Building version $VERSION..."

# Create output directory
mkdir -p "$OUTPUT_DIR"

# List of target platforms
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for PLATFORM in "${PLATFORMS[@]}"; do
  OS=$(echo $PLATFORM | cut -d '/' -f1)
  ARCH=$(echo $PLATFORM | cut -d '/' -f2)
  echo "Building for $OS/$ARCH..."

  # Set output filename
  if [ "$OS" = "windows" ]; then
    OUTPUT_NAME="tui-${OS}-${ARCH}.exe"
  else
    OUTPUT_NAME="tui-${OS}-${ARCH}"
  fi

  # Build the binary
  GOOS=$OS GOARCH=$ARCH go build -o "$OUTPUT_DIR/$OUTPUT_NAME" ./services/tui/cmd/tui/main.go

  # Compress the file
  if [ "$OS" = "windows" ]; then
    zip -j "$OUTPUT_DIR/tui-${OS}-${ARCH}.zip" "$OUTPUT_DIR/$OUTPUT_NAME"
    rm "$OUTPUT_DIR/$OUTPUT_NAME"
  else
    tar -czf "$OUTPUT_DIR/tui-${OS}-${ARCH}.tar.gz" -C "$OUTPUT_DIR" "$OUTPUT_NAME"
    rm "$OUTPUT_DIR/$OUTPUT_NAME"
  fi
done

# Generate checksums
echo "Generating checksums..."
cd "$OUTPUT_DIR"
sha256sum * >checksums.txt

echo "Build complete. Files are in $OUTPUT_DIR"
