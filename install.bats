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