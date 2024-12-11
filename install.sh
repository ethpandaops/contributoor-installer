#!/bin/sh

TOTAL_STEPS="4"
CONTRIBUTOOR_PATH=${CONTRIBUTOOR_PATH:-"$HOME/.contributoor"}
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
VERSION="latest"

# Print usage
usage() {
    echo "Usage: $0 [-p path] [-v version]"
    echo "  -p: Path to install contributoor (default: $HOME/.contributoor)"
    echo "  -v: Version to install (default: latest)"
    exit 1
}

# Error handling
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[33m'
COLOR_RESET='\033[0m'

# Print a failure message to stderr and exit
fail() {
    MESSAGE=$1
    >&2 echo -e "\n${COLOR_RED}**ERROR**\n$MESSAGE${COLOR_RESET}"
    exit 1
}

# Get CPU architecture
UNAME_VAL=$(uname -m)
ARCH=""
case $UNAME_VAL in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       fail "CPU architecture not supported: $UNAME_VAL" ;;
esac

# Get the platform type
PLATFORM=$(uname -s)
case "$PLATFORM" in
    Linux)  PLATFORM="linux" ;;
    Darwin) PLATFORM="darwin" ;;
    *)      fail "Operating system not supported: $PLATFORM" ;;
esac

while getopts "p:v:h" FLAG; do
    case "$FLAG" in
        p) CONTRIBUTOOR_PATH="$OPTARG" ;;
        v) VERSION="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

# Update bin path after potential flag override
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"

# Construct binary URL based on platform and arch
BINARY_NAME="contributoor-installer-test_${PLATFORM}_"
if [ "$ARCH" = "amd64" ]; then
    BINARY_NAME="${BINARY_NAME}x86_64"
else
    BINARY_NAME="${BINARY_NAME}${ARCH}"
fi

if [ "$VERSION" = "latest" ]; then
    BINARY_URL="https://github.com/ethpandaops/contributoor-installer-test/releases/latest/download/${BINARY_NAME}.tar.gz"
else
    BINARY_URL="https://github.com/ethpandaops/contributoor-installer-test/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"
fi

# Print progress
progress() {
    STEP_NUMBER=$1
    MESSAGE=$2
    echo "Step $STEP_NUMBER of $TOTAL_STEPS: $MESSAGE"
}

# Add to PATH if needed
add_to_path() {
    SHELL_RC=""
    case "$SHELL" in
        */bash)
            if [ -f "$HOME/.bash_profile" ]; then
                SHELL_RC="$HOME/.bash_profile"
            else
                SHELL_RC="$HOME/.bashrc"
            fi
            ;;
        */zsh)  SHELL_RC="$HOME/.zshrc" ;;
        */fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
        *)      SHELL_RC="$HOME/.profile" ;;
    esac

    if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
        if ! echo "$PATH" | tr ":" "\n" | grep -Fxq "$CONTRIBUTOOR_BIN"; then
            if [ "$(basename "$SHELL")" = "fish" ]; then
                echo "fish_add_path $CONTRIBUTOOR_BIN" >> "$SHELL_RC"
            else
                echo "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" >> "$SHELL_RC"
            fi
            echo "Added $CONTRIBUTOOR_BIN to PATH in $SHELL_RC"
            echo "Please restart your shell or run: source $SHELL_RC"
        fi
    fi
    echo "To manually add to PATH, add this line to your shell's RC file:"
    echo "  export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\""
}

case ":$PATH:" in
    *":$CONTRIBUTOOR_BIN:"*) ;; # Already in PATH
    *) add_to_path ;;
esac

# Print progress
progress 1 "Detected platform: $PLATFORM ($ARCH)"

# Create directories
progress 2 "Setting up contributoor directories..."
mkdir -p "$CONTRIBUTOOR_PATH"
mkdir -p "$CONTRIBUTOOR_BIN"

# Download and install binary
progress 3 "Installing contributoor-installer binary..."

# Create a temp file for the archive
TEMP_ARCHIVE=$(mktemp)
trap 'rm -f "$TEMP_ARCHIVE"' EXIT

# Download the archive
if ! curl -L -f "$BINARY_URL" -o "$TEMP_ARCHIVE"; then
    fail "Failed to download binary from $BINARY_URL"
fi

# Check if file is empty or too small
if [ ! -s "$TEMP_ARCHIVE" ]; then
    fail "Downloaded file is empty"
fi

# Try to extract
if ! tar -xzf "$TEMP_ARCHIVE" -C "$CONTRIBUTOOR_BIN"; then
    fail "Failed to extract archive. Archive may be corrupted or in wrong format."
fi

# Check if binary exists and is executable
if [ ! -f "$CONTRIBUTOOR_BIN/contributoor-installer" ]; then
    fail "Binary not found after extraction"
fi

chmod +x "$CONTRIBUTOOR_BIN/contributoor-installer"
if [ ! -x "$CONTRIBUTOOR_BIN/contributoor-installer" ]; then
    fail "Failed to make binary executable"
fi

export PATH="$PATH:$CONTRIBUTOOR_BIN"

progress 3 "Contributoor installer has been installed to $CONTRIBUTOOR_BIN/contributoor-installer"

# Run initial install
progress 4 "Running installation..."
contributoor-installer --config-path "$CONTRIBUTOOR_PATH" install 

