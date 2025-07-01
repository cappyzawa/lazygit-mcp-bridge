# MCP Protocol Specification

## Overview

This document describes how `lazygit-mcp-bridge` implements the Model Context Protocol (MCP).

## Protocol Version

- **Version**: 2024-11-05
- **JSON-RPC**: 2.0

## Message Flow Diagram

```
AI Assistant                    MCP Bridge
     │                              │
     ├──────────────────────────────►
     │    1. initialize             │
     │                              │
     ◄──────────────────────────────┤
     │    2. initialize result      │
     │                              │
     ├──────────────────────────────►
     │    3. initialized            │
     │                              │
     ├──────────────────────────────►
     │    4. tools/list             │
     │                              │
     ◄──────────────────────────────┤
     │    5. available tools        │
     │                              │
     ├──────────────────────────────►
     │    6. tools/call             │
     │    (check_lazygit_messages)  │
     │                              │
     ◄──────────────────────────────┤
     │    7. message content        │
     │                              │
```

## Request/Response Examples

### 1. Initialize

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {}
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "resources": {
        "subscribe": true,
        "listChanged": true
      },
      "tools": {
        "listChanged": false
      }
    },
    "serverInfo": {
      "name": "lazygit-mcp-bridge",
      "version": "1.0.0"
    }
  }
}
```

### 2. Tools List

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "check_lazygit_messages",
        "description": "Check for new messages from lazygit",
        "inputSchema": {
          "type": "object",
          "properties": {}
        }
      }
    ]
  }
}
```

### 3. Tool Call

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "check_lazygit_messages",
    "arguments": {}
  }
}
```

**Response (with message):**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "File: src/main.go\nLine: 42\nComment: This function needs error handling\n\nPlease improve this code."
      }
    ]
  }
}
```

**Response (no messages):**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "No new messages from lazygit"
      }
    ]
  }
}
```

## Resources

### Resource List

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "resources/list"
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "result": {
    "resources": [
      {
        "uri": "lazygit://messages",
        "name": "Lazygit Messages",
        "description": "Messages from lazygit for code improvement",
        "mimeType": "text/plain"
      }
    ]
  }
}
```

### Resource Read

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "resources/read",
  "params": {
    "uri": "lazygit://messages"
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "contents": [
      {
        "uri": "lazygit://messages",
        "mimeType": "text/plain",
        "text": "File: src/main.go\nLine: 42\nComment: Add validation here\n\nPlease improve this code."
      }
    ]
  }
}
```

## Error Handling

### Standard Error Response

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32601,
    "message": "Method not found"
  }
}
```

### Error Codes

- `-32700`: Parse error
- `-32600`: Invalid request
- `-32601`: Method not found
- `-32602`: Invalid params

## Future Protocol Extensions

### Planned: Notifications

```json
{
  "jsonrpc": "2.0",
  "method": "notifications/message",
  "params": {
    "level": "info",
    "message": "New message from lazygit",
    "data": {
      "file": "src/main.go",
      "line": "42"
    }
  }
}
```

### Planned: Subscriptions

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "resources/subscribe",
  "params": {
    "uri": "lazygit://messages"
  }
}
```