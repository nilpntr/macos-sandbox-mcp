package main

import (
	"context"
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed profiles/*
var profilesFS embed.FS

const version = "dev"

type BashSandboxedRequest struct {
	Command        string  `json:"command" jsonschema:"required,description=Shell command to run"`
	Profile        *string `json:"profile,omitempty" jsonschema:"description=Preset name,enum=default,enum=no-network,enum=readonly,enum=isolated,enum=network"`
	WorkingDir     *string `json:"working_dir,omitempty" jsonschema:"description=Working directory for the command"`
	TimeoutSeconds *int    `json:"timeout_seconds,omitempty" jsonschema:"description=Timeout in seconds,default=30"`
}

type BashSandboxedResponse struct {
	Stdout   string `json:"stdout" jsonschema:"description=Standard output from the command"`
	Stderr   string `json:"stderr" jsonschema:"description=Standard error from the command"`
	ExitCode int    `json:"exit_code" jsonschema:"description=Exit code of the command"`
}

func main() {
	s := server.NewMCPServer(
		"MacOS Sandbox MCP",
		version,
		server.WithToolCapabilities(false))

	t := mcp.NewTool("bash_sandboxed",
		mcp.WithDescription("Run a command in a MacOS sandbox"),
		mcp.WithInputSchema[BashSandboxedRequest](),
		mcp.WithOutputSchema[BashSandboxedResponse]())

	s.AddTool(t, mcp.NewStructuredToolHandler(getBashSandboxedHandler))

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}

func getBashSandboxedHandler(ctx context.Context, _ mcp.CallToolRequest, args BashSandboxedRequest) (BashSandboxedResponse, error) {
	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		return BashSandboxedResponse{}, fmt.Errorf("this tool only works on macOS")
	}

	// Set default profile if not specified
	profile := "default"
	if args.Profile != nil {
		profile = *args.Profile
	}

	// Set default timeout
	timeout := 30
	if args.TimeoutSeconds != nil {
		timeout = *args.TimeoutSeconds
	}

	// Create temporary directory for profile
	tempDir, err := os.MkdirTemp("", "sandbox-profiles-")
	if err != nil {
		return BashSandboxedResponse{}, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write profile to temp file
	profilePath := filepath.Join(tempDir, profile+".sb")
	profileContent, err := profilesFS.ReadFile("profiles/" + profile + ".sb")
	if err != nil {
		return BashSandboxedResponse{}, fmt.Errorf("profile '%s' not found", profile)
	}

	if err = os.WriteFile(profilePath, profileContent, 0644); err != nil {
		return BashSandboxedResponse{}, fmt.Errorf("failed to write profile: %w", err)
	}

	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Build sandbox-exec command
	cmd := exec.CommandContext(timeoutCtx, "sandbox-exec", "-f", profilePath, "sh", "-c", args.Command)

	// Set working directory if specified
	if args.WorkingDir != nil {
		cmd.Dir = *args.WorkingDir
	}

	// Execute command
	stdout, stderr, exitCode := executeCommand(cmd)

	return BashSandboxedResponse{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

func executeCommand(cmd *exec.Cmd) (string, string, int) {
	var stdout, stderr []byte
	var err error

	stdout, err = cmd.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			stderr = exitError.Stderr
			return string(stdout), string(stderr), exitError.ExitCode()
		}
		// Handle other types of errors (like timeout)
		return string(stdout), err.Error(), 1
	}

	return string(stdout), string(stderr), 0
}
