#!/bin/bash

# Claude Company Tools Installation Script
# Installs cca and ccs binaries to user's PATH

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$HOME/bin"

echo "Installing Claude Company Tools..."

# Create bin directory if it doesn't exist
mkdir -p "$INSTALL_DIR"

# Copy binaries
cp "$SCRIPT_DIR/cca" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/ccs" "$INSTALL_DIR/"

# Make them executable
chmod +x "$INSTALL_DIR/cca"
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
echo "Commands available:"
echo "  cca - Claude Company Assistant"
echo "  ccs - Claude Company Shell (tmux session manager)"
echo ""
echo "Please restart your terminal or run: source ~/.bashrc (or ~/.zshrc)"