/**
 * What setup hands the dashboard through the route state when it signs the new admin in, so the
 * trail can show the steps already behind them. Nothing else carries it: a reload drops it.
 */
export interface FirstRunEntry {
  fromSetup: true;
}

/**
 * The route state setup navigates with, handing the entry over to the dashboard.
 */
export function firstRunEntryState(): { firstRun: FirstRunEntry } {
  return { firstRun: { fromSetup: true } };
}

/**
 * Reads the setup hand-over from a route's state, or null when the user did not come from setup.
 */
export function readFirstRunEntry(state: unknown): FirstRunEntry | null {
  if (typeof state !== "object" || state === null) return null;
  const entry = (state as { firstRun?: unknown }).firstRun;
  if (typeof entry !== "object" || entry === null) return null;
  if ((entry as Partial<FirstRunEntry>).fromSetup !== true) return null;
  return { fromSetup: true };
}
