# Mayor Command

The `mayor` command provides project oversight and planning tools for your Bordertown workspace.

## Quick Start

```bash
# Start planning a new epic
bt mayor new-epic

# Alias
bt mayor plan-epic

# Plan with an initial idea
bt mayor new-epic --title "User authentication"

# Review epics and planning status
bt mayor review

# Include closed epics and issues
bt mayor review --all
```

## Commands

### `mayor new-epic`

Spawns an interactive AI session in tmux to help you plan, refine, and create a new epic with tasks.

Alias: `mayor plan-epic`

**What it does:**

1. Creates a tmux session with the AI tool (claude or opencode)
2. Asks clarifying questions about scope, goals, and constraints
3. Helps break down the work into actionable tasks
4. Identifies dependencies between tasks
5. Suggests priorities (P0-P4)
6. Creates the epic and tasks using `bd` commands

**Flags:**

- `--title, -t` - Initial title or idea for the epic
- `--tool` - Override planning tool (`claude` or `opencode`)

**Tmux Session:**

The session runs in tmux, so you can:
- Detach with `Ctrl+B, D` and return later
- Reattach with `tmux attach -t bt-epic-planning`
- The session persists even if your terminal disconnects

**Requirements:**

- `tmux` must be installed
- `claude` or `opencode` must be installed

**Configuration (town.yml):**

```yaml
mayor:
  model: openai/gpt-5.2-codex
  planning_tool: opencode  # or claude
  planning_prompt_file: .bordertown/prompts/epic-planning.md  # optional custom prompt
```

Model values use the `<provider>/<model>` format (e.g., `openai/gpt-5.2-codex`).

**Custom Prompts:**

You can customize the planning prompt by creating a file at:
- Configured path in `town.yml` under `mayor.planning_prompt_file`
- Default location: `.bordertown/prompts/epic-planning.md`

If no custom prompt exists, the built-in default is used.

**Context Injection:**

The planning session automatically receives context about:
- Project name and description (from `town.yml`)
- Existing epics and their status (from beads)

This helps the AI make suggestions aligned with your existing work.

---

### `mayor review`

Analyzes your beads issues grouped by epic and indicates which epics need more planning.

**Output includes:**

- All open epics with their child tasks
- Task counts (open/closed) per epic
- Planning warnings for epics that need attention
- Orphan issues not linked to any epic

**Flags:**

- `--all` - Include closed epics and issues in the output

**Example output:**

```
Mayor Review

○ bt-abc - Feature Implementation
  Tasks: 2 open, 1 closed
    ○ bt-def - Implement core logic
    ○ bt-ghi - Add tests
    ● bt-jkl - Design API

○ bt-xyz - Another Epic
  Tasks: 0 open, 0 closed
  ⚠ Needs planning: no tasks defined, description is brief

Orphan Issues (no epic)
  ○ bt-123 [task] - Standalone task
```

## How Epic Grouping Works

Issues are grouped by epic using dependencies:

1. An issue belongs to an epic if **the epic depends on it** (the issue blocks the epic)
2. Issues not linked to any epic appear in the "Orphan Issues" section

To link a task to an epic:
```bash
bd dep add <epic-id> <task-id>
```

## Planning Indicators

The `review` command flags epics that may need more planning:

| Warning | Meaning |
|---------|---------|
| "no tasks defined" | Epic has no linked tasks |
| "only has 1-2 tasks" | Epic may be under-planned |
| "description is brief" | Epic description is less than 50 characters |
| "all tasks complete but epic still open" | Consider closing the epic |

## Symbols

- `○` Open issue
- `●` Closed issue
- `⚠` Warning/needs attention
