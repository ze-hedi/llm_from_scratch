#!/bin/bash

# Example: Test the coder agent with OpenClaw
# This script demonstrates how to use the openclaw-wrapper.sh with an agent

echo "========================================"
echo "  OpenClaw Coder Agent Example"
echo "========================================"
echo ""

# Test query
QUERY="Write a simple Python script that prints 'Hello, World!' and includes a function to add two numbers. Add proper documentation and type hints."

echo "Agent: coder"
echo "Query: $QUERY"
echo ""
echo "Running openclaw-wrapper.sh..."
echo "----------------------------------------"
echo ""

# Run the wrapper with coder agent
./openclaw-wrapper.sh coder "$QUERY"

# Show the session file
echo ""
echo "========================================"
echo "  Session File Created"
echo "========================================"
echo ""

SESSION_FILE=$(ls -t .sessions/openclaw_session_*.md | head -1)
echo "Location: $SESSION_FILE"
echo ""
echo "Contents:"
echo "----------------------------------------"
cat "$SESSION_FILE"
