// @vitest-environment jsdom
import { describe, it, expect, beforeEach } from 'vitest';
import {
  getSidebarWidth,
  getChatWidth,
  setSidebarWidth,
  setChatWidth,
  initPanelWidths,
  handleSidebarResize,
  handleChatResize
} from '$lib/panel-widths.svelte';
import { SIDEBAR_DEFAULT, CHAT_DEFAULT, SIDEBAR_MIN, SIDEBAR_MAX, CHAT_MIN, CHAT_MAX } from '$lib/panel-resize';

describe('panel-widths reactive state', () => {
  beforeEach(() => {
    localStorage.clear();
    // Reset to defaults
    setSidebarWidth(SIDEBAR_DEFAULT);
    setChatWidth(CHAT_DEFAULT);
  });

  describe('defaults', () => {
    it('sidebar starts at default width', () => {
      expect(getSidebarWidth()).toBe(SIDEBAR_DEFAULT);
    });

    it('chat starts at default width', () => {
      expect(getChatWidth()).toBe(CHAT_DEFAULT);
    });
  });

  describe('setSidebarWidth / setChatWidth', () => {
    it('updates sidebar width', () => {
      setSidebarWidth(350);
      expect(getSidebarWidth()).toBe(350);
    });

    it('updates chat width', () => {
      setChatWidth(450);
      expect(getChatWidth()).toBe(450);
    });
  });

  describe('initPanelWidths', () => {
    it('restores saved widths from localStorage', () => {
      localStorage.setItem('panel-widths', JSON.stringify({ sidebar: 320, chat: 420 }));
      initPanelWidths();
      expect(getSidebarWidth()).toBe(320);
      expect(getChatWidth()).toBe(420);
    });

    it('keeps defaults when nothing saved', () => {
      initPanelWidths();
      expect(getSidebarWidth()).toBe(SIDEBAR_DEFAULT);
      expect(getChatWidth()).toBe(CHAT_DEFAULT);
    });

    it('keeps defaults for invalid saved data', () => {
      localStorage.setItem('panel-widths', 'garbage');
      initPanelWidths();
      expect(getSidebarWidth()).toBe(SIDEBAR_DEFAULT);
      expect(getChatWidth()).toBe(CHAT_DEFAULT);
    });
  });

  describe('handleSidebarResize', () => {
    it('increases sidebar width on positive delta', () => {
      handleSidebarResize(20, 1200);
      expect(getSidebarWidth()).toBe(SIDEBAR_DEFAULT + 20);
    });

    it('decreases sidebar width on negative delta', () => {
      handleSidebarResize(-20, 1200);
      expect(getSidebarWidth()).toBe(SIDEBAR_DEFAULT - 20);
    });

    it('clamps to sidebar minimum', () => {
      handleSidebarResize(-500, 1200);
      expect(getSidebarWidth()).toBe(SIDEBAR_MIN);
    });

    it('clamps to sidebar maximum', () => {
      handleSidebarResize(500, 1200);
      expect(getSidebarWidth()).toBe(SIDEBAR_MAX);
    });

    it('saves to localStorage', () => {
      handleSidebarResize(20, 1200);
      const stored = JSON.parse(localStorage.getItem('panel-widths')!);
      expect(stored.sidebar).toBe(SIDEBAR_DEFAULT + 20);
    });
  });

  describe('handleChatResize', () => {
    it('increases chat width on positive delta', () => {
      handleChatResize(20, 1200);
      expect(getChatWidth()).toBe(CHAT_DEFAULT + 20);
    });

    it('clamps to chat minimum', () => {
      handleChatResize(-500, 1200);
      expect(getChatWidth()).toBe(CHAT_MIN);
    });

    it('clamps to chat maximum', () => {
      handleChatResize(500, 1200);
      expect(getChatWidth()).toBe(CHAT_MAX);
    });

    it('saves to localStorage', () => {
      handleChatResize(20, 1200);
      const stored = JSON.parse(localStorage.getItem('panel-widths')!);
      expect(stored.chat).toBe(CHAT_DEFAULT + 20);
    });
  });
});
