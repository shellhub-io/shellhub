/**
 * The DOM events counted as user activity. Chosen to cover pointer, keyboard, scroll and touch,
 * so an idle timer does not fire under someone who is reading and scrolling.
 */
export const ACTIVITY_EVENTS = [
  "mousemove",
  "mousedown",
  "keydown",
  "scroll",
  "touchstart",
] as const;

/**
 * How rarely the idle clock is reset. Without it every mousemove would write, and the tracker
 * would cost more than the thing it guards.
 */
export const THROTTLE_MS = 1_000;

/**
 * How the tracker locks. lockOnHidden also locks when the tab is hidden, after hiddenGraceMs —
 * the grace exists because switching windows briefly hides the tab, and locking on that would
 * make the vault unusable.
 */
export interface TrackerOptions {
  idleTimeoutMs: number;
  lockOnHidden: boolean;
  hiddenGraceMs: number;
  onIdle: () => void;
}

let lastResetStamp = 0;

let idleTimer: ReturnType<typeof setTimeout> | null = null;
let hiddenTimer: ReturnType<typeof setTimeout> | null = null;

let currentOptions: TrackerOptions | null = null;

function fireIdle(): void {
  currentOptions?.onIdle();
}

function clearIdleTimer(): void {
  if (idleTimer !== null) {
    clearTimeout(idleTimer);
    idleTimer = null;
  }
}

function clearHiddenTimer(): void {
  if (hiddenTimer !== null) {
    clearTimeout(hiddenTimer);
    hiddenTimer = null;
  }
}

function armIdleTimer(): void {
  if (!currentOptions || currentOptions.idleTimeoutMs === 0) return;
  clearIdleTimer();
  idleTimer = setTimeout(fireIdle, currentOptions.idleTimeoutMs);
}

function onActivity(): void {
  const now = Date.now();
  if (now - lastResetStamp < THROTTLE_MS) return;
  lastResetStamp = now;
  reset();
}

function onDocumentHidden(): void {
  if (!currentOptions) return;
  hiddenTimer = setTimeout(fireIdle, currentOptions.hiddenGraceMs);
}

function onDocumentVisible(): void {
  clearHiddenTimer();
}

function onVisibilityChange(): void {
  if (document.visibilityState === "hidden") {
    onDocumentHidden();
  } else {
    onDocumentVisible();
  }
}

/**
 * Re-arm the idle timer. Exported so tests can drive it directly.
 * No-op when idleTimeoutMs is 0 or the tracker has not been started.
 */
export function reset(): void {
  armIdleTimer();
}

/**
 * Stop the tracker: clear all timers, remove all listeners, and reset state.
 * Idempotent — safe to call when never started or already stopped.
 */
export function stop(): void {
  clearIdleTimer();
  clearHiddenTimer();

  for (const event of ACTIVITY_EVENTS) {
    window.removeEventListener(event, onActivity);
  }

  document.removeEventListener("visibilitychange", onVisibilityChange);

  currentOptions = null;
  lastResetStamp = 0;
}

/**
 * Start the activity tracker with the given options.
 * Calls stop() first so repeated calls restart cleanly.
 */
export function start(opts: TrackerOptions): void {
  stop();

  currentOptions = opts;

  for (const event of ACTIVITY_EVENTS) {
    window.addEventListener(event, onActivity);
  }

  if (opts.lockOnHidden) {
    document.addEventListener("visibilitychange", onVisibilityChange);
  }

  armIdleTimer();
}
