import { formatDistanceToNow, format, differenceInSeconds } from "date-fns";

export function formatDate(dateStr: string): string {
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
}

export function formatRelative(dateStr: string): string {
  if (!dateStr) return "\u2014";
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
}

export function formatExpiry(expiresIn: number): string {
  if (expiresIn <= 0) return "Never";
  return format(new Date(expiresIn * 1000), "MMM d, yyyy");
}

export function formatDateFull(dateStr: string): string {
  if (!dateStr) return "\u2014";
  return format(new Date(dateStr), "MMM d, yyyy, hh:mm a");
}

export function formatDateShort(dateStr: string): string {
  if (!dateStr) return "\u2014";
  return format(new Date(dateStr), "MMM d, yyyy");
}

export function formatDuration(
  startedAt: string,
  lastSeen: string,
  active: boolean,
): string {
  const start = new Date(startedAt);
  const end = active ? new Date() : new Date(lastSeen);
  const secs = Math.max(0, differenceInSeconds(end, start));
  if (secs === 0) return "\u2014";
  if (secs < 60) return `${secs}s`;
  if (secs < 3600) return `${Math.floor(secs / 60)}m ${secs % 60}s`;
  const h = Math.floor(secs / 3600);
  const m = Math.floor((secs % 3600) / 60);
  return `${h}h ${m}m`;
}
