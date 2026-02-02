# Operator Command

The `operator` command prints a ready-to-paste command for starting an operator planning session.
Alias: `op`

## Quick Start

```bash
carnie operator
```

Paste the printed command into your terminal to start the session.

## Behavior

- Builds a command for `opencode` or `claude` based on `camp.yml`
- Injects project context and the planning prompt
- Prints the command only (no tmux, no auto-spawn)

## Requirements

- `opencode` or `claude` must be installed

## Configuration (camp.yml)

```yaml
operator:
  model: openai/gpt-5.2-codex
  planning_tool: opencode  # or claude
  planning_prompt_file: .carnie/prompts/epic-planning.md  # optional custom prompt
```

Model values use the `<provider>/<model>` format (e.g., `openai/gpt-5.2-codex`).

## Custom Prompts

You can customize the planning prompt by creating a file at:
- Configured path in `camp.yml` under `operator.planning_prompt_file`
- Default location: `.carnie/prompts/epic-planning.md`

If no custom prompt exists, the built-in default is used.

## Context Injection

The planning session automatically receives context about:
- Project name and description (from `camp.yml`)
