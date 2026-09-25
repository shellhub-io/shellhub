/**
 * The window controls the desktop app lets the console call. The app draws no title bar of its
 * own, so these are the only way to move, resize or close it.
 */
export interface DesktopWindow {
  close: () => Promise<void>;
  minimize: () => Promise<void>;
  toggleMaximize: () => Promise<void>;
}

declare global {
  interface Window {
    __TAURI__?: { window: { getCurrentWindow: () => DesktopWindow } };
  }
}

/**
 * The desktop window the console runs in, or undefined in a browser.
 */
export function desktopWindow(): DesktopWindow | undefined {
  return window.__TAURI__?.window.getCurrentWindow();
}
