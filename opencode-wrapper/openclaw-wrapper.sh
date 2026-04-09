#!/bin/bash

# OpenClaw Session Wrapper
# Automatically saves sessions to .md files when OpenClaw exits
# Usage: ./openclaw-wrapper.sh <agent_name> <query>
#   e.g., ./openclaw-wrapper.sh coder "Write a Python hello world"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK_DIR="${PWD}"
SESSIONS_DIR="${WORK_DIR}/.sessions"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
SESSION_FILE="${SESSIONS_DIR}/openclaw_session_${TIMESTAMP}.md"
TEMP_CONFIG="/tmp/openclaw_config_${TIMESTAMP}.json"

# Show usage if no arguments
if [ $# -eq 0 ]; then
    echo "Usage: $0 <agent_name> <query>"
    echo ""
    echo "Available agents:"
    for agent in "${SCRIPT_DIR}"/agent_*.json; do
        if [ -f "$agent" ]; then
            basename "$agent" .json | sed 's/agent_/  - /'
        fi
    done
    echo ""
    echo "Example:"
    echo "  $0 coder \"Write a Python hello world script\""
    echo ""
    echo "Or pass OpenClaw commands directly:"
    echo "  $0 gateway --help"
    exit 1
fi

# Create .sessions directory if it doesn't exist
mkdir -p "${SESSIONS_DIR}"

# Check if first argument is an agent name
AGENT_NAME=""
AGENT_FILE=""
QUERY=""
OPENCLAW_ARGS=()

# Check if first arg is an agent name (requires query as second arg)
if [ -f "${SCRIPT_DIR}/agent_${1}.json" ]; then
    if [ $# -lt 2 ]; then
        echo "Error: Agent '$1' requires a query"
        echo "Usage: $0 $1 \"your query here\""
        exit 1
    fi
    AGENT_NAME="$1"
    AGENT_FILE="${SCRIPT_DIR}/agent_${AGENT_NAME}.json"
    shift
    QUERY="$*"
else
    # Not an agent, pass all args to openclaw
    OPENCLAW_ARGS=("$@")
fi

# Extract config details
extract_config() {
    local config_file="$1"
    if [ -f "$config_file" ]; then
        MODEL=$(grep -o '"model"[[:space:]]*:[[:space:]]*"[^"]*"' "$config_file" | head -1 | sed 's/"model"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/')
        SYSTEM_PROMPT=$(grep -o '"SOUL"[[:space:]]*:[[:space:]]*"[^"]*"' "$config_file" | sed 's/"SOUL"[[:space:]]*:[[:space:]]*"\([^"]*\)"/\1/' | cut -c1-100)
        TOOLS=$(grep -o '"tools"[[:space:]]*:[[:space:]]*\[[^]]*\]' "$config_file" | sed 's/"tools"[[:space:]]*:[[:space:]]*\[\([^]]*\)\]/\1/' | tr -d '"' | tr ',' ' ')
    fi
}

# Trap EXIT signal to save session
cleanup() {
    echo ""
    echo "Saving OpenClaw session to ${SESSION_FILE}..."
    
    local config_used="${AGENT_FILE:-${HOME}/.openclaw/openclaw.json}"
    extract_config "$config_used"
    
    # Create session markdown file
    cat > "${SESSION_FILE}" << EOF
# OpenClaw Session - $(date +"%Y-%m-%d %H:%M:%S")

**Working Directory:** ${WORK_DIR}

**Session Duration:** Started at ${TIMESTAMP}

**Agent:** ${AGENT_NAME:-"default"}

**Query:** ${QUERY:-"${OPENCLAW_ARGS[*]}"}

## Configuration

**Model:** ${MODEL:-"Not configured"}

**System Prompt:** ${SYSTEM_PROMPT:-"Not configured"}

**Tools:** ${TOOLS:-"Not configured"}

---

## Session Summary

This session was automatically saved by openclaw-wrapper.

EOF

    if [ -n "$AGENT_NAME" ]; then
        cat >> "${SESSION_FILE}" << EOF
To replay this session, use:
\`\`\`bash
./openclaw-wrapper.sh ${AGENT_NAME} "${QUERY}"
\`\`\`

Or directly with OpenClaw:
\`\`\`bash
cp ${AGENT_FILE} ~/.openclaw/openclaw.json
openclaw agent --message "${QUERY}"
\`\`\`

EOF
    fi
    
    # Clean up temp config
    rm -f "$TEMP_CONFIG"
    
    echo "OpenClaw session saved successfully!"
}

# Register cleanup function to run on exit
trap cleanup EXIT INT TERM

# Launch OpenClaw
if [ -n "$AGENT_FILE" ]; then
    echo "Using agent: ${AGENT_NAME}"
    echo "Query: ${QUERY}"
    echo ""
    
    # Copy agent config to OpenClaw config location
    mkdir -p "${HOME}/.openclaw"
    cp "$AGENT_FILE" "${HOME}/.openclaw/openclaw.json"
    
    # Run OpenClaw with the query (using --agent main for main session)
    openclaw agent --agent main --message "$QUERY"
else
    # No agent specified, pass all args to openclaw
    openclaw "${OPENCLAW_ARGS[@]}"
fi
