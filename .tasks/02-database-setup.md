# Task 02: Database Setup (SQLite)

Set up SQLite database with schema migrations.

## Steps

- [ ] Add `modernc.org/sqlite` dependency
- [ ] Create `internal/store/db.go` — open/close DB, run migrations
- [ ] Create schema migration for `papers`, `chat_sessions`, and `messages` tables (see PRD data model)
- [ ] Write tests: migration runs cleanly, tables exist, re-running migration is idempotent
- [ ] Use `:memory:` SQLite databases in tests for speed
