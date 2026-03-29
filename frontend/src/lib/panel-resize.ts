export const SIDEBAR_MIN = 180;
export const SIDEBAR_MAX = 640;
export const SIDEBAR_DEFAULT = 280;

export const CHAT_MIN = 240;
export const CHAT_MAX = 800;
export const CHAT_DEFAULT = 360;

export const TOC_MIN = 160;
export const TOC_MAX = 480;
export const TOC_DEFAULT = 260;

export const CENTER_MIN = 300;

const STORAGE_KEY = 'panel-widths';

export interface PanelWidths {
  sidebar: number;
  chat: number;
  toc: number;
}

export function clampWidth(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}

export function clampResize(
  panel: 'sidebar' | 'chat',
  requested: number,
  otherWidth: number,
  totalWidth: number
): number {
  const [min, max] = panel === 'sidebar'
    ? [SIDEBAR_MIN, SIDEBAR_MAX]
    : [CHAT_MIN, CHAT_MAX];

  // Max allowed by center minimum constraint
  const maxByCenter = totalWidth - otherWidth - CENTER_MIN;

  const effectiveMax = Math.min(max, maxByCenter);
  return Math.max(min, Math.min(effectiveMax, requested));
}

export function savePanelWidths(widths: PanelWidths): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(widths));
}

export function loadPanelWidths(): PanelWidths | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;

  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return null;
  }

  if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) return null;

  const obj = parsed as Record<string, unknown>;
  if (typeof obj.sidebar !== 'number' || typeof obj.chat !== 'number') return null;

  const toc = typeof obj.toc === 'number' ? obj.toc : TOC_DEFAULT;

  return {
    sidebar: clampWidth(obj.sidebar, SIDEBAR_MIN, SIDEBAR_MAX),
    chat: clampWidth(obj.chat, CHAT_MIN, CHAT_MAX),
    toc: clampWidth(toc, TOC_MIN, TOC_MAX)
  };
}
