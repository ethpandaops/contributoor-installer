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

@test "get_latest_version returns valid version" {
    # Mock curl to return a valid GitHub API response
    function curl() {
        echo '{"tag_name": "v1.2.3"}'
        return 0
    }
    export -f curl

    run get_latest_version
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

@test "get_latest_version fails when bad response from GitHub is returned" {
    function curl() {
        return 1
    }
    export -f curl

    run get_latest_version

    echo "Status: $status"
    echo "Output: '$output'"

    [ "$status" -eq 1 ]
    echo "$output" | grep -F -q "**ERROR**"
    echo "$output" | grep -q "Failed to determine latest version"
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

@test "setup_installer downloads and verifies checksums" {
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="x86_64"
    PLATFORM="linux"
    INSTALLER_BINARY_NAME="contributoor-installer_${PLATFORM}_${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/latest/download/${INSTALLER_BINARY_NAME}.tar.gz"
    
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
            *"/checksums.txt")
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
        touch "$CONTRIBUTOOR_BIN/contributoor"
        touch "$CONTRIBUTOOR_BIN/docker-compose.yml"
        return 0
    }
    
    export -f curl sha256sum tar
    
    run setup_installer
    
    [ "$status" -eq 0 ]
    [ -f "$CONTRIBUTOOR_BIN/contributoor" ]
    [ -f "$CONTRIBUTOOR_BIN/docker-compose.yml" ]
}

@test "setup_installer fails on checksum mismatch" {
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="x86_64"
    PLATFORM="linux"
    INSTALLER_BINARY_NAME="contributoor-installer_${PLATFORM}_${ARCH}"
    INSTALLER_URL="https://github.com/ethpandaops/contributoor-installer/releases/latest/download/${INSTALLER_BINARY_NAME}.tar.gz"
    
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
            *"/checksums.txt")
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
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="x86_64"
    PLATFORM="linux"
    VERSION="1.0.0"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${VERSION}/contributoor_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    
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
                # Match the exact format expected by the script
                cat > "$output_file" << EOF
0123456789abcdef  contributoor_1.0.0_linux_x86_64.tar.gz
EOF
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
        touch "$CONTRIBUTOOR_BIN/sentry"
        return 0
    }
    
    export -f curl sha256sum tar
    
    run setup_binary_contributoor
    
    [ "$status" -eq 0 ]
    [ -f "$CONTRIBUTOOR_BIN/sentry" ]
}

@test "setup_binary_contributoor fails on missing binary" {
    mkdir -p "$CONTRIBUTOOR_BIN"
    
    # Set required variables
    ARCH="x86_64"
    PLATFORM="linux"
    VERSION="1.0.0"
    CONTRIBUTOOR_URL="https://github.com/ethpandaops/contributoor/releases/download/v${VERSION}/contributoor_${VERSION}_${PLATFORM}_${ARCH}.tar.gz"
    
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
                cat > "$output_file" << EOF
0123456789abcdef  contributoor_${VERSION}_${PLATFORM}_${ARCH}.tar.gz
EOF
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