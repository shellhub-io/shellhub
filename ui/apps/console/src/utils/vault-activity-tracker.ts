/**
 * Framework-agnostic activity tracker module.
 *
 * Tracks user activity events to drive idle-lock and hidden-lock behaviour
 * for the vault. No React imports — safe to use outside the component tree.
 */

export const ACTIVITY_EVENTS = [
  "mousemove",
  "mousedown",
  "keydown",
  "scroll",
  "touchstart",
] as const;

export const THROTTLE_MS = 1_000;

export interface TrackerOptions {
  /** Milliseconds of inactivity before onIdle fires. 0 = disabled. */
  idleTimeoutMs: number;
  /** When true, start a hidden-grace timer whenever the document becomes hidden. */
  lockOnHidden: boolean;
  /** Milliseconds to wait after the document becomes hidden before calling onIdle. */
  hiddenGraceMs: number;
  /** Called when the idle or hidden-grace timer expires. */
  onIdle: () => void;
}

// ---------------------------------------------------------------------------
// Module-level singleton state
// ---------------------------------------------------------------------------

/** Timestamp (Date.now()) of the last throttled activity reset. 0 = never. */
let lastResetStamp = 0;

let idleTimer: ReturnType<typeof setTimeout> | null = null;
let hiddenTimer: ReturnType<typeof setTimeout> | null = null;

let currentOptions: TrackerOptions | null = null;

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

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

/**
 * Throttled activity handler.
 *
 * The initial stamp is 0, which means `Date.now() - 0 >= THROTTLE_MS` is true
 * immediately after start(), so the very first event always passes through and
 * re-arms the idle timer exactly once.
 */
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

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

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
  // stop() already reset lastResetStamp to 0, guaranteeing the first activity
  // event after start() always passes the throttle check.

  for (const event of ACTIVITY_EVENTS) {
    window.addEventListener(event, onActivity);
  }

  if (opts.lockOnHidden) {
    document.addEventListener("visibilitychange", onVisibilityChange);
  }

  armIdleTimer();
}
