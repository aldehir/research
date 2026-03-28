# Task 05: PDF Serving Endpoint

Serve stored PDF files to the frontend.

## Steps

- [ ] `GET /api/papers/:id/pdf` — serve the PDF file with correct content type
- [ ] Set `Content-Type: application/pdf`
- [ ] Set `Content-Disposition: inline` (display in browser, not download)
- [ ] Return 404 if paper or file not found
- [ ] Write handler tests
