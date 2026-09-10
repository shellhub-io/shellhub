import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import Sessions from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSession } from "@/tests/factories";
import { LocationProbe } from "@/tests/LocationProbe";

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

let lastSessionsUrl: URL | null;

function setSessions(
  sessions: ReturnType<typeof mockSession>[],
  total?: number,
) {
  server.use(
    http.get("*/api/sessions", ({ request }) => {
      lastSessionsUrl = new URL(request.url);
      return jsonWithTotal(sessions, total ?? sessions.length);
    }),
  );
}

function renderSessions(initialEntries: string[] = ["/"]) {
  let lastSearch = "";
  const result = render(
    <MemoryRouter initialEntries={initialEntries}>
      <Sessions />
      <LocationProbe
        onLocation={(s) => {
          lastSearch = s;
        }}
      />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
  return { ...result, getSearch: () => lastSearch };
}

beforeEach(() => {
  vi.clearAllMocks();
  lastSessionsUrl = null;
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

    it("does not show the error banner when error is null", async () => {
      renderSessions();
      await screen.findByText("No sessions found");
      expect(
        screen.queryByText("Failed to load recording"),
      ).not.toBeInTheDocument();
    });
  });

  describe("play recording", () => {
    it("fetches the recording when Play is clicked", async () => {
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
    });

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

  describe("URL hydration", () => {
    it("passes page=3 when URL has ?page=3", async () => {
      renderSessions(["/?page=3"]);
      await waitFor(() => {
        expect(lastSessionsUrl).not.toBeNull();
        expect(lastSessionsUrl!.searchParams.get("page")).toBe("3");
      });
    });

    it("passes page=1 when URL has no page param", async () => {
      renderSessions(["/"]);
      await waitFor(() => {
        expect(lastSessionsUrl).not.toBeNull();
        expect(lastSessionsUrl!.searchParams.get("page")).toBe("1");
      });
    });
  });

  describe("URL writes", () => {
    it("writes ?page=2 to the URL when the user navigates to page 2", async () => {
      const user = userEvent.setup();
      setSessions(
        Array.from({ length: 10 }, (_, i) =>
          mockSession({ uid: `s-${i}`, username: `u-${i}` }),
        ),
        30,
      );
      renderSessions();

      await screen.findByText("u-0");

      await user.click(screen.getByRole("button", { name: "Next page" }));

      await waitFor(() => {
        expect(lastSessionsUrl!.searchParams.get("page")).toBe("2");
      });
    });

    it("omits ?page from the URL when on the default page 1", () => {
      const { getSearch } = renderSessions(["/"]);
      const sp = new URLSearchParams(getSearch());
      expect(sp.get("page")).toBeNull();
    });
  });
});
