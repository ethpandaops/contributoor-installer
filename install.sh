#!/bin/bash

###############################################################################
# Xatu Contributoor Installer
# Configures and installs dependencies for the Xatu Contributoor service
###############################################################################

###############################################################################
# Configuration
###############################################################################

# Colors for output
COLOR_RED='\033[0;31m'
COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[33m'
COLOR_CYAN='\033[0;36m'
COLOR_RESET='\033[0m'
COLOR_BOLD='\033[1m'

# Installation defaults
TOTAL_STEPS="8"
CONTRIBUTOOR_PATH=${CONTRIBUTOOR_PATH:-"$HOME/.contributoor"}
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
VERSION="latest"

###############################################################################
# UI Functions
###############################################################################

print_logo() {
    printf "${COLOR_CYAN}"
    cat << "EOF"
   ______            __       _ __          __                  
  / ____/___  ____  / /______(_) /_  __  __/ /_____  ____  _____
 / /   / __ \/ __ \/ __/ ___/ / __ \/ / / / __/ __ \/ __ \/ ___/
/ /___/ /_/ / / / / /_/ /  / / /_/ / /_/ / /_/ /_/ / /_/ / /    
\____/\____/_/ /_/\__/_/  /_/_.___/\__,_/\__/\____/\____/_/     
                                                             
EOF
    printf "${COLOR_RESET}"
    printf "${COLOR_BOLD}Authored by the team at ethpandaops.io${COLOR_RESET}"
}

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

###############################################################################
# Helper Functions
###############################################################################

fail() {
    MESSAGE=$1
    printf "\n${COLOR_RED}**ERROR**\n%b${COLOR_RESET}\n" "$MESSAGE" >&2
    exit 1
}

progress() {
    STEP_NUMBER=$1
    MESSAGE=$2
    printf "\n\n${COLOR_BLUE}Step ${STEP_NUMBER} of ${TOTAL_STEPS}${COLOR_RESET}: ${COLOR_BOLD}${MESSAGE}${COLOR_RESET}"
}

success() {
    MESSAGE=$1
    printf "\n${COLOR_GREEN}✓ %s${COLOR_RESET}" "$MESSAGE"
}

usage() {
    echo "Usage: $0 [-p path] [-v version]"
    echo "  -p: Path to install contributoor (default: $HOME/.contributoor)"
    echo "  -v: Version of contributoor to install without 'v' prefix (default: latest, example: 0.0.6)"
    exit 1
}

###############################################################################
# System Detection Functions
###############################################################################

detect_architecture() {
    local uname_val=$(uname -m)
    case $uname_val in
        x86_64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)       fail "CPU architecture not supported: $uname_val" ;;
    esac
}

detect_platform() {
    local platform=$(uname -s)
    case "$platform" in
        Linux)  echo "linux" ;;
        Darwin) echo "darwin" ;;
        *)      fail "Operating system not supported: $platform" ;;
    esac
}

###############################################################################
# Path Management
###############################################################################

add_to_path() {
    local shell_rc=""
    case "$SHELL" in
        */bash)
            if [ -f "$HOME/.bash_profile" ]; then
                shell_rc="$HOME/.bash_profile"
            else
                shell_rc="$HOME/.bashrc"
            fi
            ;;
        */zsh)  shell_rc="$HOME/.zshrc" ;;
        *)      shell_rc="$HOME/.profile" ;;
    esac

    if [ -n "$shell_rc" ] && [ -f "$shell_rc" ]; then
        local path_line="export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\""
        if ! grep -Fxq "$path_line" "$shell_rc"; then
            echo "$path_line" >> "$shell_rc"
            echo "Added $CONTRIBUTOOR_BIN to PATH in $shell_rc"
            echo "NOTE: You'll need to run 'source $shell_rc' or start a new terminal for the PATH changes to take effect"
        fi
    fi

    export PATH="$PATH:$CONTRIBUTOOR_BIN"
}

###############################################################################
# Installation Functions
###############################################################################

setup_installer() {
    local temp_archive=$(mktemp)
    local checksums_url="https://github.com/ethpandaops/contributoor-installer/releases/latest/download/checksums.txt"
    local checksums_file=$(mktemp)
    
    # Download checksums
    curl -L -f -s "$checksums_url" -o "$checksums_file" &
    wait $!
    [ ! -f "$checksums_file" ] || [ ! -s "$checksums_file" ] && {
        rm -f "$checksums_file"
        fail "Failed to download checksums file"
    }
    
    # Download installer
    curl -L -f -s "$INSTALLER_URL" -o "$temp_archive" &
    spinner $!
    wait $!
    [ ! -f "$temp_archive" ] || [ ! -s "$temp_archive" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Failed to download installer binary"
    }
    success "Downloaded installer"
    
    # Verify checksum
    local binary_name="${INSTALLER_BINARY_NAME}.tar.gz"
    local expected_checksum=$(grep "$binary_name" "$checksums_file" | cut -d' ' -f1)
    [ -z "$expected_checksum" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Checksum not found for $binary_name"
    }

    local actual_checksum=$(sha256sum "$temp_archive" | cut -d' ' -f1)
    [ "$actual_checksum" != "$expected_checksum" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Checksum mismatch:\nExpected: $expected_checksum\nActual: $actual_checksum"
    }
    success "Verified checksum: $actual_checksum"
    rm -f "$checksums_file"
    
    # Extract and set permissions
    tar --no-same-owner -xzf "$temp_archive" -C "$CONTRIBUTOOR_BIN" &
    spinner $!
    wait $!

    [ ! -f "$CONTRIBUTOOR_BIN/contributoor" ] && fail "Failed to extract installer binary"
    success "Extracted archive"
    
    chmod +x "$CONTRIBUTOOR_BIN/contributoor"
    [ -f "$CONTRIBUTOOR_BIN/docker-compose.yml" ] && {
        chmod 644 "$CONTRIBUTOOR_BIN/docker-compose.yml"
        chmod 755 "$CONTRIBUTOOR_BIN"
    } || fail "docker-compose.yml not found after extraction"
    
    success "Set installer permissions: $CONTRIBUTOOR_BIN/contributoor"
    rm -f "$temp_archive"
}

setup_docker_contributoor() {
    docker system prune -f >/dev/null 2>&1 || true

    docker pull "ethpandaops/contributoor:${VERSION}" >/dev/null 2>&1 &
    spinner $!
    wait $!
    [ $? -ne 0 ] && fail "Failed to pull docker image"
    success "Pulled docker image: ethpandaops/contributoor:${VERSION}"
}

setup_binary_contributoor() {
    local temp_archive=$(mktemp)
    local checksums_url="https://github.com/ethpandaops/contributoor/releases/download/v${VERSION}/contributoor_${VERSION}_checksums.txt"
    local checksums_file=$(mktemp)
    
    # Download checksums
    curl -L -f -s "$checksums_url" -o "$checksums_file" &
    wait $!
    [ ! -f "$checksums_file" ] || [ ! -s "$checksums_file" ] && {
        rm -f "$checksums_file"
        fail "Failed to download checksums file"
    }
    
    # Download contributoor
    curl -L -f -s "$CONTRIBUTOOR_URL" -o "$temp_archive" & 
    spinner $!
    wait $!
    [ ! -f "$temp_archive" ] || [ ! -s "$temp_archive" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Failed to download contributoor binary"
    }
    success "Downloaded contributoor"
    
    # Verify checksum
    local binary_name="contributoor_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    local expected_checksum=$(grep "$binary_name" "$checksums_file" | cut -d' ' -f1)
    [ -z "$expected_checksum" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Checksum not found for $binary_name"
    }

    local actual_checksum=$(sha256sum "$temp_archive" | cut -d' ' -f1)
    [ "$actual_checksum" != "$expected_checksum" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Checksum mismatch"
    }
    success "Verified checksum: $actual_checksum"
    rm -f "$checksums_file"
    
    # Extract and set permissions
    tar --no-same-owner -xzf "$temp_archive" -C "$CONTRIBUTOOR_BIN" &
    spinner $!
    wait $!

    [ ! -f "$CONTRIBUTOOR_BIN/sentry" ] && fail "Failed to extract contributoor binary"
    success "Extracted archive"
    
    chmod +x "$CONTRIBUTOOR_BIN/sentry"
    success "Set contributoor permissions: $CONTRIBUTOOR_BIN/sentry"
    rm -f "$temp_archive"

    # After setting permissions, create service files based on platform
    if [ "$INSTALL_MODE" = "binary" ]; then
        # Create logs directory for binary output
        mkdir -p "$CONTRIBUTOOR_PATH/logs" || fail "Could not create the contributoor logs directory"
        chmod -R 755 "$CONTRIBUTOOR_PATH/logs"
        success "Created logs directory: $CONTRIBUTOOR_PATH/logs"
    fi
}

# Check if docker is installed and running
check_docker() {
    # Check if docker command exists
    if ! command -v docker >/dev/null 2>&1; then
        fail "Docker is not installed. Please install Docker first: https://docs.docker.com/get-docker/"
    fi

    # Check if docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        fail "Docker daemon is not running. Please start Docker and try again."
    fi

    # Check if docker compose is available (either as plugin or standalone)
    if ! (docker compose version >/dev/null 2>&1 || docker-compose version >/dev/null 2>&1); then
        fail "Docker Compose is not installed. Please install Docker Compose: https://docs.docker.com/compose/install/"
    fi
}

###############################################################################
# Version Management
###############################################################################

get_latest_version() {
    local release_info=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor/releases/latest")
    [ $? -ne 0 ] && fail "Failed to check for latest version"
    
    local version=$(echo "$release_info" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//')
    [ -z "$version" ] && fail "Failed to determine latest version"
    
    echo "$version"
}

validate_version() {
    local version=$1
    local releases=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor/releases")
    [ $? -ne 0 ] && fail "Failed to check available versions"

    if ! echo "$releases" | grep -q "\"tag_name\": *\"v\{0,1\}${version}\""; then
        local available_versions=$(echo "$releases" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//' | head -n 5 | tr '\n' ', ' | sed 's/,$//')
        fail "Last 5 available versions: ${available_versions}\nProvided version ${version} not found"
    fi
}

###############################################################################
# Main Installation Flow
###############################################################################

update_config_file() {
    local config_file="$1"
    local temp_config=$(mktemp)

    # If config exists, read it into temp file
    if [ -f "$config_file" ]; then
        cp "$config_file" "$temp_config"
    else
        touch "$temp_config"
    fi

    # Update only the fields we care about, preserving the rest
    {
        # Add newline if needed
        [ -s "$temp_config" ] && [ "$(tail -c1 "$temp_config" | wc -l)" -eq 0 ] && echo >> "$temp_config"
        
        # Remove fields we want to update - compatible with BSD and GNU sed
        sed -e '/^version:/d' \
            -e '/^contributoorDirectory:/d' \
            -e '/^runMethod:/d' \
            "$temp_config" > "${temp_config}.tmp" && mv "${temp_config}.tmp" "$temp_config"
        
        # Add our fields
        cat >> "$temp_config" << EOF
version: ${VERSION}
contributoorDirectory: ${CONTRIBUTOOR_PATH}
runMethod: ${INSTALL_MODE}
EOF

        mv "$temp_config" "$config_file"
    } || {
        rm -f "$temp_config" "${temp_config}.tmp"
        fail "Failed to update configuration file"
    }
}

main() {
    # Parse arguments
    while getopts "p:v:h" FLAG; do
        case "$FLAG" in
            p) CONTRIBUTOOR_PATH="$OPTARG" ;;
            v) VERSION="$OPTARG" ;;
            h) usage ;;
            *) usage ;;
        esac
    done

    # Setup environment
    CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
    ARCH=$(detect_architecture)
    PLATFORM=$(detect_platform)
    
    # Add to PATH if needed
    case ":$PATH:" in
        *":$CONTRIBUTOOR_BIN:"*) ;; # Already in PATH
        *) add_to_path ;;
    esac

    # Clear screen and show logo
    if [ "${TEST_MODE:-}" != "true" ]; then
        clear
        print_logo
    fi

    # Platform detection
    progress 1 "Detecting platform"
    success "$PLATFORM ($ARCH)"

    # Version management
    progress 2 "Determining version"
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
        success "Latest version: $VERSION"
    else
        validate_version "$VERSION"
        success "Using specified version: $VERSION"
    fi

    # Installation path
    progress 3 "Installation path"
    if [ "${TEST_MODE:-}" != "true" ]; then
        printf "\nWhere would you like to install contributoor? [${COLOR_CYAN}~/.contributoor${COLOR_RESET}]: "
        read -r CUSTOM_PATH
    fi
    if [ -n "$CUSTOM_PATH" ]; then
        CONTRIBUTOOR_PATH="$CUSTOM_PATH"
        CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
    fi
    success "Using path: $CONTRIBUTOOR_PATH"

    # Setup URLs
    INSTALLER_BINARY_NAME="contributoor-installer_${PLATFORM}_"
    [ "$ARCH" = "amd64" ] && INSTALLER_BINARY_NAME="${INSTALLER_BINARY_NAME}x86_64" || INSTALLER_BINARY_NAME="${INSTALLER_BINARY_NAME}${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/latest/download/${INSTALLER_BINARY_NAME}.tar.gz"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${VERSION}/contributoor_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"

    # Installation mode selection
    if [ "${TEST_MODE:-}" != "true" ]; then
        selected=1
        trap 'tput cnorm' EXIT
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
            
            read -r -n1 key
            case "$key" in
                A|k) [ "$selected" -gt 1 ] && selected=$((selected - 1)) ;;
                B|j) [ "$selected" -lt 2 ] && selected=$((selected + 1)) ;;
                "")
                    tput cnorm
                    printf "Selected: "
                    if [ "$selected" = 1 ]; then
                        INSTALL_MODE="docker"
                        # Check docker is available before proceeding
                        check_docker
                        printf "${COLOR_GREEN}Docker${COLOR_RESET}"
                    else
                        INSTALL_MODE="binary"
                        printf "${COLOR_GREEN}Binary${COLOR_RESET}"
                    fi
                    break
                    ;;
            esac
        done
    fi

    # Directory setup
    progress 5 "Setting up directories"
    mkdir -p "$CONTRIBUTOOR_PATH" || fail "Could not create the contributoor user data directory"
    chmod -R 755 "$CONTRIBUTOOR_PATH"
    success "data directory: $CONTRIBUTOOR_PATH" 

    mkdir -p "$CONTRIBUTOOR_BIN" || fail "Could not create the contributoor bin directory"
    chmod -R 755 "$CONTRIBUTOOR_BIN"
    success "bin directory: $CONTRIBUTOOR_BIN" 

    # Create logs directory if needed
    mkdir -p "$CONTRIBUTOOR_PATH/logs" || fail "Could not create the contributoor logs directory"
    chmod -R 755 "$CONTRIBUTOOR_PATH/logs"
    success "logs directory: $CONTRIBUTOOR_PATH/logs"

    # Prepare installation for the mode selected.
    # If binary, download the contributoor binary.
    # If docker, pull the docker image.
    # Makes life easier later on having everything ready.
    progress 6 "Preparing installation"
    setup_installer
    [ "$INSTALL_MODE" = "binary" ] && setup_binary_contributoor

    # Docker cleanup if needed
    if [ "$INSTALL_MODE" = "docker" ] && command -v docker >/dev/null 2>&1; then
        setup_docker_contributoor
    fi

    # Configuration
    progress 7 "Writing configuration"
    local config_file="$CONTRIBUTOOR_PATH/config.yaml"
    update_config_file "$config_file"
    success "Updated config: $config_file"

    # Run installer
    progress 8 "Run install wizard"
    "$CONTRIBUTOOR_BIN/contributoor" --config-path "$CONTRIBUTOOR_PATH" install --version "$VERSION" --run-method "$INSTALL_MODE"
}

# Execute main installation
if [ "${TEST_MODE:-}" != "true" ]; then
    main "$@"
fi

