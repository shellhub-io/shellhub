import type { BadgeColor } from "@shellhub/design-system/primitives";

import type { Session } from "../client";

type SessionEvents = NonNullable<Session["events"]>;
type SessionEvent = NonNullable<SessionEvents["items"]>[number];

/**
 * The label for a session's device: the device's name, else the first 8 characters of the
 * session's device_uid, else "" when the session carries neither.
 */
export function sessionDeviceName(session: Session): string {
  return session.device?.name ?? (session.device_uid ?? "").substring(0, 8);
}

/**
 * Whether the session had a terminal, which is what makes its recording playable.
 */
export function sessionHasTerminal(session: Session): boolean {
  return (session.events?.types ?? []).includes("pty-req");
}

const CHANNEL_STARTS = new Set<string | undefined>([
  "pty-req",
  "shell",
  "exec",
  "subsystem",
]);

/**
 * The seats that may have a recording of their own, lowest first: those whose timeline shows a
 * terminal request, and those whose timeline does not show how their channel started. The server
 * caps the timeline at its oldest events, so a seat whose start fell past the cap may still have
 * asked for a terminal, and one that did not is offered too. It is empty when the session carries
 * no timeline, as on the sessions list or when the server could not read it.
 */
export function terminalSeats(session: Session): number[] {
  const items = session.events?.items;
  if (!items) return [];
  const seatOf = (event: SessionEvent) => event.seat ?? 0;
  const started = new Set(
    items.filter((event) => CHANNEL_STARTS.has(event.type)).map(seatOf),
  );
  const seats = [
    ...items.filter((event) => event.type === "pty-req").map(seatOf),
    ...(session.events?.seats ?? []).filter((seat) => !started.has(seat)),
  ];
  return [...new Set(seats)].sort((a, b) => a - b);
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
