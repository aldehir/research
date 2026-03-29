---
name: task-review
description: Review completed work for a task. Finds commits by task ID and reviews for dead code, redundancies, and simplification opportunities.
disable-model-invocation: true
argument-hint: <task number>
---

# Review Task

Review the implementation of a completed task for quality issues.

## Steps

### 1. Identify the task and its changes

- Read `.tasks/$ARGUMENTS-*.md` (glob for the task file by number prefix) to understand what was implemented.
- Run `git log --all --oneline --grep='[task-$ARGUMENTS]'` to find all commits tagged with this task.
- For each commit, run `git diff <commit>~1 <commit>` to collect the full set of changes.
- If no commits are found with the tag, fall back to reading the task file's checklist and manually identifying the relevant files.

### 2. Review the changed code

Read every file that was touched by the task's commits. Review for:

- **Dead code**: Unused imports, unreachable branches, variables written but never read, functions defined but never called, commented-out code.
- **Redundancies**: Duplicated logic that could be shared, copy-pasted blocks with minor variations, repeated constants or magic numbers.
- **Simplification opportunities**: Over-engineered abstractions, unnecessary indirection, conditions that can be collapsed, verbose patterns that have shorter idiomatic equivalents.
- **Leftover artifacts**: Debug logging, TODO comments that should have been resolved, temporary workarounds that became permanent.

### 3. Report findings

Present findings grouped by category. For each issue:

- **File and line**: exact location
- **What**: describe the problem concisely
- **Suggested fix**: concrete code change or deletion

If the implementation is clean, say so — don't invent issues.

### 4. Apply fixes (with approval)

Ask the user whether to apply the suggested fixes. If approved:

1. Make the changes.
2. Run the relevant test suite (`go test ./...` and/or `cd frontend && pnpm test`).
3. Commit with message: `refactor: review cleanup [task-$ARGUMENTS]`
