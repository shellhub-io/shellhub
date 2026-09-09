import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import Sessions from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSession } from "@/tests/factories";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("../SessionPlayerDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div data-testid="player-dialog">
        <button type="button" onClick={onClose}>
          Close Player
        </button>
      </div>
    ) : null,
}));

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
  setSessions([]);
  server.use(
    http.post(
      "*/api/sessions/:uid/close",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.get("*/api/sessions/:uid/records/:seat", () =>
      HttpResponse.json({}, { status: 404 }),
    ),
  );
});

describe("Sessions", () => {
  describe("initial load", () => {
    it("shows loading state while fetching", () => {
      server.use(
        http.get("*/api/sessions", () => new Promise(() => {})),
      );
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
        http.get("*/api/sessions/:uid/records/:seat", () =>
          new Promise(() => {}),
        ),
      );
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();

      const btn = await screen.findByTitle("Play recording");
      expect(btn).not.toBeDisabled();

      await user.click(btn);

      await waitFor(() => expect(btn).toBeDisabled());
    });

    it("does not show the player dialog when there are no logs", async () => {
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();
      await screen.findByTitle("Play recording");
      expect(screen.queryByTestId("player-dialog")).not.toBeInTheDocument();
    });

    it("opens the player after recording loads and closes it on dismiss", async () => {
      const user = userEvent.setup();
      server.use(
        http.get("*/api/sessions/:uid/records/:seat", () =>
          HttpResponse.json("asciicast-data"),
        ),
      );
      setSessions([mockSession({ uid: "session-1", recorded: true })]);
      renderSessions();

      await user.click(await screen.findByTitle("Play recording"));
      expect(await screen.findByTestId("player-dialog")).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: "Close Player" }));

      await waitFor(() =>
        expect(screen.queryByTestId("player-dialog")).not.toBeInTheDocument(),
      );
    });
  });
});
