# Architecture

## Overview

`lazygit-mcp-bridge` acts as a bridge between lazygit and AI assistants using the Model Context Protocol (MCP).

## System Architecture

```
┌─────────────────┐          ┌─────────────────┐          ┌─────────────────┐
│                 │          │                 │          │                 │
│    lazygit      │          │ MCP Bridge      │          │  AI Assistant   │
│                 │          │                 │          │  (Claude, etc)  │
│                 │          │                 │          │                 │
└────────┬────────┘          └────────┬────────┘          └────────┬────────┘
         │                             │                             │
         │ 1. User presses Ctrl+Y      │                             │
         │ and enters comment          │                             │
         │                             │                             │
         ├─────────────────────────────►                             │
         │ 2. Write JSON message       │                             │
         │ to ~/.config/.../claude-    │                             │
         │ messages.json               │                             │
         │                             │                             │
         │                             │ 3. File watcher detects     │
         │                             │ change and reads message    │
         │                             │                             │
         │                             │◄────────────────────────────┤
         │                             │ 4. AI calls check_lazygit_  │
         │                             │ messages tool               │
         │                             │                             │
         │                             ├─────────────────────────────►
         │                             │ 5. Return formatted message │
         │                             │                             │
```

## Component Details

### 1. Lazygit Integration

- **Custom Command**: Configured in `~/.config/jesseduffield/lazygit/config.yml`
- **Key Binding**: `Ctrl+Y` in staging context
- **Script**: `send-to-ai.sh` writes JSON messages

### 2. Message Format

```json
{
  "file": "path/to/file.go",
  "line": "42",
  "comment": "User's comment about the code",
  "time": "2025-07-01T22:45:00Z"
}
```

### 3. MCP Bridge Server

- **Language**: Go
- **File Watching**: Uses `fsnotify` to monitor message file
- **Message Queue**: In-memory queue for pending messages
- **Protocol**: Implements MCP 2024-11-05 specification

### 4. MCP Protocol Implementation

#### Supported Methods

- `initialize`: Server initialization
- `resources/list`: List available resources
- `resources/read`: Read message resources
- `tools/list`: List available tools
- `tools/call`: Execute tools (check_lazygit_messages)

#### Capabilities

```json
{
  "resources": {
    "subscribe": true,
    "listChanged": true
  },
  "tools": {
    "listChanged": false
  }
}
```

## Data Flow

1. **User Action**: User reviews code in lazygit and presses `Ctrl+Y`
2. **Comment Input**: User enters a comment about the code
3. **Message Creation**: Shell script creates JSON message file
4. **File Detection**: MCP server detects file creation via `fsnotify`
5. **Message Queuing**: Server reads and queues the message
6. **AI Retrieval**: AI assistant calls `check_lazygit_messages` tool
7. **Message Delivery**: Server returns formatted message to AI
8. **Cleanup**: Message file is deleted after reading

## Security Considerations

- Messages are stored temporarily in user's config directory
- No network communication required between lazygit and MCP server
- File permissions follow system defaults
- Messages are deleted immediately after reading

## Future Enhancements

- [ ] Support for notifications/push messaging
- [ ] Multiple message types (error, warning, info)
- [ ] Rich diff context inclusion
- [ ] Support for other git tools beyond lazygit