import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import Sessions from "../index";

/** Exposes the current search string from inside the MemoryRouter. */
function LocationProbe({
  onLocation,
}: {
  onLocation: (search: string) => void;
}) {
  const loc = useLocation();
  onLocation(loc.search);
  return null;
}

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/hooks/useSessions", () => ({
  useSessions: vi.fn(),
}));

vi.mock("@/hooks/useSessionMutations", () => ({
  useCloseSession: vi.fn(),
}));

vi.mock("@/hooks/useSessionRecording", () => ({
  useSessionRecording: vi.fn(),
}));

vi.mock("../SessionPlayerDialog", () => ({
  default: ({ open, onClose }: { open: boolean; onClose: () => void }) =>
    open ? (
      <div data-testid="player-dialog">
        <button type="button" onClick={onClose}>Close Player</button>
      </div>
    ) : null,
}));

import { useSessions } from "@/hooks/useSessions";
import { useCloseSession } from "@/hooks/useSessionMutations";
import { useSessionRecording } from "@/hooks/useSessionRecording";

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

const mockFetchLogs = vi.fn().mockResolvedValue(undefined);
const mockClearLogs = vi.fn();

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeSession(overrides: Record<string, any> = {}) {
  return {
    uid: "session-1",
    device_uid: "device-1",
    device: { uid: "device-1", name: "my-device", online: true },
    tenant_id: "tenant-1",
    username: "root",
    ip_address: "192.168.1.1",
    started_at: "2024-01-01T00:00:00Z",
    last_seen: "2024-01-01T00:01:00Z",
    active: false,
    authenticated: true,
    recorded: false,
    type: "shell",
    term: "xterm",
    position: { latitude: 0, longitude: 0 },
    events: { types: [], seats: [] },
    ...overrides,
  };
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeSessions(overrides: Record<string, any> = {}) {
  return {
    sessions: [],
    totalCount: 0,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
    ...overrides,
  };
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeRecording(overrides: Record<string, any> = {}) {
  return {
    logs: null,
    isLoading: false,
    error: null,
    fetchLogs: mockFetchLogs,
    clearLogs: mockClearLogs,
    ...overrides,
  };
}

function renderSessions(initialEntries: string[] = ["/"]) {
  let lastSearch = "";
  const result = render(
    <MemoryRouter initialEntries={initialEntries}>
      <Sessions />
      <LocationProbe onLocation={(s) => { lastSearch = s; }} />
    </MemoryRouter>,
  );
  return { ...result, getSearch: () => lastSearch };
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useSessions).mockReturnValue(makeSessions());
  vi.mocked(useCloseSession).mockReturnValue({ mutateAsync: vi.fn(), isPending: false, error: null } as unknown as ReturnType<typeof useCloseSession>);
  vi.mocked(useSessionRecording).mockReturnValue(makeRecording());
});

/* ------------------------------------------------------------------ */
/* Tests                                                               */
/* ------------------------------------------------------------------ */

describe("Sessions", () => {
  describe("initial load", () => {
    it("shows loading state while fetching", () => {
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({ isLoading: true }),
      );

      renderSessions();

      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });

    it("shows empty state when there are no sessions", () => {
      renderSessions();

      expect(screen.getByText("No sessions found")).toBeInTheDocument();
    });
  });

  describe("session row", () => {
    it("navigates to session detail when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: [makeSession({ uid: "session-abc" })],
        }),
      );

      renderSessions();

      await user.click(screen.getByText("root"));

      expect(mockNavigate).toHaveBeenCalledWith("/sessions/session-abc");
    });
  });

  describe("logsError banner", () => {
    it("shows an error banner when logsError is set", () => {
      vi.mocked(useSessionRecording).mockReturnValue(
        makeRecording({ error: "Failed to load recording" }),
      );

      renderSessions();

      expect(screen.getByText("Failed to load recording")).toBeInTheDocument();
    });

    it("does not show the error banner when error is null", () => {
      renderSessions();

      expect(
        screen.queryByText("Failed to load recording"),
      ).not.toBeInTheDocument();
    });
  });

  describe("play recording", () => {
    it("calls fetchLogs with the session uid when Play is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: [makeSession({ uid: "session-1", recorded: true })],
        }),
      );

      renderSessions();

      await user.click(screen.getByTitle("Play recording"));

      expect(mockFetchLogs).toHaveBeenCalledWith("session-1");
    });

    it("disables the play button after clicking while logs are loading", async () => {
      const user = userEvent.setup();
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: [makeSession({ uid: "session-1", recorded: true })],
        }),
      );
      vi.mocked(useSessionRecording).mockReturnValue(
        makeRecording({ isLoading: true }),
      );

      renderSessions();

      const btn = screen.getByTitle("Play recording");
      // Not disabled yet — playTarget is null
      expect(btn).not.toBeDisabled();

      await user.click(btn);

      // After click, playTarget === "session-1" and logsLoading is true
      expect(btn).toBeDisabled();
    });

    it("does not show the player dialog when there are no logs", () => {
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: [makeSession({ uid: "session-1", recorded: true })],
        }),
      );

      renderSessions();

      expect(screen.queryByTestId("player-dialog")).not.toBeInTheDocument();
    });

    it("calls clearLogs when the player dialog is closed", async () => {
      const user = userEvent.setup();
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: [makeSession({ uid: "session-1", recorded: true })],
        }),
      );
      vi.mocked(useSessionRecording).mockReturnValue(
        makeRecording({ logs: "recorded-session-data", isLoading: false }),
      );

      renderSessions();

      await user.click(screen.getByTitle("Play recording"));
      expect(screen.getByTestId("player-dialog")).toBeInTheDocument();

      await user.click(screen.getByRole("button", { name: "Close Player" }));

      expect(mockClearLogs).toHaveBeenCalledOnce();
    });
  });

  // ── URL-driven state (usePaginatedListState adoption) ────────────────────────

  describe("URL hydration", () => {
    it("calls useSessions with page hydrated from ?page=3", () => {
      renderSessions(["/?page=3"]);
      expect(vi.mocked(useSessions)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 3 }),
      );
    });

    it("calls useSessions with page=1 when URL has no page param", () => {
      renderSessions(["/"]);
      expect(vi.mocked(useSessions)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });
  });

  describe("URL writes", () => {
    it("writes ?page=2 to the URL when the user navigates to page 2", async () => {
      const user = userEvent.setup();
      vi.mocked(useSessions).mockReturnValue(
        makeSessions({
          sessions: Array.from({ length: 10 }, (_, i) =>
            makeSession({ uid: `s-${i}`, username: `u-${i}` }),
          ),
          totalCount: 30,
        }),
      );
      renderSessions();

      await user.click(screen.getByRole("button", { name: "Next page" }));

      expect(vi.mocked(useSessions)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );
    });

    it("omits ?page from the URL when on the default page 1", () => {
      const { getSearch } = renderSessions(["/"]);
      const sp = new URLSearchParams(getSearch());
      expect(sp.get("page")).toBeNull();
    });
  });
});
