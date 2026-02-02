package session

import (
	"strconv"
	"strings"
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

// Options configures how a session command is built.
type Options struct {
	Tool         Tool
	Model        string // Model for the tool
	Prompt       string // Initial prompt/message to send
	SystemPrompt string // System prompt (for claude)
	Interactive  bool   // When true, do not send the prompt as a one-shot
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

// Command returns the command line to start the tool with the given options.
func Command(opts Options) string {
	toolCmd := buildToolCommand(opts)
	if len(toolCmd) == 0 {
		return ""
	}
	return formatCommand(toolCmd[0], toolCmd[1:])
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

// buildToolCommand returns the command and args to run the tool.
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
