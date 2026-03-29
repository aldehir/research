import {
  SIDEBAR_DEFAULT,
  CHAT_DEFAULT,
  TOC_DEFAULT,
  TOC_MIN,
  TOC_MAX,
  clampWidth,
  clampResize,
  savePanelWidths,
  loadPanelWidths
} from '$lib/panel-resize';

let sidebarWidth = $state(SIDEBAR_DEFAULT);
let chatWidth = $state(CHAT_DEFAULT);
let tocWidth = $state(TOC_DEFAULT);
let sidebarCollapsed = $state(false);
let savedSidebarWidth = SIDEBAR_DEFAULT;

export function getSidebarWidth(): number {
  return sidebarWidth;
}

export function getChatWidth(): number {
  return chatWidth;
}

export function getTocWidth(): number {
  return tocWidth;
}

export function setSidebarWidth(w: number): void {
  sidebarWidth = w;
}

export function setChatWidth(w: number): void {
  chatWidth = w;
}

export function setTocWidth(w: number): void {
  tocWidth = w;
}

export function initPanelWidths(): void {
  const saved = loadPanelWidths();
  if (saved) {
    sidebarWidth = saved.sidebar;
    chatWidth = saved.chat;
    tocWidth = saved.toc;
  }
}

export function isSidebarCollapsed(): boolean {
  return sidebarCollapsed;
}

export function setSidebarCollapsed(collapsed: boolean): void {
  sidebarCollapsed = collapsed;
}

export function toggleSidebarCollapsed(): void {
  if (sidebarCollapsed) {
    sidebarCollapsed = false;
    sidebarWidth = savedSidebarWidth;
  } else {
    savedSidebarWidth = sidebarWidth;
    sidebarCollapsed = true;
  }
}

function currentWidths() {
  return { sidebar: sidebarWidth, chat: chatWidth, toc: tocWidth };
}

export function handleSidebarResize(delta: number, totalWidth: number): void {
  sidebarWidth = clampResize('sidebar', sidebarWidth + delta, chatWidth, totalWidth);
  savePanelWidths(currentWidths());
}

export function handleChatResize(delta: number, totalWidth: number): void {
  chatWidth = clampResize('chat', chatWidth + delta, sidebarWidth, totalWidth);
  savePanelWidths(currentWidths());
}

export function handleTocResize(delta: number): void {
  tocWidth = clampWidth(tocWidth + delta, TOC_MIN, TOC_MAX);
  savePanelWidths(currentWidths());
}
