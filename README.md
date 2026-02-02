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
carnie init

# Talk to Randy the Operator
carnie operator

# Check project status and beads
carnie status

# View the beads dashboard
carnie dashboard
```

## Workflows

### Planning with Randy

Talk to Randy naturally - be as vague or specific as you want:

```bash
carnie operator
> "Let's add user authentication"
> "We need to improve the dashboard performance"
> "What should we work on next?"
```

Randy will help expand your ideas into structured Beads with proper hierarchy and dependencies.

### Executing Work

Once Beads are created, assign them to Carnies:

```bash
# Randy assigns epic/feature to a Carnie
# Carnie opens in a tmux session and starts working through tasks
# You can monitor progress via status or dashboard
```

The Carnie autonomously works through tasks, creating new Beads for side-tracks or blockers, and notifies Randy when complete.

## CLI Commands

- `carnie init` - Initialize Carnie Camp in your project
- `carnie operator` - Talk to Randy the Operator
- `carnie status` - Show project details and beads summary
- `carnie dashboard` - Launch full-screen beads dashboard

## Core Concepts

### The Operator 🎛️

Randy the Operator is your main interface to Carnie Camp. He helps you plan work, manage agents, and coordinate tasks. Randy runs on Opencode by default (Claude Code also supported) and can trigger workflows based on your conversations.

**What Randy can do:**

- Break down ideas into structured Beads (epic → feature → task → subtask hierarchy)
- Create and manage dependencies between Beads
- Assign work to Carnies (agent workers)
- Provide overviews of existing work
- Answer questions about the project

### Camp 🏕️

Your project-local workspace created with `carnie init`. The `.carniecamp/` directory contains:

- `camp.yml` - Configuration for tools, preferences, and project settings
- Other relevant content (git-tracked except explicitly .gitignored)
- In the future: `carniecamp.db` - Store executed workflow data

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

Carnies are the workers - Opencode/Claude Code tmux sessions with primed context. When assigned work, they:

1. Claim tasks (mark as in-progress)
2. Complete the work
3. Track new ideas or blockers encountered
4. Mark tasks as completed
5. Move to the next task
6. Send mail when all related work is done

Randy receives notifications when work is completed and can take actions like code review or merging.

See `docs/CAMP.md` and `docs/OPERATOR.md` for details.

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
