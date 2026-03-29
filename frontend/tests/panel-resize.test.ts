// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import {
  SIDEBAR_MIN,
  SIDEBAR_MAX,
  SIDEBAR_DEFAULT,
  CHAT_MIN,
  CHAT_MAX,
  CHAT_DEFAULT,
  TOC_MIN,
  TOC_MAX,
  TOC_DEFAULT,
  CENTER_MIN,
  clampWidth,
  clampResize,
  savePanelWidths,
  loadPanelWidths
} from '$lib/panel-resize';

describe('panel-resize', () => {
  describe('constants', () => {
    it('has correct sidebar constraints', () => {
      expect(SIDEBAR_MIN).toBe(180);
      expect(SIDEBAR_MAX).toBe(640);
      expect(SIDEBAR_DEFAULT).toBe(280);
    });

    it('has correct chat panel constraints', () => {
      expect(CHAT_MIN).toBe(240);
      expect(CHAT_MAX).toBe(800);
      expect(CHAT_DEFAULT).toBe(360);
    });

    it('has correct ToC panel constraints', () => {
      expect(TOC_MIN).toBe(160);
      expect(TOC_MAX).toBe(480);
      expect(TOC_DEFAULT).toBe(260);
    });

    it('has correct center minimum', () => {
      expect(CENTER_MIN).toBe(300);
    });
  });

  describe('clampWidth', () => {
    it('returns value within range unchanged', () => {
      expect(clampWidth(300, 180, 480)).toBe(300);
    });

    it('clamps to minimum', () => {
      expect(clampWidth(100, 180, 480)).toBe(180);
    });

    it('clamps to maximum', () => {
      expect(clampWidth(500, 180, 480)).toBe(480);
    });

    it('returns min when value equals min', () => {
      expect(clampWidth(180, 180, 480)).toBe(180);
    });

    it('returns max when value equals max', () => {
      expect(clampWidth(480, 180, 480)).toBe(480);
    });
  });

  describe('clampResize', () => {
    const totalWidth = 1800;

    it('allows resize within constraints when center has room', () => {
      expect(clampResize('sidebar', 300, 360, totalWidth)).toBe(300);
    });

    it('clamps sidebar to its minimum', () => {
      expect(clampResize('sidebar', 100, 360, totalWidth)).toBe(SIDEBAR_MIN);
    });

    it('clamps sidebar to its maximum', () => {
      expect(clampResize('sidebar', 700, 360, totalWidth)).toBe(SIDEBAR_MAX);
    });

    it('clamps chat to its minimum', () => {
      expect(clampResize('chat', 200, 280, totalWidth)).toBe(CHAT_MIN);
    });

    it('clamps chat to its maximum', () => {
      expect(clampResize('chat', 900, 280, totalWidth)).toBe(CHAT_MAX);
    });

    it('limits sidebar when center would be too narrow', () => {
      // totalWidth=800, chat=360 → max sidebar = 800-360-300 = 140, but sidebar min is 180
      // so if total is too small, center constraint wins but sidebar min still applies
      const result = clampResize('sidebar', 400, 360, 800);
      // center = 800 - result - 360 must be >= 300
      // result <= 800 - 360 - 300 = 140, but min is 180, so return 180
      // Actually: 140 < 180, so sidebar min wins → 180
      expect(result).toBe(180);
    });

    it('limits chat when center would be too narrow', () => {
      // totalWidth=800, sidebar=280 → max chat = 800-280-300 = 220, but chat min is 240
      const result = clampResize('chat', 400, 280, 800);
      expect(result).toBe(240);
    });

    it('shrinks sidebar to protect center minimum on wide enough screen', () => {
      // totalWidth=1000, chat=360 → max sidebar = 1000-360-300 = 340
      const result = clampResize('sidebar', 400, 360, 1000);
      expect(result).toBe(340);
    });

    it('shrinks chat to protect center minimum on wide enough screen', () => {
      // totalWidth=1000, sidebar=280 → max chat = 1000-280-300 = 420
      const result = clampResize('chat', 500, 280, 1000);
      expect(result).toBe(420);
    });
  });

  describe('localStorage persistence', () => {
    beforeEach(() => {
      localStorage.clear();
    });

    it('saves and loads panel widths', () => {
      savePanelWidths({ sidebar: 300, chat: 400, toc: 280 });
      expect(loadPanelWidths()).toEqual({ sidebar: 300, chat: 400, toc: 280 });
    });

    it('returns null when nothing saved', () => {
      expect(loadPanelWidths()).toBeNull();
    });

    it('returns null for invalid JSON', () => {
      localStorage.setItem('panel-widths', 'not json');
      expect(loadPanelWidths()).toBeNull();
    });

    it('returns null for non-object values', () => {
      localStorage.setItem('panel-widths', '"string"');
      expect(loadPanelWidths()).toBeNull();
    });

    it('returns null when sidebar is missing', () => {
      localStorage.setItem('panel-widths', JSON.stringify({ chat: 400, toc: 260 }));
      expect(loadPanelWidths()).toBeNull();
    });

    it('returns null when chat is missing', () => {
      localStorage.setItem('panel-widths', JSON.stringify({ sidebar: 300, toc: 260 }));
      expect(loadPanelWidths()).toBeNull();
    });

    it('returns null when values are not numbers', () => {
      localStorage.setItem('panel-widths', JSON.stringify({ sidebar: 'wide', chat: 400, toc: 260 }));
      expect(loadPanelWidths()).toBeNull();
    });

    it('loads with default toc when toc is missing (backward compat)', () => {
      localStorage.setItem('panel-widths', JSON.stringify({ sidebar: 300, chat: 400 }));
      expect(loadPanelWidths()).toEqual({ sidebar: 300, chat: 400, toc: TOC_DEFAULT });
    });

    it('clamps loaded values to valid ranges', () => {
      savePanelWidths({ sidebar: 50, chat: 900, toc: 10 });
      const loaded = loadPanelWidths();
      expect(loaded).toEqual({ sidebar: SIDEBAR_MIN, chat: CHAT_MAX, toc: TOC_MIN });
    });
  });
});
