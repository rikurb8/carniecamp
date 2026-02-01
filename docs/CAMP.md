# Camp Command

The `camp` command manages your Carnie workspace configuration.

## Quick Start

```bash
# Initialize a new workspace
carnie camp init

# Initialize with a custom name
carnie camp init --name my-project

# Initialize with description
carnie camp init --name my-project --description "My awesome project"

# Overwrite existing config
carnie camp init --force
```

## What is a Camp?

A **Camp** is a small project workspace installed inside your repo. The `camp.yml` file stores project-level settings that Carnie uses for planning and coordination.

## camp.yml Configuration

Running `carnie camp init` creates a `camp.yml` file:

```yaml
version: 1
name: my-workspace
description: "Optional description"
operator:
  model: openai/gpt-5.2-codex
defaults:
  agent_model: openai/gpt-5.2-codex
```

Model values use the `<provider>/<model>` format (e.g., `openai/gpt-5.2-codex`).

### Fields

| Field | Description | Default |
|-------|-------------|---------|
| `version` | Config schema version | `1` |
| `name` | Workspace name | Directory name |
| `description` | What this workspace is for | (empty) |
| `operator.model` | Model for operator commands | `openai/gpt-5.2-codex` |
| `defaults.agent_model` | Default model for agents | `openai/gpt-5.2-codex` |

## Commands

### `camp init`

Creates a new `camp.yml` configuration file.

**Flags:**

- `--name` - Set the workspace name (defaults to current directory name)
- `--description` - Set a description for the workspace
- `--force` - Overwrite existing `camp.yml` if it exists

**Examples:**

```bash
# Basic init (uses directory name)
carnie camp init

# With all options
carnie camp init --name mycamp --description "Development workspace" --force
```
