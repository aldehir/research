import { describe, it, expect, beforeEach } from 'vitest';
import {
  getActivePanel,
  getIsMobile,
  setIsMobile,
  toggleSidebar,
  toggleChat,
  closePanel
} from '$lib/mobile-layout.svelte';

describe('mobile-layout state', () => {
  beforeEach(() => {
    setIsMobile(false);
  });

  describe('initial state', () => {
    it('starts with no active panel', () => {
      expect(getActivePanel()).toBeNull();
    });

    it('starts in desktop mode', () => {
      expect(getIsMobile()).toBe(false);
    });
  });

  describe('setIsMobile', () => {
    it('enables mobile mode', () => {
      setIsMobile(true);
      expect(getIsMobile()).toBe(true);
    });

    it('closes any open panel when switching to desktop', () => {
      setIsMobile(true);
      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');

      setIsMobile(false);
      expect(getActivePanel()).toBeNull();
    });
  });

  describe('toggleSidebar', () => {
    it('opens sidebar when no panel is active', () => {
      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');
    });

    it('closes sidebar when sidebar is active', () => {
      toggleSidebar();
      toggleSidebar();
      expect(getActivePanel()).toBeNull();
    });

    it('switches from chat to sidebar', () => {
      toggleChat();
      expect(getActivePanel()).toBe('chat');
      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');
    });
  });

  describe('toggleChat', () => {
    it('opens chat when no panel is active', () => {
      toggleChat();
      expect(getActivePanel()).toBe('chat');
    });

    it('closes chat when chat is active', () => {
      toggleChat();
      toggleChat();
      expect(getActivePanel()).toBeNull();
    });

    it('switches from sidebar to chat', () => {
      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');
      toggleChat();
      expect(getActivePanel()).toBe('chat');
    });
  });

  describe('closePanel', () => {
    it('closes sidebar', () => {
      toggleSidebar();
      closePanel();
      expect(getActivePanel()).toBeNull();
    });

    it('closes chat', () => {
      toggleChat();
      closePanel();
      expect(getActivePanel()).toBeNull();
    });

    it('is a no-op when no panel is open', () => {
      closePanel();
      expect(getActivePanel()).toBeNull();
    });
  });

  describe('mutual exclusion', () => {
    it('never has both panels open simultaneously', () => {
      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');

      toggleChat();
      expect(getActivePanel()).toBe('chat');

      toggleSidebar();
      expect(getActivePanel()).toBe('sidebar');

      closePanel();
      expect(getActivePanel()).toBeNull();
    });
  });
});
