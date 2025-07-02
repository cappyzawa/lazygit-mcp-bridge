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
         │ 2. Execute 'send' command   │                             │
         │ which writes JSON to        │                             │
         │ ~/.config/.../mcp-          │                             │
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
- **Direct Integration**: `lazygit-mcp-bridge send` command (no shell scripts needed)

### 2. Message Format

**Input Format (from lazygit):**
```json
{
  "file": "path/to/file.go",
  "comment": "User's comment about the code (including line info if needed)",
  "project_root": "/path/to/project",
  "time": "2025-07-01T22:45:00Z"
}
```

**Internal Storage Format:**
```json
{
  "file": "path/to/file.go",
  "comment": "User's comment about the code (including line info if needed)",
  "project_root": "/path/to/project",
  "time": "2025-07-01T22:45:00Z",
  "hash": "sha256_hash_for_deduplication"
}
```

### 3. MCP Bridge Application

#### Subcommand Architecture

The application provides two main subcommands:

1. **`server`**: Runs as MCP server
   - Watches for message file changes
   - Implements MCP protocol
   - Manages message queue with deduplication
   
2. **`send`**: Client for sending messages
   - Replaces shell script functionality
   - Creates JSON message file
   - Validates input parameters

#### Technical Details

- **Language**: Go
- **CLI Framework**: Cobra for command-line interface
- **File Watching**: Uses `fsnotify` to monitor message file
- **Message Queue**: In-memory array for multiple pending messages
- **Deduplication**: SHA-256 hash-based duplicate message prevention
- **Message Limit**: Retains up to 10 most recent messages
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
3. **Message Creation**: `lazygit-mcp-bridge send` command creates/overwrites JSON message file
4. **File Detection**: MCP server detects file write via `fsnotify`
5. **Message Processing**: Server reads, validates, and stores message with deduplication
6. **Message Accumulation**: Multiple messages are stored in memory array (up to 10)
7. **AI Retrieval**: AI assistant calls `check_lazygit_messages` tool
8. **Batch Delivery**: Server returns all accumulated messages with clear separation
9. **Cleanup**: Message queue and file are cleared after successful retrieval

## Security Considerations

- Messages are stored temporarily in user's config directory
- No network communication required between lazygit and MCP server
- File permissions follow system defaults
- Messages are deleted only after successful retrieval by AI assistant
- Duplicate messages are filtered using SHA-256 hashing
- Message retention is limited to 10 most recent items

## Future Enhancements

- [x] Multiple message accumulation and batch delivery
- [x] Message deduplication system
- [ ] Support for notifications/push messaging  
- [ ] Multiple message types (error, warning, info)
- [ ] Rich diff context inclusion
- [ ] Support for other git tools beyond lazygit
- [ ] Configurable message retention limits