#!/bin/bash

# Colors
COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[33m'
COLOR_CYAN='\033[0;36m'
COLOR_RESET='\033[0m'
COLOR_BOLD='\033[1m'

# Constants
TOTAL_STEPS="8"
CONTRIBUTOOR_PATH=${CONTRIBUTOOR_PATH:-"$HOME/.contributoor"}
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
VERSION="latest"

# ASCII Art Logo
print_logo() {
    printf "${COLOR_CYAN}"
    cat << "EOF"

  ______            __       _ __            __              
 / ____/___  ____  / /______(_) /_  __  ____/_/_____  _____ 
/ /   / __ \/ __ \/ __/ ___/ / __ \/ / / / __ \/ __ \/ ___/
+/ /___/ /_/ / / / / /_/ /  / / /_/ / /_/ / /_/ / /_/ / /    
+\____/\____/_/ /_/\__/_/  /_/_.___/\__,_/\____/\____/_/     
                                                             
EOF
    printf "${COLOR_RESET}\n"
    printf "${COLOR_BOLD}Ethereum Distributed Data Collection${COLOR_RESET}\n\n"
}

# Spinner for loading states
spinner() {
    local pid=$1
    local delay=0.1
    local spinstr='|/-\'
    while ps -p $pid > /dev/null; do
        local temp=${spinstr#?}
        printf " [%c]  " "$spinstr"
        local spinstr=$temp${spinstr%"$temp"}
        tput cub 6
        sleep $delay
    done
    printf "    \n"
}

# Print usage
usage() {
    echo "Usage: $0 [-p path] [-v version]"
    echo "  -p: Path to install contributoor (default: $HOME/.contributoor)"
    echo "  -v: Version of contributoor to install without 'v' prefix (default: latest, example: 0.0.6)"
    exit 1
}

# Print a failure message to stderr and exit
fail() {
    MESSAGE=$1
    printf "\n${COLOR_RED}**ERROR**\n%s${COLOR_RESET}\n" "$MESSAGE" >&2
    exit 1
}

# Print progress
progress() {
    STEP_NUMBER=$1
    MESSAGE=$2
    printf "\n\n${COLOR_BLUE}Step ${STEP_NUMBER} of ${TOTAL_STEPS}${COLOR_RESET}: ${COLOR_BOLD}${MESSAGE}${COLOR_RESET}"
}

# Print success message
success() {
    MESSAGE=$1
    printf "\n${COLOR_GREEN}✓ %s${COLOR_RESET}" "$MESSAGE"
}

# Clear the screen and show logo
clear
print_logo

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

# Parse any arguments
while getopts "p:v:h" FLAG; do
    case "$FLAG" in
        p) CONTRIBUTOOR_PATH="$OPTARG" ;;
        v) VERSION="$OPTARG" ;;
        h) usage ;;
        *) usage ;;
    esac
done

# Construct binary URL based on platform and arch
INSTALLER_BINARY_NAME="contributoor-installer-test_${PLATFORM}_"
if [ "$ARCH" = "amd64" ]; then
    INSTALLER_BINARY_NAME="${INSTALLER_BINARY_NAME}x86_64"
else
    INSTALLER_BINARY_NAME="${INSTALLER_BINARY_NAME}${ARCH}"
fi

# Update bin path after potential flag override
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"

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
        *)      SHELL_RC="$HOME/.profile" ;;
    esac

    if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
        PATH_LINE="export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\""
        if ! grep -Fxq "$PATH_LINE" "$SHELL_RC"; then
            # This is the best we can do. When running via `curl ... | sh` any ENV changes
            # only effect the current shell session, so we're unable to `source $SHELL_RC` for them.
            echo "$PATH_LINE" >> "$SHELL_RC"
            echo "Added $CONTRIBUTOOR_BIN to PATH in $SHELL_RC"
            echo "NOTE: You'll need to run 'source $SHELL_RC' or start a new terminal for the PATH changes to take effect"
        fi
    fi

    # Add to PATH for the rest of this script. 
    export PATH="$PATH:$CONTRIBUTOOR_BIN"
}

case ":$PATH:" in
    *":$CONTRIBUTOOR_BIN:"*) ;; # Already in PATH
    *) add_to_path ;;
esac

# Get installation path first
clear
print_logo
progress 1 "Detecting platform"
success "$PLATFORM ($ARCH)"

# Determine version
if [ "$VERSION" = "latest" ]; then
    progress 2 "Determining latest version"
    
    # Get latest release info from GitHub API
    RELEASE_INFO=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor-test/releases/latest")
    if [ $? -ne 0 ]; then
        fail "Failed to check for latest version.\nPlease check your internet connection and try again."
    fi

    # Extract version from release info (removes 'v' prefix if present)
    VERSION=$(echo "$RELEASE_INFO" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//')
    if [ -z "$VERSION" ]; then
        fail "Failed to determine latest version.\nPlease specify a version manually with -v."
    fi
    success "Latest version: $VERSION"
else
    progress 2 "Validating version"
    # Get all releases to validate version
    RELEASES=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor-test/releases")
    if [ $? -ne 0 ]; then
        fail "Failed to check available versions.\nPlease check your internet connection and try again."
    fi

    # Check if version exists (with or without v prefix)
    if ! echo "$RELEASES" | grep -q "\"tag_name\": *\"v\{0,1\}${VERSION}\""; then
        # Get available versions for error message
        AVAILABLE_VERSIONS=$(echo "$RELEASES" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//' | head -n 5 | tr '\n' ', ' | sed 's/,$//')
        success "Last 5 available versions: ${AVAILABLE_VERSIONS}"
        fail "Provided version ${VERSION} not found."
    fi
    success "Using specified version: $VERSION"
fi

progress 3 "Installation path"
printf "\nWhere would you like to install contributoor? [${COLOR_CYAN}~/.contributoor${COLOR_RESET}]: "
read -r CUSTOM_PATH

if [ -n "$CUSTOM_PATH" ]; then
    CONTRIBUTOOR_PATH="$CUSTOM_PATH"
    # Update bin path after potential path change
    CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
fi
success "Using path: $CONTRIBUTOOR_PATH"

# Construct binary URLs.
INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer-test/releases/latest/download/${INSTALLER_BINARY_NAME}.tar.gz"
CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor-test/releases/download/v${VERSION}/contributoor-test_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"

# Handle menu selection
selected=1
trap 'tput cnorm' EXIT  # Ensure cursor is restored on exit

# Hide cursor
tput civis
while true; do
    clear
    print_logo
    progress 1 "Detecting platform..."
    success "$PLATFORM ($ARCH)"
    progress 2 "Determining latest version"
    success "Using version: $VERSION"
    progress 3 "Installation path"
    success "Using path: $CONTRIBUTOOR_PATH"
    progress 4 "Select installation mode"
    printf "\n  %s Docker (${COLOR_CYAN}recommended${COLOR_RESET})\n" "$([ "$selected" = 1 ] && echo ">" || echo " ")"
    printf "  %s Binary\n" "$([ "$selected" = 2 ] && echo ">" || echo " ")"
    printf "\nUse arrow keys (↑/↓) or j/k to select, Enter to confirm\n"
    
    # Read a single character
    read -r -n1 key
    
    case "$key" in
        A|k)  # Up arrow or k
            [ "$selected" -gt 1 ] && selected=$((selected - 1))
            ;;
        B|j)  # Down arrow or j
            [ "$selected" -lt 2 ] && selected=$((selected + 1))
            ;;
        "")  # Enter key
            tput cnorm  # Show cursor again
            printf "Selected: "
            if [ "$selected" = 1 ]; then
                INSTALL_MODE="docker"
                printf "${COLOR_GREEN}Docker${COLOR_RESET}"
            else
                INSTALL_MODE="binary"
                printf "${COLOR_GREEN}Binary${COLOR_RESET}"
            fi
            break
            ;;
    esac
done

# Create directories
progress 5 "Setting up directories"
if ! mkdir -p "$CONTRIBUTOOR_PATH"; then
    fail "Could not create the contributoor user data directory"
fi
success "data directory: $CONTRIBUTOOR_PATH" 

if ! mkdir -p "$CONTRIBUTOOR_BIN"; then
    fail "Could not create the contributoor bin directory"
fi
success "bin directory: $CONTRIBUTOOR_BIN" 

# Download and install binary
progress 6 "Preparing installation"

# Function to do all installation steps
install_installer_binary() {
    # Create a temp file for the archive
    TEMP_ARCHIVE=$(mktemp)

    # Download and verify checksums
    CHECKSUMS_URL="https://github.com/ethpandaops/contributoor-installer-test/releases/latest/download/checksums.txt"
    CHECKSUMS_FILE=$(mktemp)
    
    curl -L -f -s "$CHECKSUMS_URL" -o "$CHECKSUMS_FILE" &
    wait $!
    if [ ! -f "$CHECKSUMS_FILE" ] || [ ! -s "$CHECKSUMS_FILE" ]; then
        rm -f "$CHECKSUMS_FILE"
        fail "Failed to download checksums file"
    fi
    
    # Download and verify
    curl -L -f -s "$INSTALLER_URL" -o "$TEMP_ARCHIVE" &
    spinner $!
    wait $!
    if [ ! -f "$TEMP_ARCHIVE" ] || [ ! -s "$TEMP_ARCHIVE" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Failed to download installer binary"
    fi
    success "Downloaded installer"
    
    # Verify checksum
    BINARY_NAME="${INSTALLER_BINARY_NAME}.tar.gz"
    EXPECTED_CHECKSUM=$(grep "$BINARY_NAME" "$CHECKSUMS_FILE" | cut -d' ' -f1)
    if [ -z "$EXPECTED_CHECKSUM" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Checksum not found for $BINARY_NAME"
    fi

    ACTUAL_CHECKSUM=$(sha256sum "$TEMP_ARCHIVE" | cut -d' ' -f1)
    if [ "$ACTUAL_CHECKSUM" != "$EXPECTED_CHECKSUM" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Checksum mismatch:\nExpected: $EXPECTED_CHECKSUM\nActual: $ACTUAL_CHECKSUM"
    fi
    success "Verified checksum: $ACTUAL_CHECKSUM"
    rm -f "$CHECKSUMS_FILE"
    
    # Extract to bin directory
    tar -xzf "$TEMP_ARCHIVE" -C "$CONTRIBUTOOR_BIN" & 
    if [ ! -f "$CONTRIBUTOOR_BIN/contributoor" ]; then
        fail "Failed to extract installer binary"
    fi
    success "Extracted archive"
    
    # Make binary executable
    chmod +x "$CONTRIBUTOOR_BIN/contributoor"
    success "Set installer permissions: $CONTRIBUTOOR_BIN/contributoor"
    
    # Cleanup
    rm -f "$TEMP_ARCHIVE"
}

# Function to install contributoor binary
install_contributoor_binary() {
    # Create a temp file for the archive
    TEMP_ARCHIVE=$(mktemp)

    # Download and verify checksums
    CHECKSUMS_URL="https://github.com/ethpandaops/contributoor-test/releases/download/v${VERSION}/contributoor-test_${VERSION}_checksums.txt"
    CHECKSUMS_FILE=$(mktemp)
    
    curl -L -f -s "$CHECKSUMS_URL" -o "$CHECKSUMS_FILE" &
    wait $!
    if [ ! -f "$CHECKSUMS_FILE" ] || [ ! -s "$CHECKSUMS_FILE" ]; then
        rm -f "$CHECKSUMS_FILE"
        fail "Failed to download checksums file"
    fi
    
    # Download and verify
    curl -L -f -s "$CONTRIBUTOOR_URL" -o "$TEMP_ARCHIVE" & 
    spinner $!
    wait $!
    if [ ! -f "$TEMP_ARCHIVE" ] || [ ! -s "$TEMP_ARCHIVE" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Failed to download contributoor binary:\n- File exists: $([ -f "$TEMP_ARCHIVE" ] && echo "yes" || echo "no")\n- File has content: $([ -s "$TEMP_ARCHIVE" ] && echo "yes" || echo "no")"
    fi
    success "Downloaded contributoor"
    
    # Verify checksum
    BINARY_NAME="contributoor-test_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    EXPECTED_CHECKSUM=$(grep "$BINARY_NAME" "$CHECKSUMS_FILE" | cut -d' ' -f1)
    if [ -z "$EXPECTED_CHECKSUM" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Checksum not found for $BINARY_NAME"
    fi

    ACTUAL_CHECKSUM=$(sha256sum "$TEMP_ARCHIVE" | cut -d' ' -f1)
    if [ "$ACTUAL_CHECKSUM" != "$EXPECTED_CHECKSUM" ]; then
        rm -f "$CHECKSUMS_FILE" "$TEMP_ARCHIVE"
        fail "Checksum mismatch:\nExpected: $EXPECTED_CHECKSUM\nActual: $ACTUAL_CHECKSUM"
    fi
    success "Verified checksum: $ACTUAL_CHECKSUM"
    rm -f "$CHECKSUMS_FILE"
    
    # Extract to bin directory
    tar -xzf "$TEMP_ARCHIVE" -C "$CONTRIBUTOOR_BIN" & 
    if [ ! -f "$CONTRIBUTOOR_BIN/sentry" ]; then
        fail "Failed to extract contributoor binary:\n- Archive exists: $([ -f "$TEMP_ARCHIVE" ] && echo "yes" || echo "no")\n- Bin dir exists: $([ -d "$CONTRIBUTOOR_BIN" ] && echo "yes" || echo "no")\n- Bin dir contents:\n$(ls -la "$CONTRIBUTOOR_BIN")"
    fi
    success "Extracted archive"
    
    # Make binary executable
    chmod +x "$CONTRIBUTOOR_BIN/sentry"
    success "Set contributoor permissions: $CONTRIBUTOOR_BIN/sentry"
    
    # Cleanup
    rm -f "$TEMP_ARCHIVE"
}

# Run installation of installer.
install_installer_binary

# If binary mode selected, install the contributoor binary (it'll later be run by the installer).
if [ "$INSTALL_MODE" = "binary" ]; then
    install_contributoor_binary
fi

# If docker mode selected, check system resources.
if [ "$INSTALL_MODE" = "docker" ]; then
    # Clean up any stale Docker resources, will remove potential conflicts from previous failed installations
    # and all round give docker a fresh start.
    if command -v docker >/dev/null 2>&1; then
        docker system prune -f >/dev/null 2>&1 || true
    fi
fi

# Write installer config file
progress 7 "Writing configuration"
CONFIG_FILE="$CONTRIBUTOOR_PATH/contributoor.yaml"
cat > "$CONFIG_FILE" << EOF
version: ${VERSION}
contributoorDirectory: ${CONTRIBUTOOR_PATH}
runMethod: ${INSTALL_MODE}
EOF
success "Created config: $CONFIG_FILE"

# Run initial install
progress 8 "Run install wizard"



"$CONTRIBUTOOR_BIN/contributoor" --config-path "$CONTRIBUTOOR_PATH" install --version "$VERSION" --run-method "$INSTALL_MODE"

