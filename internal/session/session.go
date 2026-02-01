package session

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Tool represents the CLI tool to use for sessions.
type Tool string

const (
	ToolClaude   Tool = "claude"
	ToolOpencode Tool = "opencode"
)

// ParseTool converts a string to a Tool, defaulting to Claude.
func ParseTool(s string) Tool {
	switch s {
	case "opencode":
		return ToolOpencode
	case "claude", "":
		return ToolClaude
	default:
		return ToolClaude
	}
}

// Options configures how a session is spawned.
type Options struct {
	Tool         Tool
	Model        string // Model for the tool
	Prompt       string // Initial prompt/message to send
	SystemPrompt string // System prompt (for claude)
	WorkDir      string // Working directory
	Interactive  bool   // Attach to tmux session for interactive use
	SessionName  string // Tmux session name (auto-generated if empty)
}

// Spawn starts a new CLI session in tmux with the given options.
// If Interactive is true, attaches to the tmux session.
// Returns when the user detaches or the session exits.
func Spawn(opts Options) error {
	normalizedOpts, err := normalizeOptions(opts)
	if err != nil {
		return err
	}

	// Check tmux is available
	if !TmuxAvailable() {
		return fmt.Errorf("tmux is not installed or not in PATH")
	}

	createArgs := buildCreateArgs(normalizedOpts)

	createCmd := exec.Command("tmux", createArgs...)
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("create tmux session: %w", err)
	}

	if normalizedOpts.Interactive && normalizedOpts.Prompt != "" {
		// Give the tool a moment to start before sending input.
		time.Sleep(200 * time.Millisecond)
		if err := sendInitialPrompt(normalizedOpts.SessionName, normalizedOpts.Prompt); err != nil {
			return fmt.Errorf("send initial prompt: %w", err)
		}
	}

	if normalizedOpts.Interactive {
		// Attach to the session
		attachCmd := exec.Command("tmux", "attach-session", "-t", normalizedOpts.SessionName)
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr

		if err := attachCmd.Run(); err != nil {
			// Session may have ended, which is fine
			if !strings.Contains(err.Error(), "exit status") {
				return fmt.Errorf("attach to session: %w", err)
			}
		}
	}

	return nil
}

// SpawnCommand returns the tmux command used to create the session.
func SpawnCommand(opts Options) (string, error) {
	normalizedOpts, err := normalizeOptions(opts)
	if err != nil {
		return "", err
	}

	createArgs := buildCreateArgs(normalizedOpts)
	return formatCommand("tmux", createArgs), nil
}

func normalizeOptions(opts Options) (Options, error) {
	if opts.WorkDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return Options{}, fmt.Errorf("get working directory: %w", err)
		}
		opts.WorkDir = cwd
	}

	if opts.SessionName == "" {
		opts.SessionName = fmt.Sprintf("cn-%s-%d", opts.Tool, time.Now().Unix())
	}

	return opts, nil
}

func buildCreateArgs(opts Options) []string {
	toolCmd := buildToolCommand(opts)
	createArgs := []string{
		"new-session",
		"-d",                   // detached
		"-s", opts.SessionName, // session name
		"-c", opts.WorkDir, // working directory
	}

	return append(createArgs, toolCmd...)
}

// NormalizeModel ensures a model string is compatible with the tool.
func NormalizeModel(tool Tool, model string) string {
	if model == "" {
		return ""
	}

	if strings.Contains(model, "/") {
		return model
	}

	if tool == ToolOpencode {
		return "openai/" + model
	}

	return model
}

func formatCommand(name string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteArg(name))
	for _, arg := range args {
		parts = append(parts, quoteArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteArg(arg string) string {
	if arg == "" {
		return `""`
	}

	for _, r := range arg {
		switch r {
		case ' ', '\t', '\n', '"', '\'', '\\', '$', '`':
			return strconv.Quote(arg)
		}
	}

	return arg
}

// buildToolCommand returns the command and args to run inside tmux.
func buildToolCommand(opts Options) []string {
	switch opts.Tool {
	case ToolOpencode:
		return buildOpencodeArgs(opts)
	case ToolClaude:
		return buildClaudeArgs(opts)
	default:
		return []string{string(opts.Tool)}
	}
}

// buildClaudeArgs returns args for claude command.
func buildClaudeArgs(opts Options) []string {
	args := []string{"claude"}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", opts.SystemPrompt)
	}

	if opts.Prompt != "" && !opts.Interactive {
		args = append(args, "--print", opts.Prompt)
	}

	return args
}

// buildOpencodeArgs returns args for opencode command.
func buildOpencodeArgs(opts Options) []string {
	args := []string{"opencode"}

	if opts.Model != "" {
		args = append(args, "--model", opts.Model)
	}

	if opts.SystemPrompt != "" {
		args = append(args, "--prompt", opts.SystemPrompt)
	}

	if opts.Prompt != "" && !opts.Interactive {
		args = append(args, "-p", opts.Prompt)
	}

	return args
}

func sendInitialPrompt(sessionName, prompt string) error {
	cmd := exec.Command("tmux", "send-keys", "-t", sessionName, prompt, "C-m")
	return cmd.Run()
}

// Available checks if the specified tool is available on the system.
func Available(tool Tool) bool {
	var name string
	switch tool {
	case ToolClaude:
		name = "claude"
	case ToolOpencode:
		name = "opencode"
	default:
		return false
	}

	_, err := exec.LookPath(name)
	return err == nil
}

// TmuxAvailable checks if tmux is available on the system.
func TmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

// SessionExists checks if a tmux session with the given name exists.
func SessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	return cmd.Run() == nil
}

// KillSession kills a tmux session by name.
func KillSession(name string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", name)
	return cmd.Run()
}
