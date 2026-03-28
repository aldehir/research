# PERT Chart

## Task Dependency Graph

```
                        ┌──────────────────┐
                        │  01 Scaffolding  │
                        └────────┬─────────┘
                 ┌───────────────┼───────────────┐
                 ▼               ▼               ▼
          ┌─────────────┐ ┌───────────┐ ┌──────────────┐
          │ 02 Database │ │ 06 Paper  │ │ 10 Anthropic │
          │    Setup    │ │ Library UI│ │    Client    │
          └──────┬──────┘ └─────┬─────┘ └──────┬───────┘
           ┌─────┴─────┐       │               │
           ▼           ▼       │               │
    ┌─────────────┐ ┌─────────────┐            │
    │ 03 Papers   │ │ 09 Chat     │            │
    │  CRUD API   │ │ Sessions API│            │
    └──────┬──────┘ └──────┬──────┘            │
           ▼               │                   │
    ┌─────────────┐        │                   │
    │ 04 PDF      │        └─────────┬─────────┘
    │ Upload      │                  ▼
    └──────┬──────┘        ┌──────────────────┐
           ▼               │ 11 Messages      │
    ┌─────────────┐        │ Endpoint + SSE   │
    │ 05 PDF      │        └────────┬─────────┘
    │ Serving     │                 │
    └──────┬──────┘                 │
           │                        │
           ▼                        ▼
    ┌─────────────┐        ┌──────────────┐
    │ 07 PDF      │        │ 12 Chat      │
    │ Viewer      │◄───────│ Panel UI     │◄─── 06
    └──────┬──────┘        └──────┬───────┘
           ▼                      │
    ┌─────────────┐               │
    │ 08 Text     │               │
    │ Selection   │               │
    └──────┬──────┘               │
           │                      │
           └──────────┬───────────┘
                      ▼
              ┌──────────────┐
              │ 13 Layout &  │
              │ Integration  │
              └──────────────┘
```

## Dependency Table

| Task | Name                        | Depends On  |
|------|-----------------------------|-------------|
| 01   | Project Scaffolding         | —           |
| 02   | Database Setup              | 01          |
| 03   | Papers CRUD API             | 02          |
| 04   | PDF Upload & Storage        | 03          |
| 05   | PDF Serving                 | 04          |
| 06   | Frontend: Paper Library     | 01          |
| 07   | Frontend: PDF Viewer        | 05, 06      |
| 08   | Text Selection & Context    | 07          |
| 09   | Chat Sessions API           | 02          |
| 10   | Anthropic Client            | 01          |
| 11   | Messages Endpoint + SSE     | 09, 10      |
| 12   | Frontend: Chat Panel        | 06, 11      |
| 13   | Layout & Integration        | 07, 08, 12  |

## Critical Path

```
01 → 02 → 03 → 04 → 05 → 07 → 08 → 13
```

This is the longest dependency chain (8 tasks). The chat track (02 → 09, 10 → 11 → 12) runs in parallel and is shorter, so it has slack.

## Parallel Tracks

| Track         | Tasks                  | Length |
|---------------|------------------------|--------|
| PDF backend   | 01 → 02 → 03 → 04 → 05 | 5    |
| PDF frontend  | 05 + 06 → 07 → 08     | 3      |
| Chat backend  | 02 → 09, 01 → 10 → 11 | 3      |
| Chat frontend | 06 + 11 → 12          | 1      |
| Integration   | 08 + 12 → 13          | 1      |
