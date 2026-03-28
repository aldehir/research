# Task 03: Fix crypto.randomUUID error on plain HTTP

Replace `crypto.randomUUID()` with a fallback that works outside secure contexts (plain HTTP).

## Context

- `crypto.randomUUID()` is called in `frontend/src/lib/chat.svelte.ts` (lines 54, 77)
- It generates temporary client-side IDs for `Message` objects added to the local reactive array
- These IDs are only used for Svelte rendering — the server assigns real IDs
- `crypto.randomUUID()` requires a secure context (HTTPS or localhost); fails on plain HTTP

## Checklist

- [x] Replace `crypto.randomUUID()` with a fallback that works in all contexts
- [x] Verify no other uses of `crypto.randomUUID()` exist in the codebase

## Notes

- `crypto.getRandomValues()` works in all contexts and can generate a UUID v4
- Alternatively, a simple incrementing counter would suffice since these IDs are ephemeral client-side only
