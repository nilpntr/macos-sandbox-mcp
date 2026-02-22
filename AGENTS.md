# macos-sandbox-mcp

An MCP server that exposes a `bash_sandboxed` tool, wrapping commands in macOS `sandbox-exec` for safe, restricted execution. Intended for use with Claude Code, Cursor, OpenCode, or any MCP-compatible client.

## Project Structure

```
macos-sandbox-mcp/
├── main.go          # Entry point, MCP server setup
├── server.go        # Tool registration and request handling
├── sandbox.go       # sandbox-exec invocation logic
├── profiles/        # Built-in .sb profile presets
│   ├── default.sb
│   ├── no-network.sb
│   ├── readonly.sb
│   └── isolated.sb
└── AGENTS.md
```

## Tech Stack

- **Language**: Go
- **MCP SDK**: `github.com/mark3labs/mcp-go` (use this, it's the most mature Go MCP library)
- **Transport**: stdio (standard for local MCP servers)
- **Target OS**: macOS only — make this clear in errors if run elsewhere

## Tool Specification

### `bash_sandboxed`

Runs a shell command inside a `sandbox-exec` sandbox.

**Input schema:**
```json
{
  "command": "string (required) — shell command to run",
  "profile": "string (optional) — preset name: default | no-network | readonly | isolated",
  "working_dir": "string (optional) — working directory for the command",
  "timeout_seconds": "number (optional, default 30)"
}
```

**Output:**
```json
{
  "stdout": "...",
  "stderr": "...",
  "exit_code": 0
}
```

## Sandbox Profiles

### `default.sb`
Allow most things, deny network. Safe baseline for running scripts.
```scheme
(version 1)
(allow default)
(deny network*)
```

### `no-network.sb`
Same as default. Explicit preset for clarity.

### `readonly.sb`
Allow reads, deny all writes and network.
```scheme
(version 1)
(deny default)
(allow file-read*)
(allow process-exec)
(allow sysctl-read)
```

### `isolated.sb`
Maximum restriction. Deny everything except process execution basics.
```scheme
(version 1)
(deny default)
(allow process-fork)
(allow process-exec)
(allow file-read-data (regex "^/usr/lib"))
(allow file-read-data (regex "^/System/Library"))
```

## Implementation Notes

- Use `os/exec` to invoke `sandbox-exec -f <profile_path> sh -c <command>`
- Write built-in profiles to a temp dir on startup, clean up on exit
- If `working_dir` is set, use `cmd.Dir`
- Wrap execution in a context with the specified timeout
- Check `runtime.GOOS == "darwin"` on startup and return a clear error if not macOS
- Return non-zero exit codes as part of the response (not as MCP errors) — the caller should decide what to do with them
- Keep the binary self-contained: embed profiles using `//go:embed profiles/*`

## Build & Install

```bash
go build -o macos-sandbox-mcp .
```

Users add to their MCP config:
```json
{
  "mcpServers": {
    "sandbox": {
      "command": "/path/to/macos-sandbox-mcp"
    }
  }
}
```

## Out of Scope

- Dynamic profile generation from natural language
- GUI or web interface
- Windows/Linux support
- Profile validation or linting