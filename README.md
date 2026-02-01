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

## Core Concepts

### The Operator 🎛️

Operator is the main interaction point for all things related to Carnie. You can ask it about Carnie, plan work and manage agents.

### Camp 🏕️

Your project-local workspace. Initialize it with `carnie camp init` to create `camp.yml` inside the repo. Contains general configuration related to tools, user preferences, and project-specific settings.

### Beads 📌

Beads are the unit of work in Carnie. They represent a single task or a group of tasks that need to be completed. Beads can be created, assigned, and tracked using the Operator.

See `docs/CAMP.md` and `docs/OPERATOR.md` for details.

## Carnie Rules

- TODO

## Installation

- TODO

## References / inspiration

- Carnie is heavily inspired by [gastown](https://github.com/steveyegge/gastown)
- Git-based task management handled by [beads](https://github.com/steveyegge/beads)
- All execution is done using [dagger](https://github.com/dagger/dagger)
- Agent by [opencode](https://github.com/anomalyco/opencode)
