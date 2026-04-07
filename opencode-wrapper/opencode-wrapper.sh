#!/bin/bash

# OpenCode Session Wrapper
# Automatically saves sessions to .md files when OpenCode exits

WORK_DIR="${PWD}"
SESSIONS_DIR="${WORK_DIR}/.sessions"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
SESSION_FILE="${SESSIONS_DIR}/session_${TIMESTAMP}.md"

# Create .sessions directory if it doesn't exist
mkdir -p "${SESSIONS_DIR}"

# Trap EXIT signal to save session
cleanup() {
    echo "Saving session to ${SESSION_FILE}..."
    
    # Create session markdown file
    cat > "${SESSION_FILE}" << EOF
# OpenCode Session - $(date +"%Y-%m-%d %H:%M:%S")

**Working Directory:** ${WORK_DIR}

**Session Duration:** Started at ${TIMESTAMP}

---

## Session Summary

This session was automatically saved by opencode-wrapper.

EOF
    
    echo "Session saved successfully!"
}

# Register cleanup function to run on exit
trap cleanup EXIT INT TERM

# Launch OpenCode with all passed arguments
opencode "$@"
