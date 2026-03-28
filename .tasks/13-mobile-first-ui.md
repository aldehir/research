# Task 13: Mobile-First Slide-Over Panel Layout

Make the app usable on iPad and tablets by switching to a slide-over panel layout below 1024px. The PDF viewer becomes full-width; sidebar and chat slide in as overlays from the left and right edges respectively.

## Context

The current layout is a fixed three-panel flexbox (`+page.svelte`):
- Sidebar: 280px fixed, min 220px
- Content: flex: 1, min 300px
- Chat: 360px fixed, min 300px (or ~40px collapsed)

Total minimum is ~940px, which doesn't fit on iPad (1024px landscape, 768px portrait). There are **zero media queries** anywhere in the frontend. All styling is component-scoped `<style>` blocks — no CSS framework.

**Primary use case on tablet**: Extended chat conversations while referencing the PDF. Chat panel is the most important overlay. Sidebar (paper management) is secondary — functional but not optimized.

### Key files
- `frontend/src/routes/+page.svelte` — Main layout shell (`.app-shell`, `.app-layout`, `.sidebar`, `.content`)
- `frontend/src/lib/ChatPanel.svelte` — Chat panel with `collapsed` state toggle, 360px width
- `frontend/src/lib/PdfViewer.svelte` — PDF viewer, toolbar, TOC toggle, ResizeObserver for fit-to-width
- `frontend/src/lib/MessageInput.svelte` — Chat input area
- `frontend/src/lib/TocPanel.svelte` — TOC panel (260px, inside viewer-body)

### Existing patterns
- Chat panel already has a collapse/expand toggle (`let collapsed = $state(false)`)
- PDF viewer has ResizeObserver that recalculates fit-to-width on container resize
- TOC panel already slides in/out within the viewer body
- All state uses Svelte 5 runes (`$state`, `$derived`, `$effect`)

## Design

### Breakpoint: 1024px

Below 1024px, switch to mobile layout. Above 1024px, desktop layout is unchanged.

### Default state (PDF focused)
```
┌──────────────────────────────────┐
│ ☰  Header                    💬 │
├──────────────────────────────────┤
│                                  │
│         PDF Viewer               │
│        (full width)              │
│                                  │
└──────────────────────────────────┘
```
- Sidebar and chat panel are hidden
- Header shows: hamburger (☰) on left, chat toggle (💬) on right
- PDF viewer takes 100% width

### Chat active (slide from right)
```
┌──────────────────────────────────┐
│ ☰  Header                    💬 │
├──────────────┬───────────────────┤
│  backdrop    │  Chat Panel       │
│  (dimmed)    │  (~70% width)     │
│              │                   │
│  tap to      │  Sessions         │
│  dismiss     │  Messages         │
│              │  Input            │
└──────────────┴───────────────────┘
```
- Chat slides in from right, takes ~70% of viewport width
- Semi-transparent backdrop over PDF; tap backdrop to dismiss
- CSS `transform: translateX()` with ~250ms ease transition

### Sidebar active (slide from left)
```
┌──────────────────────────────────┐
│ ☰  Header                    💬 │
├───────────┬──────────────────────┤
│ Sidebar   │  backdrop            │
│ (~280px)  │  (dimmed)            │
│           │                      │
│ Papers    │  tap to              │
│ Upload    │  dismiss             │
└───────────┴──────────────────────┘
```
- Sidebar slides from left, ~280px (same as desktop width)
- Same backdrop pattern

### Rules
- No simultaneous overlays — opening one closes the other
- Touch targets: minimum 44px for all interactive elements in mobile layout
- Backdrop: semi-transparent dark overlay, click/tap to dismiss panel
- Desktop layout (>1024px) stays exactly as-is

## Checklist

- [ ] Add media query breakpoint and mobile detection to `+page.svelte`
  - Add `1024px` breakpoint media query
  - Hide sidebar and chat panel from normal flow below breakpoint
  - Make `.content` full-width
  - Add hamburger button to header (left) and chat toggle button (right), visible only below breakpoint
- [ ] Implement slide-over panel system
  - Add overlay/backdrop component or markup (semi-transparent, click-to-dismiss)
  - Position sidebar as fixed/absolute overlay from left with slide transition
  - Position chat panel as fixed/absolute overlay from right with slide transition
  - Wire open/close state: hamburger toggles sidebar, chat button toggles chat
  - Ensure opening one panel closes the other
- [ ] Adjust chat panel for overlay mode
  - Chat panel takes ~70% viewport width when in overlay mode
  - Ensure `MessageInput` and `MessageThread` work well at overlay width
  - Chat panel gets full height (below header)
- [ ] Adjust sidebar for overlay mode
  - Sidebar keeps ~280px width, full height below header
  - Paper list and upload zone remain functional
- [ ] Touch target sizing
  - Audit and increase button/control sizes to 44px minimum in mobile layout
  - Header buttons, toolbar buttons, chat toggle, paper list items
- [ ] PDF viewer adjustments
  - Fit-to-width should recalculate when overlays open/close (ResizeObserver already handles this)
  - TOC panel: consider hiding the TOC toggle on mobile (or making TOC full-screen overlay too)
  - Toolbar controls should remain usable at mobile sizes

## Notes

- The ResizeObserver in `PdfViewer.svelte` should automatically handle PDF rescaling when the layout changes — no special handling needed there.
- The chat panel's existing `collapsed` state may need to be adapted or replaced with the new overlay state for mobile. Keep the desktop collapse behavior unchanged.
- Swipe gestures (edge-swipe to open panels) are a nice-to-have for a follow-up task, not in scope here.
- No portrait vs landscape differentiation — same fluid layout for both orientations.
