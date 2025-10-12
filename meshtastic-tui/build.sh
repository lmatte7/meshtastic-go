#!/bin/bash

# Build script for Meshtastic TUI

set -e

echo "Building Meshtastic TUI..."

# Get dependencies
echo "Downloading dependencies..."
go mod download

# Build the binary
echo "Compiling..."
go build -o meshtastic-tui

echo ""
echo "✓ Build successful!"
echo ""
echo "Run the application with:"
echo "  ./meshtastic-tui"
echo ""

