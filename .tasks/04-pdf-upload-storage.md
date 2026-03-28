# Task 04: PDF Upload & Storage

Handle PDF file uploads and filesystem storage.

## Steps

- [ ] Create `internal/pdf/storage.go` — save uploaded PDF to filesystem, return path
- [ ] Configure storage directory (default: `./data/pdfs/`)
- [ ] `POST /api/papers` — accept `multipart/form-data` with a PDF file
  - Validate file is PDF (check content type / magic bytes)
  - Store file to disk
  - Create paper record in DB with title (from filename), file path, file size
  - Return paper metadata as JSON
- [ ] Write tests: upload success, reject non-PDF, file lands on disk
- [ ] `DELETE /api/papers/:id` — also remove the file from disk
- [ ] Test: delete removes both DB record and file
