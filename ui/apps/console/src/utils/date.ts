import { formatDistanceToNow, format } from "date-fns";

/**
 * Formats a timestamp as a distance from now ("3 hours ago"). An empty string renders as an em
 * dash, so a missing date reads as absent rather than as the epoch.
 */
export function formatRelative(dateStr: string): string {
  if (!dateStr) return "—";
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
}

/**
 * Formats a Unix expiry as a date. Zero or negative means the credential does not expire, and
 * renders as "Never" — not as a date in 1970.
 */
export function formatExpiry(expiresIn: number): string {
  if (expiresIn <= 0) return "Never";
  return format(new Date(expiresIn * 1000), "MMM d, yyyy");
}

/**
 * Formats a timestamp as date and time, for where the exact moment matters. Empty renders as an
 * em dash.
 */
export function formatDateFull(dateStr: string): string {
  if (!dateStr) return "—";
  return format(new Date(dateStr), "MMM d, yyyy, HH:mm");
}

/**
 * Formats a timestamp as a date alone, for a column too narrow to carry the time. Empty renders
 * as an em dash.
 */
export function formatDateShort(dateStr: string): string {
  if (!dateStr) return "—";
  return format(new Date(dateStr), "MMM d, yyyy");
}

/**
 * How long a session has run. An active session is measured to now, so the value grows as it is
 * re-rendered; a closed one is measured to lastSeen. Under a second renders as an em dash,
 * since a duration rounded to 0s reads as an error rather than as brevity.
 */
export function formatDuration(
  startedAt: string,
  lastSeen: string,
  active: boolean,
): string {
  const start = new Date(startedAt).getTime();
  const end = active ? Date.now() : new Date(lastSeen).getTime();
  const secs = Math.max(0, Math.floor((end - start) / 1000));
  if (secs === 0) return "—";
  if (secs < 60) return `${secs}s`;
  if (secs < 3600) return `${Math.floor(secs / 60)}m ${secs % 60}s`;
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  return `${h}h ${m}m`;
}
