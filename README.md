# macOS Sandbox MCP

A Model Context Protocol (MCP) server that provides secure sandboxed command execution on macOS using Apple's built-in sandbox-exec functionality.

## Description

This MCP server allows AI assistants and other MCP clients to execute shell commands in a secure, sandboxed environment on macOS. It leverages Apple's sandbox-exec to restrict what commands can access, providing multiple security profiles for different use cases.

## Features

- **Multiple Security Profiles**: Choose from different sandbox configurations
  - `default`: Basic sandbox with no network access
  - `no-network`: Explicitly blocks network access
  - `readonly`: Read-only file system access
  - `isolated`: Maximum isolation
  - `network`: Allows network access for testing web requests
- **Configurable Timeouts**: Set custom timeout limits for command execution
- **Working Directory Support**: Execute commands in specific directories
- **Go Binary**: Single binary with embedded sandbox profiles

## Installation

### From Source

```bash
git clone https://github.com/nilpntr/macos-sandbox-mcp
cd macos-sandbox-mcp
go build -o macos-sandbox-mcp
```

### Using Go Install

```bash
go install github.com/nilpntr/macos-sandbox-mcp@latest
```

## Usage

### As MCP Server

Add to your MCP client configuration:

```json
{
  "mcpServers": {
    "macos-sandbox": {
      "command": "/path/to/macos-sandbox-mcp"
    }
  }
}
```

### Tool Parameters

- `command` (required): The shell command to execute
- `profile` (optional): Security profile (`default`, `no-network`, `readonly`, `isolated`, `network`)
- `working_dir` (optional): Working directory for command execution
- `timeout_seconds` (optional): Timeout in seconds (default: 30)

## Examples

```bash
# Basic command execution
{"command": "ls -la"}

# Network-enabled command
{"command": "curl -s google.com", "profile": "network"}

# Command with custom timeout and working directory
{"command": "find . -name '*.go'", "working_dir": "/Users/user/project", "timeout_seconds": 60}
```

## Security Profiles

### default
- Allows most standard operations
- Blocks network access
- Standard file system access

### no-network
- Explicitly denies all network operations
- Similar to default but more restrictive on network

### readonly
- Read-only file system access
- No write operations allowed
- No network access

### isolated
- Maximum security restrictions
- Minimal system access
- No network access

### network
- Allows network operations
- Useful for testing web requests and downloads
- Standard file system access

## Requirements

- macOS (uses Apple's sandbox-exec)
- Go 1.25+ (for building from source)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details
