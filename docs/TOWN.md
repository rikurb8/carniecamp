# Town Command

The `town` command manages your Bordertown workspace configuration.

## Quick Start

```bash
# Initialize a new workspace
bt town init

# Initialize with a custom name
bt town init --name my-project

# Initialize with description
bt town init --name my-project --description "My awesome project"

# Overwrite existing config
bt town init --force
```

## What is a Town?

A **Town** is a small project workspace installed inside your repo. The `town.yml` file stores project-level settings that Bordertown uses for planning and coordination.

## town.yml Configuration

Running `bt town init` creates a `town.yml` file:

```yaml
version: 1
name: my-workspace
description: "Optional description"
mayor:
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
| `mayor.model` | Model for mayor commands | `openai/gpt-5.2-codex` |
| `defaults.agent_model` | Default model for agents | `openai/gpt-5.2-codex` |

## Commands

### `town init`

Creates a new `town.yml` configuration file.

**Flags:**

- `--name` - Set the workspace name (defaults to current directory name)
- `--description` - Set a description for the workspace
- `--force` - Overwrite existing `town.yml` if it exists

**Examples:**

```bash
# Basic init (uses directory name)
bt town init

# With all options
bt town init --name mytown --description "Development workspace" --force
```
