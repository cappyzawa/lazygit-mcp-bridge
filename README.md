# lazygit-mcp-bridge

Bridge between lazygit and AI assistants via Model Context Protocol (MCP)

## Overview

`lazygit-mcp-bridge` enables seamless integration between lazygit and AI coding assistants (like Claude, ChatGPT, etc.) through the Model Context Protocol. When you're reviewing code changes in lazygit, you can instantly send comments and context to your AI assistant without copy-pasting.

## Features

- 📝 Send code diff comments directly from lazygit
- 🚀 Real-time message delivery via file watching
- 🤖 Works with any MCP-compatible AI assistant
- ⚡ Zero-copy workflow - no manual clipboard operations needed
- 🔧 Single binary with both server and client functionality
- 🗂️ Multiple message accumulation with deduplication

## Installation

### 1. Install the tool

```bash
go install github.com/cappyzawa/lazygit-mcp-bridge/cmd/lazygit-mcp-bridge@latest
```

### 2. Configure Claude Code

Add to your Claude Code MCP settings:
```json
"mcpServers": {
  "lazygit-mcp-bridge": {
    "command": "lazygit-mcp-bridge",
    "args": ["server"]
  }
}
```

Or use the CLI:
```bash
claude mcp add lazygit-mcp-bridge "lazygit-mcp-bridge server"
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
    lazygit-mcp-bridge send \
      --file "{{ .SelectedPath }}" \
      --line "{{ .SelectedLine }}" \
      --comment "{{ index .PromptResponses 0 }}"
```

That's it! No shell scripts needed anymore. The `lazygit-mcp-bridge` binary handles everything.

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
2. The `send` subcommand creates a JSON message file
3. The MCP server (running as `server` subcommand) watches for file changes
4. When detected, it queues messages with deduplication
5. Your AI assistant retrieves all accumulated messages via MCP tools

## Command Line Interface

```bash
# Run as MCP server
lazygit-mcp-bridge server

# Send a message from lazygit (usually called automatically)
lazygit-mcp-bridge send --file main.go --line 42 --comment "Add error handling"

# Show help
lazygit-mcp-bridge --help
lazygit-mcp-bridge server --help
lazygit-mcp-bridge send --help
```

## Multiple Message Support

The tool supports multiple message accumulation:

- Messages no longer overwrite each other
- Up to 10 messages are retained in memory
- SHA-256 hash-based deduplication prevents duplicates
- All messages delivered together when requested
- Clear separation between messages in the response

## Development

```bash
# Clone the repository
git clone https://github.com/cappyzawa/lazygit-mcp-bridge
cd lazygit-mcp-bridge

# Install dependencies
go mod tidy

# Build with make
make build

# Run server locally
make run-server

# Install to GOPATH/bin
make install

# Run tests
make test
```

## Project Structure

```
lazygit-mcp-bridge/
├── cmd/lazygit-mcp-bridge/    # CLI entry point
│   └── main.go                 # Cobra command definitions
├── internal/                   # Internal packages
│   ├── server/                 # MCP server implementation
│   │   └── server.go
│   └── client/                 # Send command implementation
│       └── client.go
├── docs/                       # Documentation
├── Makefile                    # Build automation
├── go.mod                      # Go modules
└── README.md
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