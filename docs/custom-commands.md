# Custom Commands Documentation

This document explains how to use custom commands to efficiently integrate lazygit-mcp-bridge with Claude Code.

## Overview

Previously, when checking comments from lazygit in Claude, you had to manually say "I commented" each time. Custom commands allow you to efficiently check messages without depending on conversation context.

## Setup

### 1. Create Command Directory

```bash
mkdir -p /path/to/your/project/.claude/commands
```

### 2. Create Command Files

Create a `.claude/commands/` directory in your project root and place the following command files.

## Available Commands

### `/project:lg` - Quick Check

**File**: `.claude/commands/lg.md`

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

**Usage**:
```
/project:lg
/project:lg keep it concise
/project:lg focus on security
/project:lg apply Go best practices
```

## Benefits

### 1. Context Independence
- Immediately usable in new Claude sessions
- No dependency on long conversation history
- Managed as project-specific configuration

### 2. Efficiency
- No need for manual "I commented" notifications
- Instant checks with short command input
- Customizable through arguments
- **Multiple messages processed in single command**

### 3. Consistency
- Shareable commands across team members
- Version-controlled as project-specific settings
- Standardized workflow

### 4. Multiple Message Support (New!)
- **Batch Processing**: All accumulated comments processed together
- **No Message Loss**: Rapid commenting won't overwrite previous messages
- **Smart Deduplication**: Duplicate comments automatically filtered
- **Retention Management**: Up to 10 most recent messages maintained

## Workflow

### Traditional Workflow
1. Create comment in lazygit
2. Manually tell Claude "I commented"
3. Claude checks messages using MCP tool
4. Receive single message at a time
5. Repeat for each comment

### Custom Command Workflow (with Multiple Message Support)
1. Create multiple comments in lazygit (they accumulate automatically)
2. Execute `/project:lg` in Claude
3. Instantly receive all accumulated messages with clear separation
4. All comments processed in single response

## Advanced Usage Examples

### Analysis from Specific Perspectives

```
/project:lg analyze from security perspective
/project:lg focus on performance optimization
/project:lg emphasize code readability
```

### Language-Specific Analysis

```
/project:lg follow Go idioms
/project:lg apply Rust best practices
/project:lg prioritize TypeScript type safety
```

### Multiple Message Scenarios

```
# After commenting on 3 different files
/project:lg provide comprehensive review of all changes

# For batch refactoring
/project:lg focus on consistent patterns across all files

# Code review preparation
/project:lg prepare summary for pull request
```

## Troubleshooting

### Command Not Recognized
- Verify `.claude/commands/` directory exists in project root
- Check that `.md` file YAML frontmatter is correctly written
- Restart Claude Code to reload commands

### MCP Tool Unavailable
- Confirm `lazygit-mcp-bridge` is running properly
- Verify lazygit-mcp-bridge is enabled in Claude Code MCP settings

## Customization

### Creating Custom Commands

Create your own commands based on project-specific needs:

```markdown
---
allowed-tools: mcp__lazygit-mcp-bridge__check_lazygit_messages
description: Custom review command for specific project needs
---

# Custom Code Review

Your custom review logic here.

Context: $ARGUMENTS
```

### Combining Multiple MCP Tools

```markdown
---
allowed-tools: mcp__lazygit-mcp-bridge__check_lazygit_messages, mcp__github__get_pull_request
description: Combined review with lazygit and GitHub PR context
---

# Enhanced Code Review

Combine lazygit comments with GitHub PR context for comprehensive review.
```

## Summary

Custom commands significantly improve the lazygit-mcp-bridge user experience, particularly:

- **Efficiency**: No manual notifications required
- **Reusability**: Usable in any Claude session
- **Consistency**: Standardized workflow across teams
- **Multiple Message Support**: Batch processing of accumulated comments
- **No Data Loss**: Robust message accumulation prevents comment overwrites

### Key Improvements in v2.0

1. **Multiple Message Accumulation**: Comments no longer overwrite each other
2. **Smart Deduplication**: Prevents duplicate message processing
3. **Batch Processing**: All messages delivered together with clear separation
4. **Retention Management**: Automatic cleanup with configurable limits

Use these commands to achieve more efficient code review and improvement cycles with comprehensive multi-file analysis capabilities.