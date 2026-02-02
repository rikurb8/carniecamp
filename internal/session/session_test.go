package session

import (
	"testing"
)

func TestParseTool(t *testing.T) {
	tests := []struct {
		input    string
		expected Tool
	}{
		{"claude", ToolClaude},
		{"opencode", ToolOpencode},
		{"", ToolClaude},
		{"unknown", ToolClaude},
		{"CLAUDE", ToolClaude}, // case sensitive, falls back to default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseTool(tt.input)
			if got != tt.expected {
				t.Errorf("ParseTool(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestBuildClaudeArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		wantArgs []string
	}{
		{
			name:     "no options",
			opts:     Options{Tool: ToolClaude},
			wantArgs: []string{"claude"},
		},
		{
			name: "with prompt non-interactive",
			opts: Options{
				Tool:   ToolClaude,
				Prompt: "hello world",
			},
			wantArgs: []string{"claude", "--print", "hello world"},
		},
		{
			name: "with prompt interactive",
			opts: Options{
				Tool:        ToolClaude,
				Prompt:      "hello world",
				Interactive: true,
			},
			wantArgs: []string{"claude"},
		},
		{
			name: "with system prompt",
			opts: Options{
				Tool:         ToolClaude,
				SystemPrompt: "You are a helpful assistant",
			},
			wantArgs: []string{"claude", "--system-prompt", "You are a helpful assistant"},
		},
		{
			name: "with both prompts",
			opts: Options{
				Tool:         ToolClaude,
				SystemPrompt: "system",
				Prompt:       "user",
			},
			wantArgs: []string{"claude", "--system-prompt", "system", "--print", "user"},
		},
		{
			name: "with both prompts interactive",
			opts: Options{
				Tool:         ToolClaude,
				SystemPrompt: "system",
				Prompt:       "user",
				Interactive:  true,
			},
			wantArgs: []string{"claude", "--system-prompt", "system"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs := buildClaudeArgs(tt.opts)
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("got %d args %v, want %d args %v", len(gotArgs), gotArgs, len(tt.wantArgs), tt.wantArgs)
				return
			}
			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildOpencodeArgs(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		wantArgs []string
	}{
		{
			name:     "no options",
			opts:     Options{Tool: ToolOpencode},
			wantArgs: []string{"opencode"},
		},
		{
			name: "with prompt non-interactive",
			opts: Options{
				Tool:   ToolOpencode,
				Prompt: "hello world",
			},
			wantArgs: []string{"opencode", "-p", "hello world"},
		},
		{
			name: "with prompt interactive",
			opts: Options{
				Tool:        ToolOpencode,
				Prompt:      "hello world",
				Interactive: true,
			},
			wantArgs: []string{"opencode"},
		},
		{
			name: "with system prompt",
			opts: Options{
				Tool:         ToolOpencode,
				SystemPrompt: "system",
			},
			wantArgs: []string{"opencode", "--prompt", "system"},
		},
		{
			name: "with system prompt interactive",
			opts: Options{
				Tool:         ToolOpencode,
				SystemPrompt: "system",
				Prompt:       "hello world",
				Interactive:  true,
			},
			wantArgs: []string{"opencode", "--prompt", "system"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotArgs := buildOpencodeArgs(tt.opts)
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("got %d args %v, want %d args %v", len(gotArgs), gotArgs, len(tt.wantArgs), tt.wantArgs)
				return
			}
			for i, arg := range gotArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildToolCommand(t *testing.T) {
	tests := []struct {
		name      string
		opts      Options
		wantFirst string
	}{
		{
			name:      "claude tool",
			opts:      Options{Tool: ToolClaude},
			wantFirst: "claude",
		},
		{
			name:      "opencode tool",
			opts:      Options{Tool: ToolOpencode},
			wantFirst: "opencode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildToolCommand(tt.opts)
			if len(got) == 0 {
				t.Error("expected non-empty command")
				return
			}
			if got[0] != tt.wantFirst {
				t.Errorf("first arg = %q, want %q", got[0], tt.wantFirst)
			}
		})
	}
}

func TestCommand(t *testing.T) {
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{
			name: "claude interactive",
			opts: Options{
				Tool:         ToolClaude,
				Model:        "openai/gpt-5.2-codex",
				SystemPrompt: "system",
				Prompt:       "user",
				Interactive:  true,
			},
			want: `claude --model openai/gpt-5.2-codex --system-prompt system`,
		},
		{
			name: "opencode non-interactive",
			opts: Options{
				Tool:         ToolOpencode,
				Model:        "openai/gpt-5.2-codex",
				SystemPrompt: "system",
				Prompt:       "hello world",
			},
			want: `opencode --model openai/gpt-5.2-codex --prompt system -p $'hello world'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Command(tt.opts)
			if got != tt.want {
				t.Errorf("Command() = %q, want %q", got, tt.want)
			}
		})
	}
}
