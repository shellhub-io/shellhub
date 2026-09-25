import type { BadgeColor } from "@shellhub/design-system/primitives";

import type { Session } from "../client";

type SessionEvents = NonNullable<Session["events"]>;
type SessionEvent = NonNullable<SessionEvents["items"]>[number];

/**
 * What a session is called on its tab and wherever it needs a short name: its device's name, or
 * the start of the device uid when the device is gone.
 */
export function sessionTitle(session: Session): string {
  return session.device?.name ?? (session.device_uid ?? "").substring(0, 8);
}

/**
 * Whether the session had a terminal, which is what makes its recording playable.
 */
export function sessionHasTerminal(session: Session): boolean {
  return (session.events?.types ?? []).includes("pty-req");
}

/**
 * What a session did, as the label and colour of its badge: an SFTP transfer, a single exec, or an
 * interactive shell. It is null when the session opened no channel. This is the one place that
 * decision is made; every screen renders it through SessionTypeBadge.
 */
export function sessionType(
  session: Session,
): { label: string; color: BadgeColor } | null {
  switch (session.events?.first) {
    case "subsystem":
      return { label: "sftp", color: "cyan" };
    case "exec":
      return { label: "exec", color: "yellow" };
    case "shell":
    case "pty-req":
      return { label: "shell", color: "primary" };
    default:
      return null;
  }
}

/**
 * The terminal the client asked for, read from the pty request the session recorded. It is
 * undefined when no terminal was requested, and on the sessions list, which carries no timeline.
 *
 * A multi-seat session records one pty request per seat; this answers for the first, which is the
 * single-terminal case the detail page is built around.
 */
export function sessionTerminal(session: Session): string | undefined {
  const ptyRequest = (session.events?.items ?? []).find(
    (event: SessionEvent) => event.type === "pty-req",
  );

  const payload: unknown = ptyRequest?.data;
  if (typeof payload !== "object" || payload === null) return undefined;

  const term = (payload as Record<string, unknown>).term;

  return typeof term === "string" && term !== "" ? term : undefined;
}
