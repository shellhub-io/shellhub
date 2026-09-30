/**
 * What setup hands the dashboard through the route state when it signs the new admin in, so the
 * trail can show the steps already behind them. Nothing else carries it: a reload drops it.
 */
export interface FirstRunEntry {
  fromSetup: true;
  survey: boolean;
}

/**
 * The route state setup navigates with, handing the entry over to the dashboard.
 */
export function firstRunEntryState(survey: boolean): {
  firstRun: FirstRunEntry;
} {
  return { firstRun: { fromSetup: true, survey } };
}

/**
 * Reads the setup hand-over from a route's state, or null when the user did not come from setup.
 */
export function readFirstRunEntry(state: unknown): FirstRunEntry | null {
  if (typeof state !== "object" || state === null) return null;
  const entry = (state as { firstRun?: unknown }).firstRun;
  if (typeof entry !== "object" || entry === null) return null;
  const { fromSetup, survey } = entry as Partial<FirstRunEntry>;
  if (fromSetup !== true) return null;
  return { fromSetup: true, survey: survey === true };
}
