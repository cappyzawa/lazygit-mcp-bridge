#!/bin/bash

# Script to copy file diff information with comments for lazygit
# Usage: ./copy-comment.sh <selected_file> <comment>

set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <selected_file> <comment>" >&2
  exit 1
fi

SELECTED_FILE="$1"
COMMENT="$2"

# Get current line information from git diff
# Estimate line numbers from diff information displayed in lazygit's main panel
get_current_line_info() {
  local file="$1"

  # Get changed line information from git diff
  # Use -U0 to set context lines to 0 and show only changed lines
  git diff --no-index --unified=0 /dev/null "$file" 2>/dev/null |
    grep -E "^@@|^\+" |
    head -20 |
    tail -10 |
    grep -E "^@@" |
    tail -1 |
    sed 's/@@.*+\([0-9]*\).*/\1/' || echo "1"
}

# Check if file exists
if [ ! -f "$SELECTED_FILE" ]; then
  echo "File not found: $SELECTED_FILE" >&2
  exit 1
fi

# Get line number information (use first changed line as simple approach)
LINE_NUM=$(get_current_line_info "$SELECTED_FILE")

# Send message to MCP server via JSON file
MESSAGE_FILE="$HOME/.config/jesseduffield/lazygit/claude-messages.json"
cat > "$MESSAGE_FILE" <<EOF
{
  "file": "$SELECTED_FILE",
  "line": "$LINE_NUM",
  "comment": "$COMMENT", 
  "time": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF

# Also copy to clipboard as fallback
OUTPUT=$(
  cat <<EOF
File: $SELECTED_FILE
Line: $LINE_NUM
Comment: $COMMENT

Please improve this code.
EOF
)

echo "$OUTPUT" | pbcopy

# Silent completion

