# Carnie

```
🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪
🎠🤡🎈🎉🎭🎪🎟️🎡🎢🎠🤹‍♂️🎪🎭🎉🎈🤡🎠🎢🎡🎟️🎪🎭🎉🎈🤡🎠🎢🎡🎟️🎪
🎪                                                                🎪
🎪   🤹‍♀️🎉  W E L C O M E   T O   T H E   C A R N I V A L  🎉🤹‍ 🎪
🎪                                                                🎪
🎠🤡🎈🎉🎭🎪🎟️🎡🎢🎠🤹‍♂️🎪🎭🎉🎈🤡🎠🎢🎡🎟️🎪🎭🎉🎈🤡🎠🎢🎡🎟️🎪
🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪🎪
```

## What is this?

Carnie is a opinionated agent orchestration system with sane defaults installed inside your project. It brings a crew of helpers into your repo, keeps their work tracked, and helps with project management (and much more).

## Quick Start

```bash
# Initialize Carnie Camp in your project
carnie camp init

# Print the operator command
carnie operator

# View the beads dashboard
carnie dashboard
```

## Workflows

### Planning with Randy

Run `carnie operator` to print a ready-to-paste command. Paste it in your terminal to start the operator planning session.

## CLI Commands

- `carnie camp init` - Initialize Carnie Camp in your project
- `carnie operator` - Print the operator command
- `carnie dashboard` - Launch full-screen beads dashboard
- `carnie workorder` - Create and manage work orders

## Core Concepts

### The Operator 🎛️

Randy the Operator is your main interface to Carnie Camp. He helps you plan work, manage agents, and coordinate tasks. Randy runs on Opencode by default (Claude Code also supported) and can trigger workflows based on your conversations.

**What Randy can do:**

- Break down ideas into structured Beads (epic → feature → task → subtask hierarchy)
- Suggest dependencies between Beads
- Draft the exact `bd` commands to run

### Camp 🏕️

Your project-local workspace created with `carnie camp init`. The `camp.yml` file contains:

- `camp.yml` - Configuration for tools, preferences, and project settings
- In the future: `carniecamp.db` - Store executed workflow data
 - `carniecamp.db` - Store local work orders

Example `camp.yml`:

```yaml
version: 1
name: my-project

# Optional: Project description
description: My awesome project

# Operator configuration (Randy)
operator:
  # Model for the operator (default: openai/gpt-5.2-codex)
  model: openai/gpt-5.2-codex

  # Planning tool: "opencode" or "claude" (default: opencode)
  planning_tool: opencode

  # Optional: Custom planning prompt file path
  # planning_prompt_file: .carniecamp/custom_prompt.md

# Default settings for Carnies (workers)
defaults:
  # Model for agent workers (default: openai/gpt-5.2-codex)
  agent_model: openai/gpt-5.2-codex
```

### Beads 📌

Beads are the unit of work in Carnie Camp with a hierarchical structure:

- **Epic** - Large body of work
- **Feature** - Specific capability or improvement
- **Task** - Concrete work item
- **Subtask** - Granular step

Beads can have dependencies and are tracked through their lifecycle.

### Carnies 🎪

Carnies are the workers - Opencode/Claude Code sessions with primed context. When assigned work, they:

1. Claim tasks (mark as in-progress)
2. Complete the work
3. Track new ideas or blockers encountered
4. Mark tasks as completed
5. Move to the next task

See `docs/CAMP.md` and `docs/OPERATOR.md` for details.
Work orders are documented in `docs/WORK_ORDERS.md`.

## Carnie Rules

- All actions must be observable.
- Given Bead hierarchy must be completed before finishing. After all tasks are marked complete, report completion to Randy.

## Installation

- see Taskfile.yml

## References / inspiration

- Carnie is heavily inspired by [gastown](https://github.com/steveyegge/gastown)
- Git-based task management handled by [beads](https://github.com/steveyegge/beads)
- All execution is done using [dagger](https://github.com/dagger/dagger)
- Agent by [opencode](https://github.com/anomalyco/opencode)
