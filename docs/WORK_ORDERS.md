# Work Orders

Work Orders are lightweight, local tasks that route specific work to a Carnie.
They live in a SQLite database inside your repo and are designed for fast CLI workflows.

## Quick Start

```bash
# Create a work order
carnie workorder create --title "Add WorkOrder persistence" --description "Store work orders in SQLite" --bead cn-ta1.1

# List work orders
carnie workorder list

# Show a work order
carnie workorder show 1

# Update status
carnie workorder update 1 --status in_progress

# Render a Carnie prompt (copied to clipboard when possible)
carnie workorder prompt 1
```

## Status Flow

Work orders follow a simple, enforced state machine:

- `draft` -> `ready` -> `in_progress` -> `done`
- `ready` and `in_progress` can move to `blocked`
- `blocked` can return to `ready` or `in_progress`
- Any active state can move to `canceled`

## Storage Location

Work Orders are stored in `.carnie/carniecamp.db` at the Camp root (where `camp.yml` lives).

## Templates

The `workorder prompt` command renders a prompt using the embedded template and the Carnie role context.
It includes work order details and bead descriptions when available.
