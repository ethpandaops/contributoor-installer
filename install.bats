#!/usr/bin/env bats

setup() {
    # Create temp test directory first
    export TEST_DIR="$(mktemp -d)"
    [ -d "$TEST_DIR" ] || fail "Failed to create temp directory"
    
    # Set test environment
    export TEST_MODE=true
    export CONTRIBUTOOR_PATH="$TEST_DIR/.contributoor"
    export VERSION="1.0.0"
    export INSTALL_MODE="docker"
    export CUSTOM_PATH=""
    
    # Source the script without running main
    source './install.sh'
}

teardown() {
    # Clean up test directory if it exists
    [ -d "$TEST_DIR" ] && rm -rf "$TEST_DIR"
}

@test "detect_architecture returns valid architecture" {
    run detect_architecture
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^(amd64|arm64)$ ]]
}

@test "detect_platform returns valid platform" {
    run detect_platform
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^(linux|darwin)$ ]]
}

@test "setup directories creates expected structure" {
    run mkdir -p "$CONTRIBUTOOR_PATH"
    [ "$status" -eq 0 ]
    
    run mkdir -p "$CONTRIBUTOOR_PATH/bin"
    [ "$status" -eq 0 ]
    
    run mkdir -p "$CONTRIBUTOOR_PATH/logs"
    [ "$status" -eq 0 ]
    
    [ -d "$CONTRIBUTOOR_PATH" ]
    [ -d "$CONTRIBUTOOR_PATH/bin" ]
    [ -d "$CONTRIBUTOOR_PATH/logs" ]
}

@test "update_config_file creates valid yaml" {
    # Setup directory first
    mkdir -p "$CONTRIBUTOOR_PATH"
    
    local config_file="$CONTRIBUTOOR_PATH/config.yaml"
    
    # Create initial config with some custom settings
    cat > "$config_file" << EOF
customSetting: value
anotherSetting: true
version: old-version
contributoorDirectory: /old/path
runMethod: old-method
yetAnotherSetting: 123
EOF
    
    # Debug output
    echo "Initial config:"
    cat "$config_file"
    
    run update_config_file "$config_file"
    echo "Status: $status"
    echo "Output: $output"
    
    [ "$status" -eq 0 ]
    [ -f "$config_file" ]
    
    # Debug final config
    echo "Final config:"
    cat "$config_file"
    
    # Verify custom settings were preserved first
    run grep "customSetting: value" "$config_file"
    [ "$status" -eq 0 ]
    
    run grep "anotherSetting: true" "$config_file"
    [ "$status" -eq 0 ]
    
    run grep "yetAnotherSetting: 123" "$config_file"
    [ "$status" -eq 0 ]
    
    # Now verify our updated settings
    run grep -F "version: $VERSION" "$config_file"
    [ "$status" -eq 0 ] || {
        echo "Expected version: $VERSION"
        echo "Config contents:"
        cat "$config_file"
    }
    
    run grep -F "contributoorDirectory: $CONTRIBUTOOR_PATH" "$config_file"
    [ "$status" -eq 0 ] || {
        echo "Expected path: $CONTRIBUTOOR_PATH"
        echo "Config contents:"
        cat "$config_file"
    }
    
    run grep -F "runMethod: $INSTALL_MODE" "$config_file"
    [ "$status" -eq 0 ] || {
        echo "Expected mode: $INSTALL_MODE"
        echo "Config contents:"
        cat "$config_file"
    }
}

@test "get_latest_contributoor_version returns valid version" {
    # Mock curl to return a valid GitHub API response
    function curl() {
        echo '{"tag_name": "v1.2.3"}'
        return 0
    }
    export -f curl

    run get_latest_contributoor_version
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

@test "get_latest_installer_version returns valid version" {
    # Mock curl to return a valid GitHub API response
    function curl() {
        echo '{"tag_name": "v1.2.3"}'
        return 0
    }
    export -f curl

    run get_latest_installer_version
    [ "$status" -eq 0 ]
    [[ "$output" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

@test "validate_version accepts valid version" {
    # Mock curl to return a releases response that includes version 1.0.0
    function curl() {
        echo '{"tag_name": "v1.0.0"}'
        return 0
    }
    export -f curl

    run validate_version "1.0.0"
    [ "$status" -eq 0 ]
}

@test "validate_version fails on invalid version" {
    # Mock curl to return a releases response that doesn't include version 99.99.99
    function curl() {
        echo '{
            "tag_name": "v1.0.1",
            "name": "Release 1.0.1"
        }'
        return 0
    }
    export -f curl
    
    run validate_version "99.99.99"
    [ "$status" -eq 1 ]
    echo "$output" | grep -q "not found"
}

@test "validate_version fails when bad response from GitHub is returned" {
    function curl() {
        return 1
    }
    export -f curl

    run validate_version "1.0.0"
    
    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Last 5 available versions:"
    echo "$output" | grep -q "Provided version 1.0.0 not found"
}

@test "validate_version shows last 5 versions when version not found" {
    # Mock curl to return multiple versions
    function curl() {
        cat << EOF
{
  "data": [
    {"tag_name": "v2.0.0"},
    {"tag_name": "v1.9.0"},
    {"tag_name": "v1.8.0"},
    {"tag_name": "v1.7.0"},
    {"tag_name": "v1.6.0"},
    {"tag_name": "v1.5.0"}
  ]
}
EOF
        return 0
    }
    export -f curl

    run validate_version "1.0.0"
    
    [ "$status" -eq 1 ]
    echo "$output" | grep -q "2.0.0,1.9.0,1.8.0,1.7.0,1.6.0"
    echo "$output" | grep -q "Provided version 1.0.0 not found"
}

@test "get_latest_contributoor_version fails when bad response from GitHub is returned" {
    function curl() {
        return 1
    }
    export -f curl

    run get_latest_contributoor_version

    echo "Status: $status"
    echo "Output: '$output'"

    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Failed to determine latest contributoor version"
}

@test "get_latest_installer_version fails when bad response from GitHub is returned" {
    function curl() {
        return 1
    }
    export -f curl

    run get_latest_installer_version

    echo "Status: $status"
    echo "Output: '$output'"

    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Failed to determine latest installer version"
}

@test "add_to_path adds directory to PATH" {
    # Setup test shell environment
    local test_rc="$TEST_DIR/.bashrc"
    export SHELL="/bin/bash"
    export HOME="$TEST_DIR"
    touch "$test_rc"
    
    run add_to_path
    
    grep -q "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" "$test_rc"
    [ "$?" -eq 0 ]
}

@test "add_to_path skips if directory already in PATH" {
    # Setup test shell environment
    local test_rc="$TEST_DIR/.bashrc"
    export SHELL="/bin/bash"
    export HOME="$TEST_DIR"
    touch "$test_rc"
    
    # Add path entry first
    echo "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" >> "$test_rc"
    
    # Get initial line count
    local initial_lines=$(wc -l < "$test_rc")
    
    run add_to_path
    
    # Verify line count hasn't changed
    local final_lines=$(wc -l < "$test_rc")
    [ "$initial_lines" -eq "$final_lines" ]
    
    # Verify path only appears once
    [ "$(grep -c "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" "$test_rc")" -eq 1 ]
}

@test "add_to_path shows console message when path is added" {
    # Setup test shell environment
    local test_rc="$TEST_DIR/.zshrc"
    export SHELL="/bin/zsh"
    export HOME="$TEST_DIR"
    touch "$test_rc"
    
    # Reset the flag
    export ADDED_TO_PATH=false
    
    # Export all required functions and variables
    export CONTRIBUTOOR_BIN
    export COLOR_GREEN COLOR_YELLOW COLOR_RESET
    export -f add_to_path success warn
    
    # Call add_to_path in a way that preserves the variable
    run bash -c '
        source ./install.sh
        add_to_path
        echo "ADDED_TO_PATH=$ADDED_TO_PATH"
    '
    
    # Verify path was added
    grep -q "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" "$test_rc"
    [ "$?" -eq 0 ]
    
    # Verify success message was shown
    echo "$output" | grep -q "Added.*to PATH in"
    
    # Verify ADDED_TO_PATH was set to true
    echo "$output" | grep -q "ADDED_TO_PATH=true"
}

@test "add_to_path does not show console message when path already exists" {
    # Setup test shell environment
    local test_rc="$TEST_DIR/.zshrc"
    export SHELL="/bin/zsh"
    export HOME="$TEST_DIR"
    touch "$test_rc"
    
    # Add path entry first
    echo "export PATH=\"\$PATH:$CONTRIBUTOOR_BIN\"" >> "$test_rc"
    
    # Reset the flag
    export ADDED_TO_PATH=false
    
    # Export all required functions and variables
    export CONTRIBUTOOR_BIN
    export COLOR_GREEN COLOR_YELLOW COLOR_RESET
    export -f add_to_path success warn
    
    # Call add_to_path in a way that preserves the variable
    run bash -c '
        source ./install.sh
        add_to_path
        echo "ADDED_TO_PATH=$ADDED_TO_PATH"
    '
    
    # Verify warning message was shown
    echo "$output" | grep -q "PATH entry already exists"
    
    # Verify ADDED_TO_PATH is still false
    echo "$output" | grep -q "ADDED_TO_PATH=false"
}

@test "setup_installer downloads and verifies checksums" {
    # Create required directories
    mkdir -p "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}"
    mkdir -p "$CONTRIBUTOOR_PATH/bin"
    
    # Set required variables
    ARCH="amd64"
    PLATFORM="linux"
    INSTALLER_VERSION="1.0.0"
    INSTALLER_BINARY_NAME="contributoor-installer_${INSTALLER_VERSION}_${PLATFORM}_${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/download/v${INSTALLER_VERSION}/${INSTALLER_BINARY_NAME}.tar.gz"
    
    # Mock the curl commands
    function curl() {
        local output_file=""
        local url=""
        
        # Parse arguments
        while (( "$#" )); do
            case "$1" in
                -o)
                    output_file="$2"
                    shift 2
                    ;;
                -*)
                    shift
                    ;;
                *)
                    url="$1"
                    shift
                    ;;
            esac
        done
        
        case "$url" in
            *"checksums.txt")
                echo "0123456789abcdef $INSTALLER_BINARY_NAME.tar.gz" > "$output_file"
                ;;
            *)
                echo "mock binary" > "$output_file"
                ;;
        esac
        return 0
    }
    
    # Mock sha256sum to return expected hash
    function sha256sum() {
        echo "0123456789abcdef  $1"
    }
    
    # Mock tar extraction
    function tar() {
        # Create the binary and make it executable
        touch "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/contributoor"
        chmod +x "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/contributoor"
        
        # Create compose files
        touch "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.yml"
        touch "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.ports.yml"
        touch "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.network.yml"
        
        return 0
    }

    # Mock ln for symlink creation
    function ln() {
        if [[ "$1" == "-sf" ]]; then
            # Create the symlink target
            touch "$3"
            chmod +x "$3"
            
            # Also create compose files in the same directory if it's the binary symlink
            if [[ "$3" == *"/bin/contributoor" ]]; then
                cp "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.yml" "$(dirname "$3")/docker-compose.yml"
                cp "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.ports.yml" "$(dirname "$3")/docker-compose.ports.yml"
                cp "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.network.yml" "$(dirname "$3")/docker-compose.network.yml"
            fi
        fi
        return 0
    }
    
    export -f curl sha256sum tar ln
    
    run setup_installer
    
    echo "Status: $status"
    echo "Output: $output"
    
    [ "$status" -eq 0 ]
    [ -f "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/contributoor" ]
    [ -x "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/contributoor" ]
    [ -f "$CONTRIBUTOOR_PATH/bin/contributoor" ]
    [ -x "$CONTRIBUTOOR_PATH/bin/contributoor" ]
    [ -f "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.yml" ]
    [ -f "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.ports.yml" ]
    [ -f "$CONTRIBUTOOR_PATH/releases/installer-${INSTALLER_VERSION}/docker-compose.network.yml" ]
}

@test "setup_installer fails on checksum mismatch" {
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="amd64"
    PLATFORM="linux"
    INSTALLER_VERSION="1.0.0"
    INSTALLER_BINARY_NAME="contributoor-installer_${INSTALLER_VERSION}_${PLATFORM}_${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/download/v${INSTALLER_VERSION}/${INSTALLER_BINARY_NAME}.tar.gz"
    
    # Mock curl to return different checksums
    function curl() {
        local output_file=""
        local url=""
        
        # Parse arguments
        while (( "$#" )); do
            case "$1" in
                -o)
                    output_file="$2"
                    shift 2
                    ;;
                -*)
                    shift
                    ;;
                *)
                    url="$1"
                    shift
                    ;;
            esac
        done
        
        case "$url" in
            *"checksums.txt")
                echo "different_hash $INSTALLER_BINARY_NAME.tar.gz" > "$output_file"
                ;;
            *)
                echo "mock binary" > "$output_file"
                ;;
        esac
        return 0
    }
    
    function sha256sum() {
        echo "0123456789abcdef  $1"
    }
    
    export -f curl sha256sum
    
    run setup_installer
    
    [ "$status" -eq 1 ]
    echo "$output" | grep -q "Checksum mismatch"
}

@test "setup_docker_contributoor pulls image" {
    # Mock docker commands
    function docker() {
        case "$1" in
            "system")
                return 0
                ;;
            "pull")
                return 0
                ;;
        esac
    }
    
    export -f docker
    
    run setup_docker_contributoor
    [ "$status" -eq 0 ]
}

@test "setup_binary_contributoor downloads and verifies binary checksum" {
    # Create required directories
    mkdir -p "$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}"
    mkdir -p "$CONTRIBUTOOR_PATH/bin"
    
    # Set required variables
    ARCH="amd64"
    PLATFORM="linux"
    CONTRIBUTOOR_VERSION="1.0.0"
    CONTRIBUTOOR_BINARY_NAME="contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${CONTRIBUTOOR_VERSION}/contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    
    # Mock the curl commands
    function curl() {
        local output_file=""
        local url=""
        
        # Parse arguments
        while (( "$#" )); do
            case "$1" in
                -o)
                    output_file="$2"
                    shift 2
                    ;;
                -*)
                    shift
                    ;;
                *)
                    url="$1"
                    shift
                    ;;
            esac
        done
        
        case "$url" in
            *"checksums.txt")
                echo "0123456789abcdef $CONTRIBUTOOR_BINARY_NAME.tar.gz" > "$output_file"
                ;;
            *)
                echo "mock binary" > "$output_file"
                ;;
        esac
        return 0
    }
    
    # Mock sha256sum
    function sha256sum() {
        echo "0123456789abcdef  $1"
    }
    
    # Mock tar
    function tar() {
        # Create the binary and make it executable
        touch "$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}/sentry"
        chmod +x "$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}/sentry"
        return 0
    }
    
    # Mock ln for symlink creation
    function ln() {
        if [[ "$1" == "-sf" ]]; then
            # Create the symlink target
            touch "$3"
            chmod +x "$3"
        fi
        return 0
    }
    
    export -f curl sha256sum tar ln
    
    run setup_binary_contributoor
    
    echo "Status: $status"
    echo "Output: $output"
    
    [ "$status" -eq 0 ]
    [ -f "$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}/sentry" ]
    [ -x "$CONTRIBUTOOR_PATH/releases/contributoor-${CONTRIBUTOOR_VERSION}/sentry" ]
    [ -f "$CONTRIBUTOOR_PATH/bin/sentry" ]
    [ -x "$CONTRIBUTOOR_PATH/bin/sentry" ]
}

@test "setup_binary_contributoor fails on missing binary" {
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="amd64"
    PLATFORM="linux"
    CONTRIBUTOOR_VERSION="1.0.0"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${CONTRIBUTOOR_VERSION}/contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    
    function curl() {
        local output_file=""
        local url=""
        
        # Parse arguments
        while (( "$#" )); do
            case "$1" in
                -o)
                    output_file="$2"
                    shift 2
                    ;;
                -*)
                    shift
                    ;;
                *)
                    url="$1"
                    shift
                    ;;
            esac
        done
        
        case "$url" in
            *"checksums.txt")
                echo "0123456789abcdef contributoor_${CONTRIBUTOOR_VERSION}_${PLATFORM}_${ARCH}.tar.gz" > "$output_file"
                ;;
            *)
                echo "mock binary" > "$output_file"
                ;;
        esac
        return 0
    }
    
    function sha256sum() {
        echo "0123456789abcdef  $1"
    }
    
    # Mock tar to not create the binary
    function tar() {
        return 0
    }
    
    export -f curl sha256sum tar
    
    run setup_binary_contributoor
    
    [ "$status" -eq 1 ]
    echo "$output" | grep -q "Failed to extract contributoor binary"
}

@test "check_docker fails when docker not installed" {
    # Mock command to simulate docker not being installed
    function command() {
        case "$2" in
            "docker") return 1 ;;
            *) command "$@" ;;
        esac
    }
    export -f command

    run check_docker 2>/dev/null
    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Docker is not installed. Please install Docker first: https://docs.docker.com/get-docker/"
}

@test "check_docker fails when docker daemon not running" {
    # Mock docker info to simulate daemon not running
    function docker() {
        case "$1" in
            "info") return 1 ;;
            *) return 0 ;;
        esac
    }
    export -f docker

    run check_docker 2>/dev/null
    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Docker daemon is not running. Please start Docker and try again."
}

@test "check_docker fails when docker compose not available" {
    # Mock docker and command to simulate compose missing
    function docker() {
        case "$1" in
            "compose")
                return 1
                ;;
            "info") return 0 ;;  # Need docker daemon to appear running
            *) return 0 ;;
        esac
    }
    function docker-compose() {
        return 1
    }
    export -f docker docker-compose

    run check_docker 2>/dev/null
    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Docker Compose is not installed. Please install Docker Compose: https://docs.docker.com/compose/install/"
}

@test "check_docker succeeds when everything available" {
    # Mock successful docker environment
    function docker() {
        case "$1" in
            "info") return 0 ;;
            "compose") return 0 ;;
            *) return 0 ;;
        esac
    }
    export -f docker

    run check_docker
    [ "$status" -eq 0 ]
    [ -z "$output" ]
}

@test "[linux] setup_systemd_contributoor creates service file correctly" {
    # Set platform for test
    function detect_platform() {
        echo "linux"
    }
    export -f detect_platform

    # Mock sudo to capture commands
    function sudo() {
        case "$1" in
            "tee")
                # Capture service file content
                cat > "$TEST_DIR/contributoor.service"
                ;;
            *) return 0 ;;
        esac
    }
    export -f sudo

    run setup_systemd_contributoor

    # Check status
    [ "$status" -eq 0 ]

    # Verify service file was created and has correct content
    [ -f "$TEST_DIR/contributoor.service" ]
    grep -q "Description=Contributoor Service" "$TEST_DIR/contributoor.service"
    grep -q "User=$USER" "$TEST_DIR/contributoor.service"
    grep -q "ExecStart=$CONTRIBUTOOR_BIN/sentry" "$TEST_DIR/contributoor.service"
}

@test "[linux] setup_systemd_contributoor removes any existing systemd service before installing" {
    # Set platform for test
    function detect_platform() {
        echo "linux"
    }
    export -f detect_platform

    # Mock sudo and systemctl
    function sudo() {
        case "$1" in
            "systemctl")
                case "$2" in
                    "list-unit-files") echo "contributoor.service enabled" ;;
                    *) return 0 ;;
                esac
                ;;
            *) return 0 ;;
        esac
    }
    export -f sudo

    run setup_systemd_contributoor

    # Check status
    [ "$status" -eq 0 ]

    # Verify it detected and handled existing service
    echo "$output" | grep -q "Stopped and disabled existing systemd service"
}

@test "[linux] check_systemd validates systemd availability" {
    # Set platform for test
    function detect_platform() {
        echo "linux"
    }
    export -f detect_platform

    # Mock commands to simulate working systemd
    function pidof() { echo "1"; }
    function sudo() { return 0; }
    function command() {
        case "$2" in
            "systemctl") return 0 ;;
            *) command "$@" ;;
        esac
    }
    export -f pidof sudo command
 
    run check_systemd_or_launchd
    [ "$status" -eq 0 ]
}

@test "[linux] check_systemd fails when systemd not available" {
    # Set platform for test
    function detect_platform() {
        echo "linux"
    }
    export -f detect_platform

    # Mock pidof to simulate no systemd
    function pidof() { return 1; }
    export -f pidof
 
    run check_systemd_or_launchd
    [ "$status" -eq 1 ]
    echo "$output" | grep -q "Systemd is not available on this system. Please choose a different installation mode."
}

@test "[darwin] check_systemd_or_launchd validates launchd availability" {
    # Set platform for test
    function detect_platform() {
        echo "darwin"
    }
    export -f detect_platform

    # Mock launchctl to simulate working launchd
    function command() {
        case "$2" in
            "launchctl") return 0 ;;
            "sudo") return 0 ;;
            *) command "$@" ;;
        esac
    }
    export -f command

    run check_systemd_or_launchd
    [ "$status" -eq 0 ]
}

@test "[darwin] setup_systemd_contributoor removes any existing launchd service before installing" {
    # Set platform for test
    function detect_platform() {
        echo "darwin"
    }
    export -f detect_platform

    # Mock sudo and launchctl
    function sudo() {
        case "$1" in
            "launchctl")
                case "$2" in
                    "list") echo "12345 0 io.ethpandaops.contributoor" ;;
                    *) return 0 ;;
                esac
                ;;
            *) return 0 ;;
        esac
    }
    export -f sudo

    run setup_systemd_contributoor

    # Check status
    [ "$status" -eq 0 ]

    # Verify it detected and handled existing service
    echo "$output" | grep -q "Stopped and unloaded existing launchd service"
}

@test "[darwin] setup_systemd_contributoor creates launchd service file correctly" {
    # Set platform for test
    function detect_platform() {
        echo "darwin"
    }
    export -f detect_platform

    # Mock sudo to capture commands
    function sudo() {
        case "$1" in
            "tee")
                # Capture service file content
                cat > "$TEST_DIR/io.ethpandaops.contributoor.plist"
                ;;
            *) return 0 ;;
        esac
    }
    export -f sudo

    run setup_systemd_contributoor

    # Check status
    [ "$status" -eq 0 ]

    # Verify service file was created and has correct content
    [ -f "$TEST_DIR/io.ethpandaops.contributoor.plist" ]
    grep -q "<string>io.ethpandaops.contributoor</string>" "$TEST_DIR/io.ethpandaops.contributoor.plist"
    grep -q "<string>$CONTRIBUTOOR_BIN/sentry</string>" "$TEST_DIR/io.ethpandaops.contributoor.plist"
    grep -q "<string>$USER</string>" "$TEST_DIR/io.ethpandaops.contributoor.plist"
}

@test "service start prompt handles yes response" {
    # Create test script
    cat > "$TEST_DIR/prompt_test.sh" << 'EOF'
#!/bin/bash
source ./install.sh
contributoor() { return 0; }
export -f contributoor
CONTRIBUTOOR_BIN="$TEST_DIR/bin"
printf "\nWould you like to start contributoor now? [y/N]: "
read -r START_SERVICE
case "$(echo "$START_SERVICE" | tr "[:upper:]" "[:lower:]")" in
    y|yes)
        contributoor --config-path "$TEST_DIR" restart
        ;;
    *)
        printf "You can start contributoor later by running: contributoor start"
        ;;
esac
EOF
    chmod +x "$TEST_DIR/prompt_test.sh"

    # Run test with yes input
    printf 'y\n' | "$TEST_DIR/prompt_test.sh"
    local result=$?

    [ "$result" -eq 0 ]
}

@test "service start prompt handles no response" {
    # Create test script
    cat > "$TEST_DIR/prompt_test.sh" << 'EOF'
#!/bin/bash
source ./install.sh
contributoor() { return 0; }
export -f contributoor
CONTRIBUTOOR_BIN="$TEST_DIR/bin"
printf "\nWould you like to start contributoor now? [y/N]: "
read -r START_SERVICE
case "$(echo "$START_SERVICE" | tr "[:upper:]" "[:lower:]")" in
    y|yes)
        contributoor --config-path "$TEST_DIR" restart
        ;;
    *)
        printf "You can start contributoor later by running: contributoor start"
        ;;
esac
EOF
    chmod +x "$TEST_DIR/prompt_test.sh"

    # Run test with no input
    output=$(printf 'n\n' | "$TEST_DIR/prompt_test.sh")
    local result=$?

    [ "$result" -eq 0 ]
    echo "$output" | grep -q "You can start contributoor later"
}

@test "service start prompt handles empty response" {
    # Create test script
    cat > "$TEST_DIR/prompt_test.sh" << 'EOF'
#!/bin/bash
source ./install.sh
contributoor() { return 0; }
export -f contributoor
CONTRIBUTOOR_BIN="$TEST_DIR/bin"
printf "\nWould you like to start contributoor now? [y/N]: "
read -r START_SERVICE
case "$(echo "$START_SERVICE" | tr "[:upper:]" "[:lower:]")" in
    y|yes)
        contributoor --config-path "$TEST_DIR" restart
        ;;
    *)
        printf "You can start contributoor later by running: contributoor start"
        ;;
esac
EOF
    chmod +x "$TEST_DIR/prompt_test.sh"

    # Run test with empty input
    output=$(printf '\n' | "$TEST_DIR/prompt_test.sh")
    local result=$?

    [ "$result" -eq 0 ]
    echo "$output" | grep -q "You can start contributoor later"
}