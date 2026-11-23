import { describe, expect, it } from "vitest";

import type { Session } from "@/client";
import {
  sessionHasTerminal,
  sessionTerminal,
  sessionDeviceName,
  sessionType,
  terminalSeats,
} from "@/utils/session";

const session = (events: Session["events"]) => ({ events }) as Session;

const multiplexed = session({
  first: "pty-req",
  types: ["pty-req", "shell", "exec", "exit-status"],
});

describe("sessionHasTerminal", () => {
  it("holds for a terminal session that also ran commands, so its recording stays reachable", () => {
    expect(sessionHasTerminal(multiplexed)).toBe(true);
  });

  it("does not hold for a session that asked for no terminal", () => {
    expect(
      sessionHasTerminal(session({ first: "exec", types: ["exec"] })),
    ).toBe(false);
  });
});

describe("sessionType", () => {
  it("badges a terminal session that also ran commands by what it opened with", () => {
    expect(sessionType(multiplexed)?.label).toBe("shell");
  });

  it.each([
    ["subsystem", "sftp"],
    ["exec", "exec"],
    ["shell", "shell"],
    ["pty-req", "shell"],
  ] as const)("badges a session that opened with %s as %s", (first, label) => {
    expect(sessionType(session({ first, types: [first] }))?.label).toBe(label);
  });

  it("has no badge for a session that opened no channel", () => {
    expect(sessionType(session({ types: ["env"] }))).toBeNull();
  });
});

describe("sessionTerminal", () => {
  const withItems = (items: NonNullable<Session["events"]>["items"]) =>
    session({ items });

  it("reads the requested terminal from the pty request, past the events before it", () => {
    expect(
      sessionTerminal(
        withItems([
          { type: "env", data: { name: "LANG" } },
          {
            type: "pty-req",
            data: { term: "xterm-256color", columns: 141, rows: 30 },
          },
        ]),
      ),
    ).toBe("xterm-256color");
  });

  it("answers for the first seat, which is the terminal the detail page shows", () => {
    expect(
      sessionTerminal(
        withItems([
          { type: "pty-req", data: { term: "xterm" } },
          { type: "pty-req", data: { term: "screen" } },
        ]),
      ),
    ).toBe("xterm");
  });

  it.each([
    [
      "no terminal was requested",
      withItems([{ type: "exec", data: { command: "uptime" } }]),
    ],
    [
      "the payload carries no terminal name",
      withItems([{ type: "pty-req", data: {} }]),
    ],
    [
      "the timeline is absent, as it is on the list",
      session({ types: ["pty-req"] }),
    ],
    [
      "the payload is not an object, as a request with no body records",
      withItems([{ type: "pty-req", data: "" }]),
    ],
  ])("is undefined when %s", (_, input) => {
    expect(sessionTerminal(input)).toBeUndefined();
  });
});

describe("sessionDeviceName", () => {
  it.each([
    [{ device: { name: "web-01" }, device_uid: "abcdef0123456789" }, "web-01"],
    [{ device_uid: "abcdef0123456789" }, "abcdef01"],
    [{}, ""],
  ])("names %j as %j", (fields, name) => {
    expect(sessionDeviceName(fields as Session)).toBe(name);
  });
});

describe("terminalSeats", () => {
  const event = (type: string, seat: number) => ({
    type,
    seat,
    timestamp: "2024-01-01T00:00:00Z",
    data: {},
  });

  it.each([
    { name: "no timeline", items: undefined, seats: [0, 1], expected: [] },
    {
      name: "one terminal",
      items: [event("pty-req", 0)],
      seats: [0],
      expected: [0],
    },
    {
      name: "two terminals, listed out of order",
      items: [event("pty-req", 1), event("pty-req", 0)],
      seats: [0, 1],
      expected: [0, 1],
    },
    {
      name: "a command with no terminal",
      items: [event("exec", 0)],
      seats: [0],
      expected: [],
    },
    {
      name: "a seat the capped timeline does not show",
      items: [event("pty-req", 0)],
      seats: [0, 1],
      expected: [0, 1],
    },
    {
      name: "a seat whose terminal request fell past the cap",
      items: [event("pty-req", 0), event("auth-agent-req@openssh.com", 1)],
      seats: [0, 1],
      expected: [0, 1],
    },
  ])("offers the seats of $name", ({ items, seats, expected }) => {
    expect(terminalSeats(session({ items, seats }))).toEqual(expected);
  });
});
