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

# Check if terminal type is supported, fallback to xterm-256color if not
if ! tput clear >/dev/null 2>&1; then
    export TERM=xterm-256color
fi

# Installation defaults
TOTAL_STEPS="9"
CONTRIBUTOOR_PATH=${CONTRIBUTOOR_PATH:-"$HOME/.contributoor"}
CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
CONTRIBUTOOR_VERSION="latest"
ADDED_TO_PATH=false

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

warn() {
    MESSAGE=$1
    printf "\n${COLOR_YELLOW}⚠ %s${COLOR_RESET}" "$MESSAGE"
}

usage() {
    echo "Usage: $0 [-p path] [-v version] [-u]"
    echo "  -p: Path to install contributoor (default: $HOME/.contributoor)"
    echo "  -v: Version of contributoor to install without 'v' prefix (default: latest, example: 0.0.6)"
    echo "  -u: Uninstall Contributoor"
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
        # Less strict grep check
        if ! grep -F "$CONTRIBUTOOR_BIN" "$shell_rc" >/dev/null 2>&1; then
            echo "$path_line" >> "$shell_rc"
            ADDED_TO_PATH=true
            success "Added $CONTRIBUTOOR_BIN to PATH in $shell_rc"
        else
            warn "PATH entry already exists in $shell_rc"
        fi
    else
        warn "Could not determine shell configuration file"
    fi

    export PATH="$PATH:$CONTRIBUTOOR_BIN"
}

###############################################################################
# Installation Functions
###############################################################################

setup_installer() {
    local temp_archive=$(mktemp)
    local checksums_url="https://github.com/ethpandaops/contributoor-installer/releases/download/v${CONTRIBUTOOR_VERSION}/contributoor-installer_${CONTRIBUTOOR_VERSION}_checksums.txt"
    local checksums_file=$(mktemp)
    local release_dir="$CONTRIBUTOOR_PATH/releases/installer-${CONTRIBUTOOR_VERSION}"
    
    # Create version-specific release directory
    mkdir -p "$release_dir"
    
    # Download checksums
    curl -L -f -s "$checksums_url" -o "$checksums_file" &
    wait $!
    [ ! -f "$checksums_file" ] || [ ! -s "$checksums_file" ] && {
        rm -f "$checksums_file"
        fail "Failed to download checksums file from: $checksums_url"
    }
    
    # Download installer
    curl -L -f -s "$INSTALLER_URL" -o "$temp_archive" &
    spinner $!
    wait $!
    [ ! -f "$temp_archive" ] || [ ! -s "$temp_archive" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Failed to download installer binary from: $INSTALLER_URL"
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
    tar --no-same-owner -xzf "$temp_archive" -C "$release_dir" &
    spinner $!
    wait $!

    [ ! -f "$release_dir/contributoor" ] && fail "Failed to extract installer binary"
    success "Extracted archive"
    
    chmod +x "$release_dir/contributoor"

    [ -f "$release_dir/docker-compose.yml" ] && {
        chmod 644 "$release_dir/docker-compose.yml"
        chmod 755 "$release_dir"
    } || fail "docker-compose.yml not found after extraction"

    [ -f "$release_dir/docker-compose.metrics.yml" ] && {
        chmod 644 "$release_dir/docker-compose.metrics.yml"
        chmod 755 "$release_dir"
    } || fail "docker-compose.metrics.yml not found after extraction"

    [ -f "$release_dir/docker-compose.health.yml" ] && {
        chmod 644 "$release_dir/docker-compose.health.yml"
        chmod 755 "$release_dir"
    } || fail "docker-compose.health.yml not found after extraction"

    [ -f "$release_dir/docker-compose.network.yml" ] && {
        chmod 644 "$release_dir/docker-compose.network.yml"
        chmod 755 "$release_dir"
    } || fail "docker-compose.network.yml not found after extraction"
    
    # Create/update symlink
    rm -f "$CONTRIBUTOOR_BIN/contributoor" # Remove existing symlink or file
    if ! ln -sf "$release_dir/contributoor" "$CONTRIBUTOOR_BIN/contributoor"; then
        fail "Failed to create symlink from $CONTRIBUTOOR_BIN/contributoor to $release_dir/contributoor"
    fi
    
    success "Set installer permissions and created symlink: $CONTRIBUTOOR_BIN/contributoor -> $release_dir/contributoor"
    rm -f "$temp_archive"
}

setup_docker_contributoor() {
    docker pull "ethpandaops/contributoor:${CONTRIBUTOOR_VERSION}" >/dev/null 2>&1 &
    spinner $!
    wait $!
    [ $? -ne 0 ] && fail "Failed to pull docker image"
    success "Pulled docker image: ethpandaops/contributoor:${CONTRIBUTOOR_VERSION}"
}

setup_binary_contributoor() {
    local temp_archive=$(mktemp)
    local checksums_url="https://github.com/ethpandaops/contributoor/releases/download/v${CONTRIBUTOOR_VERSION}/contributoor_${CONTRIBUTOOR_VERSION}_checksums.txt"
    local checksums_file=$(mktemp)
    local release_dir="$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}"
    
    # Create version-specific release directory
    mkdir -p "$release_dir"
    
    # Download checksums
    curl -L -f -s "$checksums_url" -o "$checksums_file" &
    wait $!
    [ ! -f "$checksums_file" ] || [ ! -s "$checksums_file" ] && {
        rm -f "$checksums_file"
        fail "Failed to download checksums file from: $checksums_url"
    }
    
    # Download contributoor
    curl -L -f -s "$CONTRIBUTOOR_URL" -o "$temp_archive" & 
    spinner $!
    wait $!
    [ ! -f "$temp_archive" ] || [ ! -s "$temp_archive" ] && {
        rm -f "$checksums_file" "$temp_archive"
        fail "Failed to download contributoor binary from: $CONTRIBUTOOR_URL"
    }
    success "Downloaded contributoor"
    
    # Verify checksum
    local binary_name="contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}.tar.gz"
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
    tar --no-same-owner -xzf "$temp_archive" -C "$release_dir" &
    spinner $!
    wait $!

    [ ! -f "$release_dir/sentry" ] && fail "Failed to extract contributoor binary"
    success "Extracted archive"
    
    chmod +x "$release_dir/sentry"
    chmod 755 "$release_dir"
    
    # Create/update symlink
    rm -f "$CONTRIBUTOOR_BIN/sentry" # Remove existing symlink or file
    if ! ln -sf "$release_dir/sentry" "$CONTRIBUTOOR_BIN/sentry"; then
        fail "Failed to create symlink from $CONTRIBUTOOR_BIN/sentry to $release_dir/sentry"
    fi
    
    success "Set contributoor permissions and created symlink: $CONTRIBUTOOR_BIN/sentry -> $release_dir/sentry"
    rm -f "$temp_archive"
}

setup_systemd_contributoor() {
    # Detect platform and use appropriate service manager
    case "$(detect_platform)" in
        "darwin")
            setup_macos_launchd
            ;;
        *)
            setup_linux_systemd
            ;;
    esac
}

# Setup macOS launchd service
setup_macos_launchd() {
    # Warn about sudo requirement
    warn "Setting up launchd service requires sudo access. "

    # Verify sudo access before proceeding
    if ! sudo -p "Please enter your password: " true; then
        fail "sudo access is required to setup launchd service. Installation aborted."
    fi

    # Stop and unload existing service if it exists
    if sudo launchctl list | grep -q "io.ethpandaops.contributoor"; then
        sudo launchctl stop io.ethpandaops.contributoor
        sudo launchctl unload -w "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
        
        # Remove existing service file
        sudo rm -f "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
        
        success "Stopped and unloaded existing launchd service"
    fi

    # Create launchd plist directory
    sudo mkdir -p "/Library/LaunchDaemons"

    # Create the service file
    sudo tee "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist" >/dev/null << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.ethpandaops.contributoor</string>
    <key>ProgramArguments</key>
    <array>
        <string>$CONTRIBUTOOR_BIN/sentry</string>
        <string>--config</string>
        <string>$CONTRIBUTOOR_PATH/config.yaml</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>$CONTRIBUTOOR_PATH</string>
    <key>StandardOutPath</key>
    <string>$CONTRIBUTOOR_PATH/logs/debug.log</string>
    <key>StandardErrorPath</key>
    <string>$CONTRIBUTOOR_PATH/logs/service.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>
    <key>UserName</key>
    <string>$USER</string>
</dict>
</plist>
EOF

    # Set permissions
    sudo chown root:wheel "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
    sudo chmod 644 "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"

    # Load service (but don't start it)
    sudo launchctl load -w "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"

    success "Created launchd service: /Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
    success "Service configured for manual start"
}

# Setup Linux systemd service
setup_linux_systemd() {
    # Warn about sudo requirement
    warn "Setting up systemd service requires sudo access. "

    # Verify sudo access before proceeding
    if ! sudo -p "Please enter your password: " true; then
        fail "sudo access is required to setup systemd service. Installation aborted."
    fi

    # Stop and disable existing service if it exists
    if sudo systemctl list-unit-files | grep -q "contributoor.service"; then
        sudo systemctl stop contributoor.service
        sudo systemctl disable contributoor.service >/dev/null 2>&1
        
        # Remove existing service file
        sudo rm -f "/etc/systemd/system/contributoor.service"
        
        # Remove any leftover runtime files
        sudo rm -rf "/etc/systemd/system/contributoor.service.d"
        sudo rm -f "/etc/systemd/system/contributoor.service.wants"
        sudo rm -f "/etc/systemd/system/multi-user.target.wants/contributoor.service"
        
        # Reload systemd to recognize the removal
        sudo systemctl daemon-reload

        success "Stopped and disabled existing systemd service"
    fi

    # Create systemd directory
    sudo mkdir -p "/etc/systemd/system"

    # Create the service file
    sudo tee "/etc/systemd/system/contributoor.service" >/dev/null << EOF
[Unit]
Description=Contributoor Service
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=0

[Service]
Type=simple
User=$USER
Group=$USER
ExecStart=$CONTRIBUTOOR_BIN/sentry --config $CONTRIBUTOOR_PATH/config.yaml
WorkingDirectory=$CONTRIBUTOOR_PATH
Restart=always
RestartSec=5

# Environment setup
Environment=HOME=$HOME
Environment=USER=$USER
Environment=PATH=/usr/local/bin:/usr/bin:/bin

# Hardening
NoNewPrivileges=true
ProtectSystem=full
ProtectHome=read-only
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # Set permissions
    sudo chmod 644 "/etc/systemd/system/contributoor.service"

    # Reload systemd
    sudo systemctl daemon-reload

    # Enable but don't start the service
    sudo systemctl enable contributoor.service >/dev/null 2>&1

    success "Created systemd service: /etc/systemd/system/contributoor.service"
    success "Service configured for manual start"
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

# Check if systemd/launchd is available and running
check_systemd_or_launchd() {
    case "$(detect_platform)" in
        "darwin")
            # Check if launchd is available
            if ! command -v launchctl >/dev/null 2>&1; then
                fail "Launchd not found. This is unexpected on macOS."
            fi

            # Check if sudo is available
            if ! command -v sudo >/dev/null 2>&1; then
                fail "sudo access required for launchd service management."
            fi
            ;;
        *)
            # Linux systemd checks
            # Check if systemd is the init system
            if ! pidof systemd >/dev/null 2>&1; then
                fail "Systemd is not available on this system. Please choose a different installation mode."
            fi

            # Check if systemctl is available
            if ! command -v systemctl >/dev/null 2>&1; then
                fail "Systemctl command not found. Please choose a different installation mode."
            fi

            # Check if we can access system services
            if ! sudo systemctl status >/dev/null 2>&1; then
                fail "System service management not available. Please check your sudo permissions."
            fi
            ;;
    esac
}

###############################################################################
# Version Management
###############################################################################

get_latest_contributoor_version() {
    local release_info=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor/releases/latest")
    [ $? -ne 0 ] && fail "Failed to check for latest contributoor version"
    
    local version=$(echo "$release_info" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//')
    [ -z "$version" ] && fail "Failed to determine latest contributoor version"
    
    echo "$version"
}

get_latest_installer_version() {
    local release_info=$(curl -s "https://api.github.com/repos/ethpandaops/contributoor-installer/releases/latest")
    [ $? -ne 0 ] && fail "Failed to check for latest installer version"
    
    local version=$(echo "$release_info" | grep -o '"tag_name": *"[^"]*"' | cut -d'"' -f4 | sed 's/^v//')
    [ -z "$version" ] && fail "Failed to determine latest installer version: $release_info"
    
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
version: ${CONTRIBUTOOR_VERSION}
contributoorDirectory: ${CONTRIBUTOOR_PATH}
runMethod: ${INSTALL_MODE}
EOF

        mv "$temp_config" "$config_file"
    } || {
        rm -f "$temp_config" "${temp_config}.tmp"
        fail "Failed to update configuration file"
    }
}

uninstall() {
    printf "\n${COLOR_RED}Warning, this will:${COLOR_RESET}\n"
    printf " • Stop and remove any contributoor services (systemd/launchd)\n"
    printf " • Stop and remove any contributoor Docker containers and images\n"
    printf " • Remove contributoor from your PATH\n"
    printf " • Delete all contributoor data from ${HOME}/.contributoor\n\n"
    printf "Are you sure you want to uninstall? [y/N]: "
    read -r confirm
    case "$(echo "$confirm" | tr '[:upper:]' '[:lower:]')" in
        y|yes) ;;
        *) printf "\nUninstall cancelled\n"; exit 1 ;;
    esac

    printf "\n${COLOR_RED}Uninstalling contributoor...${COLOR_RESET}\n"

    # First stop all services based on platform
    case "$(detect_platform)" in
        "darwin")
            if sudo launchctl list | grep -q "io.ethpandaops.contributoor"; then
                sudo launchctl stop io.ethpandaops.contributoor
                sudo launchctl unload -w "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
                success "Stopped launchd service"
                
                sudo rm -f "/Library/LaunchDaemons/io.ethpandaops.contributoor.plist"
                success "Removed launchd service files"
            fi
            ;;
        *)
            if command -v systemctl >/dev/null 2>&1 && sudo systemctl list-unit-files | grep -q "contributoor.service"; then
                sudo systemctl stop contributoor.service
                success "Stopped systemd service"

                sudo systemctl disable contributoor.service >/dev/null 2>&1
                sudo rm -f "/etc/systemd/system/contributoor.service"
                sudo rm -rf "/etc/systemd/system/contributoor.service.d"
                sudo rm -f "/etc/systemd/system/contributoor.service.wants"
                sudo rm -f "/etc/systemd/system/multi-user.target.wants/contributoor.service"
                sudo systemctl daemon-reload
                success "Removed systemd service files"
            fi
            ;;
    esac

    # Stop and clean up docker containers and images if they exist
    if command -v docker >/dev/null 2>&1; then
        # Stop running containers first
        if docker ps | grep -q "contributoor"; then
            docker stop $(docker ps | grep "contributoor" | awk '{print $1}') >/dev/null 2>&1
            success "Stopped Docker containers"
        fi

        # Remove containers
        if docker ps -a | grep -q "contributoor"; then
            docker rm -f $(docker ps -a | grep "contributoor" | awk '{print $1}') >/dev/null 2>&1
            success "Removed Docker containers"
        fi

        # Remove images
        if docker images | grep -q "ethpandaops/contributoor"; then
            docker rmi -f $(docker images | grep "ethpandaops/contributoor" | awk '{print $3}') >/dev/null 2>&1
            success "Removed Docker images"
        fi
    fi

    # Remove PATH entry from shell config
    for rc in "$HOME/.bashrc" "$HOME/.zshrc" "$HOME/.bash_profile" "$HOME/.profile"; do
        if [ -f "$rc" ]; then
            temp_file=$(mktemp)
            grep -v "export PATH=.*contributoor.*bin" "$rc" > "$temp_file"
            mv "$temp_file" "$rc"
            success "Cleaned PATH from $rc"
        fi
    done

    # Remove contributoor directory
    if [ -d "$HOME/.contributoor" ]; then
        rm -rf "$HOME/.contributoor"
        success "Removed contributoor directory"
    fi

    printf "\n\n${COLOR_GREEN}Contributoor has been uninstalled successfully${COLOR_RESET}\n\n"
    exit 0
}

main() {
    # Parse arguments
    while getopts "p:v:hu" FLAG; do
        case "$FLAG" in
            p) CONTRIBUTOOR_PATH="$OPTARG" ;;
            v) CONTRIBUTOOR_VERSION="$OPTARG" ;;
            u) uninstall ;;
            h) usage ;;
            *) usage ;;
        esac
    done

    # Setup environment
    CONTRIBUTOOR_BIN="$CONTRIBUTOOR_PATH/bin"
    ARCH=$(detect_architecture)
    PLATFORM=$(detect_platform)
    
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
    if [ "$CONTRIBUTOOR_VERSION" = "latest" ]; then
        CONTRIBUTOOR_VERSION=$(get_latest_contributoor_version)
        success "Latest contributoor version: $CONTRIBUTOOR_VERSION"
    else
        validate_version "$CONTRIBUTOOR_VERSION"
        success "Using specified version: $CONTRIBUTOOR_VERSION"
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

    # Create logs directory if needed
    mkdir -p "$CONTRIBUTOOR_PATH/logs" || fail "Could not create the contributoor logs directory"
    chmod -R 755 "$CONTRIBUTOOR_PATH/logs"
    success "logs directory: $CONTRIBUTOOR_PATH/logs"

    # Setup URLs
    INSTALLER_BINARY_NAME="contributoor-installer_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/download/v${CONTRIBUTOOR_VERSION}/${INSTALLER_BINARY_NAME}.tar.gz"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${CONTRIBUTOOR_VERSION}/contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}.tar.gz"

    # Installation mode selection
    if [ "${TEST_MODE:-}" != "true" ]; then
        selected=1
        
        # Check if tput works with current terminal
        if ! tput clear >/dev/null 2>&1; then
            export TERM=xterm-256color
        fi
        
        trap 'tput cnorm 2>/dev/null || true' EXIT
        tput civis 2>/dev/null || true
        while true; do
            clear
            print_logo
            progress 1 "Detecting platform..."
            success "$PLATFORM ($ARCH)"
            progress 2 "Determining latest version"
            success "Using version: $CONTRIBUTOOR_VERSION"
            progress 3 "Installation path"
            success "Using path: $CONTRIBUTOOR_PATH"
            progress 4 "Select installation mode"
            printf "\n  %s docker (${COLOR_CYAN}recommended${COLOR_RESET})\n" "$([ "$selected" = 1 ] && echo ">" || echo " ")"
            case "$(detect_platform)" in
                "darwin")
                    printf "  %s launchd\n" "$([ "$selected" = 2 ] && echo ">" || echo " ")"
                    ;;
                *)
                    printf "  %s systemd\n" "$([ "$selected" = 2 ] && echo ">" || echo " ")"
                    ;;
            esac
            printf "  %s binary (development)\n" "$([ "$selected" = 3 ] && echo ">" || echo " ")"
            printf "\nUse arrow keys (↑/↓) or j/k to select, Enter to confirm\n"
            
            read -r -n1 key
            case "$key" in
                A|k) [ "$selected" -gt 1 ] && selected=$((selected - 1)) ;;
                B|j) [ "$selected" -lt 3 ] && selected=$((selected + 1)) ;;
                "")
                    tput cnorm
                    printf "Selected: "
                    if [ "$selected" = 1 ]; then
                        INSTALL_MODE="RUN_METHOD_DOCKER"
                        # Check docker is available before proceeding
                        check_docker
                        printf "${COLOR_GREEN}docker${COLOR_RESET}"
                    elif [ "$selected" = 2 ]; then
                        INSTALL_MODE="RUN_METHOD_SYSTEMD"
                        # Check systemd is available
                        check_systemd_or_launchd
                        printf "${COLOR_GREEN}systemd${COLOR_RESET}"
                    else
                        INSTALL_MODE="RUN_METHOD_BINARY"
                        printf "${COLOR_GREEN}binary${COLOR_RESET}"
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

    # Create releases directory
    mkdir -p "$CONTRIBUTOOR_PATH/releases" || fail "Could not create the contributoor releases directory"
    chmod -R 755 "$CONTRIBUTOOR_PATH/releases"
    success "releases directory: $CONTRIBUTOOR_PATH/releases"

    # Now that all directories are set up and mode is selected, add to PATH if needed
    add_to_path

    # Prepare installation for the mode selected.
    # If binary, download the contributoor binary.
    # If docker, pull the docker image.
    # Makes life easier later on having everything ready.
    progress 6 "Preparing installation"
    setup_installer
    # Both binary and systemd modes need the binary
    if [ "$INSTALL_MODE" = "RUN_METHOD_BINARY" ] || [ "$INSTALL_MODE" = "RUN_METHOD_SYSTEMD" ]; then
        setup_binary_contributoor
    fi
    
    # Setup systemd service if needed
    [ "$INSTALL_MODE" = "RUN_METHOD_SYSTEMD" ] && setup_systemd_contributoor

    # Docker cleanup if needed
    if [ "$INSTALL_MODE" = "RUN_METHOD_DOCKER" ] && command -v docker >/dev/null 2>&1; then
        setup_docker_contributoor
    fi

    # Configuration
    progress 7 "Writing configuration"
    local config_file="$CONTRIBUTOOR_PATH/config.yaml"
    update_config_file "$config_file"
    success "Updated config: $config_file"

    # Run installer
    progress 8 "Run install wizard"
    "$CONTRIBUTOOR_BIN/contributoor" --config-path "$CONTRIBUTOOR_PATH" install --version "$CONTRIBUTOOR_VERSION" --run-method "$INSTALL_MODE"

    # Ask user if they want to start the service
    printf "\nWould you like to start contributoor now? [y/N]: "
    read -r START_SERVICE
    case "$(echo "$START_SERVICE" | tr '[:upper:]' '[:lower:]')" in
        y|yes)
            "$CONTRIBUTOOR_BIN/contributoor" --config-path "$CONTRIBUTOOR_PATH" restart
            ;;
        *)
            printf "${COLOR_YELLOW}You can start contributoor later by running:${COLOR_RESET} contributoor start"
            ;;
    esac

    # Show PATH refresh message if needed.
    if [ "$ADDED_TO_PATH" = true ]; then
        printf "\n\n${COLOR_YELLOW}NOTE: To use contributoor commands, either start a new terminal or run:${COLOR_RESET} source ~/.$(basename "$SHELL")rc\n"
    fi
}

# Execute main installation
if [ "${TEST_MODE:-}" != "true" ]; then
    main "$@"
fi

