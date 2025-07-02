# lazygit-mcp-bridge

Bridge between lazygit and AI assistants via Model Context Protocol (MCP)

## Overview

`lazygit-mcp-bridge` enables seamless integration between lazygit and AI coding assistants (like Claude, ChatGPT, etc.) through the Model Context Protocol. When you're reviewing code changes in lazygit, you can instantly send comments and context to your AI assistant without copy-pasting.

## Features

- 📝 Send code diff comments directly from lazygit
- 🚀 Real-time message delivery via file watching
- 🤖 Works with any MCP-compatible AI assistant
- ⚡ Zero-copy workflow - no manual clipboard operations needed

## Installation

### 1. Install the MCP server

```bash
go install github.com/cappyzawa/lazygit-mcp-bridge@latest
```

### 2. Register with your AI assistant

For Claude Code:
```bash
claude mcp add lazygit-mcp-bridge lazygit-mcp-bridge
```

### 3. Configure lazygit

Add to your `~/.config/jesseduffield/lazygit/config.yml`:

```yaml
customCommands:
- key: "<c-y>"
  context: "staging"
  description: "Send comment to AI assistant"
  loadingText: "Sending comment…"
  prompts:
  - type: "input"
    title: "Comment:"
  command: |
    FILE_PATH="{{ .SelectedPath }}"
    COMMENT="{{ index .PromptResponses 0 }}"
    if [ -n "$FILE_PATH" ]; then
      ~/.config/jesseduffield/lazygit/send-to-ai.sh "$FILE_PATH" "$COMMENT" >/dev/null 2>&1
    fi
```

### 4. Create the helper script

Create `~/.config/jesseduffield/lazygit/send-to-ai.sh`:

```bash
#!/bin/bash
set -euo pipefail

SELECTED_FILE="$1"
COMMENT="$2"

# Send message to MCP server
MESSAGE_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/jesseduffield/lazygit/mcp-messages.json"
cat > "$MESSAGE_FILE" <<EOF
{
  "file": "$SELECTED_FILE",
  "line": "1",
  "comment": "$COMMENT",
  "time": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
}
EOF
```

Make it executable:
```bash
chmod +x ~/.config/jesseduffield/lazygit/send-to-ai.sh
```

## Usage

### Basic Usage

1. Open lazygit
2. Navigate to the diff view
3. Press `Ctrl+Y`
4. Enter your comment
5. The AI assistant receives your message automatically!

### Using Custom Commands (Recommended)

For the best experience, set up a custom Claude Code command:

1. Create the command directory in your project:
   ```bash
   mkdir -p .claude/commands
   ```

2. Create `.claude/commands/lg.md`:
   ```markdown
   ---
   allowed-tools: mcp__lazygit-mcp-bridge__check_lazygit_messages
   description: Check for new lazygit comments and provide concise code improvement suggestions
   ---

   # lazygit Comment Check

   Use the MCP tool `mcp__lazygit-mcp-bridge__check_lazygit_messages` to retrieve the latest comment from lazygit.

   Then provide concise, focused code improvement suggestions based on the received message.

   Keep responses brief and actionable.

   Additional context: $ARGUMENTS
   ```

3. Now you can use `/project:lg` in Claude Code to instantly check for lazygit messages!

## How it works

1. When you press `Ctrl+Y` in lazygit, it executes the custom command
2. The script writes your comment to a JSON file
3. The MCP server watches for file changes
4. When detected, it queues the message
5. Your AI assistant can retrieve messages via MCP tools

## Development

```bash
# Clone the repository
git clone https://github.com/cappyzawa/lazygit-mcp-bridge
cd lazygit-mcp-bridge

# Install dependencies
go mod tidy

# Build
go build

# Run tests
go test ./...
```

## License

MIT

## Documentation

- [Architecture Overview](docs/architecture.md) - System design and components
- [MCP Protocol Specification](docs/mcp-protocol.md) - Protocol implementation details
- [Custom Commands Guide](docs/custom-commands.md) - Advanced Claude Code integration
- [Development Guide](docs/development.md) - Building and contributing

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. See the [Development Guide](docs/development.md) for details.