import {
  SIDEBAR_DEFAULT,
  CHAT_DEFAULT,
  clampResize,
  savePanelWidths,
  loadPanelWidths
} from '$lib/panel-resize';

let sidebarWidth = $state(SIDEBAR_DEFAULT);
let chatWidth = $state(CHAT_DEFAULT);

export function getSidebarWidth(): number {
  return sidebarWidth;
}

export function getChatWidth(): number {
  return chatWidth;
}

export function setSidebarWidth(w: number): void {
  sidebarWidth = w;
}

export function setChatWidth(w: number): void {
  chatWidth = w;
}

export function initPanelWidths(): void {
  const saved = loadPanelWidths();
  if (saved) {
    sidebarWidth = saved.sidebar;
    chatWidth = saved.chat;
  }
}

export function handleSidebarResize(delta: number, totalWidth: number): void {
  sidebarWidth = clampResize('sidebar', sidebarWidth + delta, chatWidth, totalWidth);
  savePanelWidths({ sidebar: sidebarWidth, chat: chatWidth });
}

export function handleChatResize(delta: number, totalWidth: number): void {
  chatWidth = clampResize('chat', chatWidth + delta, sidebarWidth, totalWidth);
  savePanelWidths({ sidebar: sidebarWidth, chat: chatWidth });
}
