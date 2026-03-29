---
name: task-work
description: Select a task from .tasks/ and work on it using TDD RED-GREEN-REFACTOR. Updates task status and commits when done.
disable-model-invocation: true
argument-hint: [task number]
---

# Work on Task

Pick a task and implement it using strict TDD.

## Steps

### 1. Select task

- If `$ARGUMENTS` specifies a task number, use that task.
- Otherwise, read `.tasks/index.md` and pick the first `TODO` task.
- Read the task file (`.tasks/NN-slug.md`) to understand the work.
- Update `.tasks/index.md` status to `IN PROGRESS`.

### 2. Plan

- Use the **Plan agent** (`subagent_type: "Plan"`) to design the implementation.
- Pass it the full task description from the task file and any relevant context (e.g. related source files, API conventions from CLAUDE.md).
- Review the plan it returns. If it looks reasonable, proceed. If not, refine before continuing.

### 3. Implement with TDD

For each checklist item in the task file:

1. **RED**: Write a failing test that describes the expected behavior. Run the test — confirm it fails.
2. **GREEN**: Write the minimal production code to make the test pass. Run the test — confirm it passes.
3. **REFACTOR**: Clean up code while keeping tests green. Run tests again.
4. Check off the item in the task file.

**Frontend TDD scope**: Only use TDD for business logic (stores, utilities, data transformations). Do NOT use TDD for visual/UI work (component markup, styling, layout).

Do NOT skip the failing test step. Do NOT write production code without a test.

Run the full test suite periodically to catch regressions:
- Backend: `go test ./...`
- Frontend: `cd frontend && pnpm test`

### 4. Finish up

- Verify all checklist items are checked off.
- Run the full test suite one final time.
- Update `.tasks/index.md` status to `COMPLETE`.
- Create a git commit. **Include `[task-NN]` in the commit message** (e.g. `feat: add streaming markdown renderer [task-14]`). This tag allows commits to be queried by task.
