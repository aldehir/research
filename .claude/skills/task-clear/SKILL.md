---
name: task-clear
description: Clear all tasks from the .tasks/ directory. Use when starting a new project phase or resetting task tracking.
disable-model-invocation: true
---

# Clear Tasks

Remove all task files from `.tasks/` and reset `index.md`.

**Do NOT delete `.tasks/next-id`** — task IDs must never be reused because git commits reference them via `[task-NN]` tags.

## Steps

1. Delete all `.md` files in `.tasks/` (task files, index, pert chart — everything)
2. Create a fresh `.tasks/index.md` with the template below
3. Confirm what was removed

## index.md template

```markdown
# Task Index

## Tasks

(no tasks yet)
```
