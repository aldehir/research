import { describe, it, expect } from 'vitest';

// Test the resize delta logic extracted from ResizeHandle behavior:
// For a left-side panel, dragging right → positive delta
// For a right-side panel, dragging right → negative delta (inverted)
describe('resize delta logic', () => {
  function computeDelta(side: 'left' | 'right', startX: number, currentX: number): number {
    const rawDelta = currentX - startX;
    return side === 'right' ? -rawDelta : rawDelta;
  }

  it('left panel: dragging right produces positive delta', () => {
    expect(computeDelta('left', 200, 250)).toBe(50);
  });

  it('left panel: dragging left produces negative delta', () => {
    expect(computeDelta('left', 200, 150)).toBe(-50);
  });

  it('right panel: dragging right produces negative delta', () => {
    expect(computeDelta('right', 200, 250)).toBe(-50);
  });

  it('right panel: dragging left produces positive delta', () => {
    expect(computeDelta('right', 200, 150)).toBe(50);
  });

  it('no movement produces zero delta', () => {
    expect(computeDelta('left', 200, 200)).toBe(0);
    expect(Math.abs(computeDelta('right', 200, 200))).toBe(0);
  });
});
