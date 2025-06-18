#!/bin/bash

# Claude Company Installation Script
# Builds and installs the ccs binary

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/bin"

echo "Building and installing Claude Company..."

# Build the binary
cd "$SCRIPT_DIR"
go build -o bin/ccs

# Create bin directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Copy binary
cp "$SCRIPT_DIR/bin/ccs" "$INSTALL_DIR/"

# Make it executable
chmod +x "$INSTALL_DIR/ccs"

# Add to PATH in shell configs
add_to_path() {
    local shell_config="$1"
    if [[ -f "$shell_config" ]]; then
        if ! grep -q 'export PATH="$HOME/bin:$PATH"' "$shell_config"; then
            echo 'export PATH="$HOME/bin:$PATH"' >> "$shell_config"
            echo "Added PATH to $shell_config"
        else
            echo "PATH already configured in $shell_config"
        fi
    fi
}

add_to_path "$HOME/.bashrc"
add_to_path "$HOME/.zshrc"
add_to_path "$HOME/.profile"

echo "Installation complete!"
echo "Command available: ccs"
echo ""
echo "Please restart your terminal or run: source ~/.bashrc (or ~/.zshrc)"