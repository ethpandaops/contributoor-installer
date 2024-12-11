#!/bin/sh

TOTAL_STEPS="2"
CONTRIBUTOOR_PATH=${CONTRIBUTOOR_PATH:-"$HOME/.contributoor"}

while getopts "p:" FLAG; do
    case "$FLAG" in
        p) CONTRIBUTOOR_PATH="$OPTARG" ;;
        *) fail "Incorrect usage." ;;
    esac
done

# Print progress
progress() {
    STEP_NUMBER=$1
    MESSAGE=$2
    echo "Step $STEP_NUMBER of $TOTAL_STEPS: $MESSAGE"
}

progress 1 "Checking for existing installation at $CONTRIBUTOOR_PATH..."
if [ -d $CONTRIBUTOOR_PATH ]; then 
    echo "Existing installation found, skipping..."
    exit 0
fi

progress 2 "Creating contributoor directory at $CONTRIBUTOOR_PATH..."
mkdir -p "$CONTRIBUTOOR_PATH"

# Further remote installation steps here...
