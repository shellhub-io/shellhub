import { formatDistanceToNow, format } from "date-fns";

export function formatRelative(dateStr: string): string {
  if (!dateStr) return "—";
  return formatDistanceToNow(new Date(dateStr), { addSuffix: true });
}

export function formatExpiry(expiresIn: number): string {
  if (expiresIn <= 0) return "Never";
  return format(new Date(expiresIn * 1000), "MMM d, yyyy");
}

export function formatDateFull(dateStr: string): string {
  if (!dateStr) return "—";
  return format(new Date(dateStr), "MMM d, yyyy, HH:mm");
}

export function formatDateShort(dateStr: string): string {
  if (!dateStr) return "—";
  return format(new Date(dateStr), "MMM d, yyyy");
}

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
