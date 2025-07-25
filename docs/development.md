# Development Guide

## Project Structure

```
lazygit-mcp-bridge/
├── cmd/lazygit-mcp-bridge/    # CLI entry point
│   └── main.go                # Cobra command definitions
├── internal/                  # Internal packages
│   ├── server/                # MCP server implementation
│   │   └── server.go
│   └── client/                # Send command implementation
│       └── client.go
├── docs/                      # Documentation
│   ├── architecture.md        # System architecture
│   ├── mcp-protocol.md        # Protocol specification
│   ├── custom-commands.md     # Custom commands guide
│   └── development.md         # This file
├── Makefile                   # Build automation
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # User documentation
└── .gitignore                 # Git ignore patterns
```

## Development Setup

### Prerequisites

- Go 1.24 or later
- Git
- lazygit (for testing)
- Claude Code or other MCP-compatible AI assistant

### Building from Source

```bash
# Clone the repository
git clone https://github.com/cappyzawa/lazygit-mcp-bridge
cd lazygit-mcp-bridge

# Install dependencies
go mod download

# Build the binary
make build

# Or build manually
go build -o build/lazygit-mcp-bridge cmd/lazygit-mcp-bridge/main.go

# Install to $GOPATH/bin
make install

# Or install manually
go install ./cmd/lazygit-mcp-bridge
```

## Code Structure

### Main Components

#### CLI Entry Point (cmd/lazygit-mcp-bridge/main.go)

```go
// Root command
var rootCmd = &cobra.Command{
    Use:   "lazygit-mcp-bridge",
    Short: "Bridge between lazygit and AI assistants using MCP",
}

// Server subcommand
var serverCmd = &cobra.Command{
    Use:   "server",
    Short: "Run as MCP server",
    RunE: func(cmd *cobra.Command, args []string) error {
        return server.Run()
    },
}

// Send subcommand
var sendCmd = &cobra.Command{
    Use:   "send",
    Short: "Send a message from lazygit to AI",
    RunE: func(cmd *cobra.Command, args []string) error {
        return client.Send(file, comment)
    },
}
```

#### Server Package (internal/server/server.go)

```go
// MCP Request/Response types
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

// Message structure for storage with deduplication
type LazygitMessage struct {
    File        string `json:"file"`
    Comment     string `json:"comment"`
    ProjectRoot string `json:"project_root"`
    Time        string `json:"time"`
    Hash        string `json:"hash"` // SHA-256 for deduplication
}

// Main server function
func Run() error { ... }
```

#### Client Package (internal/client/client.go)

```go
// Send message to the MCP server
func Send(file, comment string) error {
    // Create message
    // Write to JSON file
    // Return success/error
}
```

### Key Functions Flow

```
main()
  ├── rootCmd.Execute()
  │   ├── serverCmd
  │   │   └── server.Run()
  │   │       ├── watchMessageFile() [goroutine]
  │   │       │   └── readMessageFile()
  │   │       └── handleMCPRequest() [main loop]
  │   │           ├── handleInitialize()
  │   │           ├── handleResourcesList()
  │   │           ├── handleResourcesRead()
  │   │           ├── handleToolsList()
  │   │           └── handleToolsCall()
  │   └── sendCmd
  │       └── client.Send()
  │           ├── getConfigDir()
  │           ├── getCurrentProjectRoot()
  │           └── os.WriteFile()
```

## Build System

### Makefile Targets

The project includes a Makefile for common development tasks:

```bash
# Build the binary
make build

# Clean build artifacts
make clean

# Install to GOPATH/bin
make install

# Run tests
make test

# Run server for development
make run-server
```

### Manual Build

If you prefer not to use make:

```bash
# Build
go build -o build/lazygit-mcp-bridge cmd/lazygit-mcp-bridge/main.go

# Install
go install ./cmd/lazygit-mcp-bridge
```

## Testing

### Manual Testing

1. **Test server mode:**
```bash
# Terminal 1: Run the MCP server
./build/lazygit-mcp-bridge server

# Terminal 2: Test the send command
./build/lazygit-mcp-bridge send --file test.go --comment "Add error handling for line 42"
```

2. **Test MCP protocol:**
```bash
# Send initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./build/lazygit-mcp-bridge server
```

3. **Test with lazygit:**
```bash
# Make sure the binary is in PATH
make install

# Open lazygit and press Ctrl+Y
```

### Integration Testing

```bash
# Start Claude Code with the MCP server
claude mcp add test-bridge "lazygit-mcp-bridge server"
claude

# In lazygit, press Ctrl+Y and enter a comment
# Check if Claude receives the message
```

## Message Processing Features

### Multiple Message Support

The system now supports accumulating multiple messages before delivery:

```go
// Message accumulation with deduplication
func readMessageFile() {
    // Read raw message from file
    var rawMessage struct { ... }
    
    // Create hash for deduplication
    hashInput := strings.Join([]string{
        rawMessage.File, rawMessage.Line, 
        rawMessage.Comment, rawMessage.Time}, "|")
    hash := sha256.Sum256([]byte(hashInput))
    hashString := hex.EncodeToString(hash[:])
    
    // Check for duplicates
    for _, existingMsg := range messageQueue {
        if existingMsg.Hash == hashString {
            return // Skip duplicate
        }
    }
    
    // Add to queue with retention limit
    messageQueue = append(messageQueue, message)
    if len(messageQueue) > 10 {
        messageQueue = messageQueue[1:] // Keep last 10
    }
}
```

### Deduplication Algorithm

1. **Hash Generation**: SHA-256 of file + comment + time
2. **Duplicate Check**: Compare against existing message hashes
3. **Skip Processing**: Ignore messages with existing hashes
4. **Memory Efficiency**: Only store hash string, not full message content

### Batch Message Delivery

```go
func handleToolsCall(req MCPRequest) {
    if name == "check_lazygit_messages" {
        if len(messageQueue) > 0 {
            // Format all messages with separators
            var allMessages []string
            for i, msg := range messageQueue {
                formattedMessage := fmt.Sprintf(
                    "Message %d:\nFile: %s\nLine: %s\nComment: %s\nTime: %s\n\nPlease improve this code.",
                    i+1, msg.File, msg.Line, msg.Comment, msg.Time)
                allMessages = append(allMessages, formattedMessage)
            }
            
            // Join with clear separators
            finalMessage := strings.Join(allMessages, "\n" + strings.Repeat("-", 50) + "\n\n")
            
            // Clear queue after successful delivery
            messageQueue = []LazygitMessage{}
            os.Remove(messageFile)
        }
    }
}
```

## Adding New Features

### 1. Adding a New Tool

```go
// In handleToolsList(), add:
{
    Name:        "your_new_tool",
    Description: "Description of the tool",
    InputSchema: ToolSchema{
        Type: "object",
        Properties: map[string]interface{}{
            "param1": map[string]string{
                "type": "string",
                "description": "Parameter description",
            },
        },
        Required: []string{"param1"},
    },
}

// In handleToolsCall(), add:
case "your_new_tool":
    // Handle the tool call
    params := req.Params.(map[string]interface{})
    // Process and return result
```

### 2. Adding Notifications (Future)

```go
// Send notification when new message arrives
func sendNotification(message string) {
    notification := MCPNotification{
        JSONRPC: "2.0",
        Method:  "notifications/message",
        Params: map[string]interface{}{
            "level": "info",
            "message": "New message from lazygit",
            "data": map[string]string{
                "content": message,
            },
        },
    }
    data, _ := json.Marshal(notification)
    fmt.Println(string(data))
}
```

## Debugging

### Enable Debug Logging

```go
// Add debug flag
var debug = os.Getenv("DEBUG") == "1"

func debugLog(format string, args ...interface{}) {
    if debug {
        log.Printf("[DEBUG] "+format, args...)
    }
}
```

### Common Issues

1. **File not found:**
   - Check if `~/.config/jesseduffield/lazygit/` exists
   - Verify file permissions

2. **MCP server not responding:**
   - Check if the binary is in PATH
   - Verify JSON-RPC format

3. **Messages not received:**
   - Check file watcher is running
   - Verify message file format

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure code follows Go conventions
5. Submit a pull request

### Code Style

- Use `gofmt` for formatting
- Follow Go naming conventions
- Add comments for exported functions
- Keep functions focused and small

### Commit Messages

```
feat: add notification support
fix: handle empty message queue
docs: update protocol specification
refactor: simplify file watching logic
```

## Release Process

1. Update version in `main.go`
2. Update CHANGELOG.md
3. Create git tag: `git tag v1.0.0`
4. Push tag: `git push origin v1.0.0`
5. GitHub Actions will create release

## Resources

- [MCP Specification](https://github.com/anthropics/mcp)
- [Go fsnotify](https://github.com/fsnotify/fsnotify)
- [JSON-RPC 2.0](https://www.jsonrpc.org/specification)