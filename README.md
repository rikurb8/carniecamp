# Bordertown

**Sane agent orchestration**

## Overview

Bordertown is a reliable, portable agent orchestration system installed inside your project.

## CLI

- `bordertown status` shows project details and beads summary
- `bordertown dashboard` launches the full-screen beads dashboard

### Goal

Create a environment for:

- Project management
  - [x] Iterating on project ideas / technical planning
  - [ ] Creating actionable epics/features/tasks based on plan
- Implementation
  - [ ] Have agent implement task
  - [ ] Have agent(s) implement all available tasks (fully autonomous or human-in-the-loop)
  - [ ] Drop in to a specific point-in-time/context to do work yourself with Opencode
- Shipping
  - [ ] Known best practices for shipping products
  - [ ] Update / monitoring / debugging

### What Problem Does This Solve?

| Challenge                      | Bordertown Solution                          |
| ------------------------------ | -------------------------------------------- |
| Agents lose context on restart | Work persists in git-backed hooks            |
| Manual agent coordination      | Built-in mailboxes, identities, and handoffs |
| Easier project management      | Oppinionated way of managing work            |
| Scary to run YOLO agents       | Bordertown provides safe environment         |

## Core Concepts

NB: these names and concepts are inspired by the original gastown and more of a placeholder, adjust when needed. Especially better Bordertowny names are needed.

### The Mayor 🎩

Your primary AI coordinator. The Mayor is a Opencode instance with full context about your workspace, projects, and agents. **Start here** - just tell the Mayor what you want to accomplish.

### Town 🏘️

Your project-local workspace. Initialize it with `bordertown town init` to create `town.yml` inside the repo.

### Beads 📌

Work items tracked by `bd` inside the project repo.

See `docs/TOWN.md` and `docs/MAYOR.md` for details.

### Crew Members 👤

Your personal workspace within the project. Where you do hands-on work.

### Polecats 🦨

Ephemeral worker agents that spawn, complete a task, and disappear.

### Hooks 🪝

Git worktree-based persistent storage for agent work. Survives crashes and restarts.

### Convoys 🚚

Work tracking units. Bundle multiple beads that get assigned to agents.

## Installation

- TODO

### References / inspiration

- Bordertown is heavily inspired by [gastown](https://github.com/steveyegge/gastown)
- Git-based task management handled by [beads](https://github.com/steveyegge/beads)
- All execution is done using [dagger](https://github.com/dagger/dagger)
- Agent by [opencode](https://github.com/anomalyco/opencode)
