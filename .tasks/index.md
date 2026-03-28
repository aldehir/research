# Task Runner

## Instructions

1. Read `PRD.md` to understand the full product scope.
2. Read `.tasks/pert.md` to understand task dependencies.
3. Check the task status list below. Pick the next task that is:
   - **Not started** (status is `TODO`)
   - **Not blocked** (all dependencies are `COMPLETE`)
   - If multiple tasks are unblocked, prefer the one on the critical path:
     `01 → 02 → 03 → 04 → 05 → 07 → 08 → 13`
4. Update the task's status below to `IN PROGRESS`.
5. Read the task file (e.g. `.tasks/01-project-scaffolding.md`) for details.
6. Implement using TDD: RED → GREEN → REFACTOR for each checklist item.
7. When all checklist items are done, update the status below to `COMPLETE`.
8. Return to step 3.

## Task Status

- [x] 01 — Project Scaffolding `COMPLETE`
- [x] 02 — Database Setup `COMPLETE`
- [ ] 03 — Papers CRUD API `IN PROGRESS`
- [x] 04 — PDF Upload & Storage `COMPLETE`
- [x] 05 — PDF Serving `COMPLETE`
- [x] 06 — Frontend: Paper Library `COMPLETE`
- [x] 07 — Frontend: PDF Viewer `COMPLETE`
- [x] 08 — Text Selection & Context `COMPLETE`
- [x] 09 — Chat Sessions API `COMPLETE`
- [x] 10 — Anthropic Client `COMPLETE`
- [ ] 11 — Messages Endpoint + SSE `TODO`
- [ ] 12 — Frontend: Chat Panel `TODO`
- [ ] 13 — Layout & Integration `TODO`
