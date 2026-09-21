import { describe, it, expect, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { server } from "@/tests/msw";
import { defaultHandlers } from "@/tests/handlers";
import { useTerminalStore, type TerminalSession } from "@/stores/terminalStore";

const { default: TerminalInstance } = await import("../TerminalInstance");

function makeSession(
  overrides: Partial<TerminalSession> = {},
): TerminalSession {
  return {
    id: "session-1",
    deviceUid: "device-uid",
    deviceName: "my-device",
    username: "root",
    password: "secret",
    state: "docked",
    connectionStatus: "connecting",
    ...overrides,
  };
}

function renderTerminal() {
  return render(<TerminalInstance session={makeSession()} visible />, {
    wrapper: createTestWrapper({ initialEntries: ["/devices"] }),
  });
}

beforeEach(() => {
  server.use(
    ...defaultHandlers,
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
  );
  seedAuthStore();
  useTerminalStore.setState({ sessions: [makeSession()] });
});

describe("TerminalInstance", () => {
  it("tells an observer their role is the reason, and offers no retry", async () => {
    seedAuthStore({ role: "observer" });
    server.use(
      http.post("*/ws/ssh/session", () =>
        HttpResponse.json(null, { status: 403 }),
      ),
    );
    renderTerminal();

    const banner = await screen.findByRole("alert");

    expect(banner).toHaveTextContent(
      "Your role does not allow connecting to devices.",
    );
    expect(
      screen.queryByRole("button", { name: "Retry" }),
    ).not.toBeInTheDocument();
  });

  it("offers a retry when the session fails for any other reason", async () => {
    server.use(
      http.post("*/ws/ssh/session", () =>
        HttpResponse.json(null, { status: 500 }),
      ),
    );
    renderTerminal();

    const banner = await screen.findByRole("alert");

    expect(banner).toHaveTextContent("Connection failed");
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
