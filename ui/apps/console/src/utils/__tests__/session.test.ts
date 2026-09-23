import { describe, expect, it } from "vitest";

import type { Session } from "@/client";
import { sessionHasTerminal } from "@/utils/session";

const session = (events: Session["events"]) => ({ events }) as Session;

describe("a terminal session that also ran commands", () => {
  it("stays playable, so its recording is still reachable", () => {
    expect(
      sessionHasTerminal(
        session({ types: ["pty-req", "shell", "exec", "exit-status"], seats: [] }),
      ),
    ).toBe(true);
  });
});
