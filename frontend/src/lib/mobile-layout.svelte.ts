/**
 * Mobile layout state management.
 *
 * Below 1024px the sidebar and chat panel become slide-over overlays.
 * Only one panel can be open at a time.
 */

export type MobilePanel = 'sidebar' | 'chat' | null;

let activePanel = $state<MobilePanel>(null);
let isMobile = $state(false);

export function getActivePanel(): MobilePanel {
  return activePanel;
}

export function getIsMobile(): boolean {
  return isMobile;
}

export function setIsMobile(value: boolean): void {
  isMobile = value;
  if (!value) {
    activePanel = null;
  }
}

export function toggleSidebar(): void {
  activePanel = activePanel === 'sidebar' ? null : 'sidebar';
}

export function toggleChat(): void {
  activePanel = activePanel === 'chat' ? null : 'chat';
}

export function closePanel(): void {
  activePanel = null;
}
