# Carnie

**Step right up for sane agent orchestration**

## Welcome to the Big Top

Carnie is a reliable, portable agent orchestration system installed inside your project. It brings a crew of helpers into your repo, keeps their work tracked, and makes planning feel like a well-run show.

## Quick Start

```bash
carnie camp init
carnie status
carnie dashboard
carnie operator review
```

## CLI Highlights

- `carnie status` shows project details and beads summary
- `carnie dashboard` launches the full-screen beads dashboard
- `carnie operator new-epic` kicks off AI-assisted planning

## Why Carnie?

| Challenge                      | Carnie Solution                               |
| ------------------------------ | -------------------------------------------- |
| Agents lose context on restart | Work persists in git-backed hooks            |
| Manual agent coordination      | Built-in mailboxes, identities, and handoffs |
| Easier project management      | Opinionated way of managing work             |
| Scary to run YOLO agents       | Carnie runs a safer show                     |

## Core Cast

### The Operator 🎛️

Your ringmaster for AI planning. The Operator is an Opencode instance with full context about your workspace, projects, and agents. **Start here**—tell the Operator what you want to accomplish.

### Camp 🏕️

Your project-local workspace. Initialize it with `carnie camp init` to create `camp.yml` inside the repo.

### Beads 📌

Work items tracked by `bd` inside the project repo.

### Crew Members 👤

Your personal workspace within the project. Where you do hands-on work.

### Polecats 🦨

Ephemeral worker agents that spawn, complete a task, and disappear.

### Hooks 🪝

Git worktree-based persistent storage for agent work. Survives crashes and restarts.

### Convoys 🚚

Work tracking units. Bundle multiple beads that get assigned to agents.

See `docs/CAMP.md` and `docs/OPERATOR.md` for details.

## Installation

- TODO

## References / inspiration

- Carnie is heavily inspired by [gastown](https://github.com/steveyegge/gastown)
- Git-based task management handled by [beads](https://github.com/steveyegge/beads)
- All execution is done using [dagger](https://github.com/dagger/dagger)
- Agent by [opencode](https://github.com/anomalyco/opencode)
