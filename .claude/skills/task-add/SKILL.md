---
name: task-add
description: Add a new task to .tasks/ with preliminary code exploration. Use when the user describes a feature, bug fix, or piece of work to track.
disable-model-invocation: false
argument-hint: <task description>
---

# Add Task

Create a new task file in `.tasks/` and register it in `.tasks/index.md`.

## Steps

1. Read `.tasks/next-id` to get the task number. Zero-pad to 2 digits.
2. **Explore the codebase** to understand what exists relevant to the task:
   - Identify files, functions, and modules that will be touched
   - Note existing patterns, tests, and conventions in those areas
   - Identify dependencies on other tasks if any
3. Create `.tasks/NN-slug.md` using the template below, filling in findings from exploration.
4. Add the task to `.tasks/index.md` with status `TODO`.
5. **Increment `.tasks/next-id`** — write the next number (current + 1) back to the file.

## Task file template

```markdown
# Task NN: Title

Short description of what this task accomplishes.

## Context

Summary of relevant existing code discovered during exploration:
- Key files and modules involved
- Existing patterns to follow
- Dependencies or prerequisites

## Checklist

- [ ] First TDD slice (describe what the test covers)
- [ ] Second TDD slice
- [ ] (continue as needed — each item should be a small, testable unit of work)

## Notes

Any edge cases, open questions, or design decisions to resolve.
```

## index.md format

Each task is a line in the `## Tasks` section:

```
- [ ] NN — Task Title `TODO`
```

Status values: `TODO`, `IN PROGRESS`, `COMPLETE`
