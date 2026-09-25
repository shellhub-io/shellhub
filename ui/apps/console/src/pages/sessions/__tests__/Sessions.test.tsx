import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import Sessions from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockSession } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useTerminalStore } from "@/stores/terminalStore";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function setSessions(
  sessions: ReturnType<typeof mockSession>[],
  total?: number,
) {
  server.use(
    http.get("*/api/sessions", () =>
      jsonWithTotal(sessions, total ?? sessions.length),
    ),
  );
}

function renderSessions(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Sessions />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useTerminalStore.setState({ sessions: [], recordings: [] });
  setSessions([]);
  server.use(
    http.post(
      "*/api/sessions/:uid/close",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.get("*/api/sessions/:uid/records/:seat", () =>
      HttpResponse.json({}, { status: 404 }),
    ),
    http.get("*/api/namespaces/api-key", () => HttpResponse.json([])),
  );
});

describe("Sessions", () => {
  describe("initial load", () => {
    it("shows loading state while fetching", () => {
      server.use(http.get("*/api/sessions", () => new Promise(() => {})));
      renderSessions();
      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });

    it("shows empty state when there are no sessions", async () => {
      renderSessions();
      expect(await screen.findByText("No sessions found")).toBeInTheDocument();
    });
  });

  describe("session row", () => {
    it("navigates to session detail when a row is clicked", async () => {
      const user = userEvent.setup();
      setSessions([mockSession({ uid: "session-abc" })]);
      renderSessions();

      await user.click(await screen.findByText("root"));

      expect(mockNavigate).toHaveBeenCalledWith("/sessions/session-abc");
    });
  });

  describe("logsError banner", () => {
    it("shows an error banner when fetching recording fails", async () => {
      const user = userEvent.setup();
      setSessions([mockSession({ uid: "s-1", recorded: true })]);
      renderSessions();

      await user.click(await screen.findByTitle("Play recording"));

      expect(
        await screen.findByText("Failed to load recording"),
      ).toBeInTheDocument();
    });
  });

  describe("play recording", () => {
    it("disables the play button while the recording is loading", async () => {
      const user = userEvent.setup();
      server.use(
        http.get(
          "*/api/sessions/:uid/records/:seat",
          () => new Promise(() => {}),
        ),
      );
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();

      const btn = await screen.findByTitle("Play recording");
      expect(btn).not.toBeDisabled();

      await user.click(btn);

      await waitFor(() => expect(btn).toBeDisabled());
    });

    it("opens the recording in its own tab once it loads", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("*/api/sessions/:uid/records/:seat", () =>
          HttpResponse.text("asciicast-data"),
        ),
      );
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();

      await user.click(await screen.findByTitle("Play recording"));

      await waitFor(() =>
        expect(useTerminalStore.getState().recordings).toEqual([
          expect.objectContaining({
            id: "session-1",
            logs: "asciicast-data",
            shown: true,
          }),
        ]),
      );
    });

    it("opens no tab when the recording fails to load", async () => {
      const user = userEvent.setup();
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();

      await user.click(await screen.findByTitle("Play recording"));

      expect(
        await screen.findByText("Failed to load recording"),
      ).toBeInTheDocument();
      expect(useTerminalStore.getState().recordings).toEqual([]);
    });
  });

  describe("principal", () => {
    beforeEach(() => {
      seedAuthStore({ tenant: "tenant-456" });
      server.use(
        http.get("*/api/namespaces/tenant-456", () =>
          HttpResponse.json(mockNamespace()),
        ),
      );
    });

    it("names who opened each session", async () => {
      setSessions([
        mockSession({ principal: { kind: "user", id: "user-123" } }),
      ]);

      renderSessions();

      expect(await screen.findByText("admin@test.com")).toBeInTheDocument();
    });

    it("names nobody for a session under the legacy access model", async () => {
      setSessions([mockSession({ principal: undefined })]);

      renderSessions();

      await screen.findByText(mockSession().username);
      expect(screen.queryByText("admin@test.com")).not.toBeInTheDocument();
    });
  });
});
