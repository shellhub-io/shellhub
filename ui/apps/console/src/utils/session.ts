import type { Session } from "../client";

/**
 * Whether the session had a terminal, which is what makes its recording playable.
 */
export function sessionHasTerminal(session: Session): boolean {
  return (session.events?.types ?? []).includes("pty-req");
}

/**
 * The badge for a session, derived from the event types it recorded: an SFTP transfer, a single
 * exec, or null for an ordinary interactive shell, which needs no badge.
 */
export function sessionType(
  session: Session,
): { label: string; color: string } | null {
  const types = session.events?.types ?? [];
  if (types.includes("subsystem"))
    return {
      label: "sftp",
      color: "text-accent-cyan bg-accent-cyan/10 border-accent-cyan/20",
    };
  if (types.includes("exec"))
    return {
      label: "exec",
      color: "text-accent-yellow bg-accent-yellow/10 border-accent-yellow/20",
    };
  if (types.includes("shell") || types.includes("pty-req"))
    return {
      label: "shell",
      color: "text-primary bg-primary/10 border-primary/20",
    };
  return null;
}
