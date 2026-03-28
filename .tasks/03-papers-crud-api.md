# Task 03: Papers CRUD API

Implement the papers metadata endpoints (no file upload yet).

## Store Layer

- [ ] `internal/store/papers.go` — Create, List, GetByID, Delete paper records
- [ ] Table-driven tests for each store method

## API Layer

- [ ] `GET /api/papers` — list all papers (JSON array)
- [ ] `GET /api/papers/:id` — get single paper metadata
- [ ] `DELETE /api/papers/:id` — delete paper record
- [ ] Handler tests using `httptest` and an in-memory DB
- [ ] Return 404 for missing papers
