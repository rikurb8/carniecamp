# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Carnie is an agent orchestration system that installs a crew of helpers into a project. It uses **beads** (`bd` CLI) for git-based issue tracking with a hierarchy: Epic → Feature → Task → Subtask.

## Build & Development Commands

```bash
# Build
go build -o ./bin/carnie ./cmd/carnie
task build                    # Same via taskfile

# Test
go test ./...
go test ./internal/cli -run TestCampInit  # Single test

# Lint & Format
go vet ./...
gofmt -w ./cmd ./internal

# Run
go run ./cmd/carnie
task ci                       # Format + lint + test
```

## Code Architecture

```
cmd/carnie/main.go           # Entry point - calls cli.Execute()

internal/
├── cli/                     # Cobra command definitions ONLY
│   ├── root.go              # Root command, viper config setup
│   ├── camp.go              # camp init command
│   ├── operator.go          # operator command (prints session command)
│   └── dashboard.go         # dashboard command wrapper
├── bd/                      # Beads CLI integration
│   ├── types.go             # StatusSummary, Status structs
│   └── run.go               # RunJSON, RunJSONInDir helpers
├── config/                  # camp.yml configuration
│   └── camp.go              # CampConfig, OperatorConfig, Defaults
├── dashboard/               # BubbleTea TUI (carnival-themed)
│   ├── model.go             # Model struct, NewModel
│   ├── update.go            # Init, Update, tickCmd
│   ├── render.go            # View, all render* functions
│   ├── data.go              # Data loading from beads
│   └── ...                  # Styles, list state, utils
├── operator/                # Planning session builder
│   ├── command.go           # BuildPlanningCommand
│   └── prompt.go            # System prompt construction
└── session/                 # Tool-agnostic session abstraction
```

**Key pattern**: `internal/cli/` contains only cobra command wiring. Business logic lives in sibling packages (`bd`, `config`, `dashboard`, `operator`, `session`).

## Issue Tracking with Beads

This project uses `bd` (beads) for issue tracking:

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress
bd close <id>
bd sync               # Sync with git remote
```

## Session Close Protocol

Before completing work, run this checklist:

1. `git status` - check changes
2. `git add <files>` - stage code
3. `bd sync` - commit beads changes
4. `git commit -m "..."` - commit code
5. `bd sync` - commit any new beads
6. `git push` - push to remote

Work is NOT complete until `git push` succeeds.
