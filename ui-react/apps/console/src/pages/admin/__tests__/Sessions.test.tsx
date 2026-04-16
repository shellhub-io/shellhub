import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/hooks/useAdminSessionsList", () => ({
  useAdminSessionsList: vi.fn(),
}));

import { useAdminSessionsList } from "@/hooks/useAdminSessionsList";
import AdminSessions from "../Sessions";

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeSession(overrides: Record<string, any> = {}) {
  return {
    uid: "session-1",
    device_uid: "device-1",
    device: { uid: "device-1", name: "web-server-01", online: true, info: { id: "ubuntu" } },
    username: "root",
    ip_address: "192.168.0.1",
    started_at: "2024-01-01T00:00:00Z",
    last_seen: "2024-01-01T01:00:00Z",
    active: false,
    authenticated: true,
    ...overrides,
  };
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function setupHook(overrides: Record<string, any> = {}) {
  vi.mocked(useAdminSessionsList).mockReturnValue({
    sessions: [],
    totalCount: 0,
    isLoading: false,
    error: null,
    ...overrides,
  } as ReturnType<typeof useAdminSessionsList>);
}

function renderPage() {
  return render(
    <MemoryRouter>
      <AdminSessions />
    </MemoryRouter>,
  );
}

/* ------------------------------------------------------------------ */
/* Tests                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  setupHook();
});

describe("AdminSessions", () => {
  describe("loading state", () => {
    it("shows a loading spinner while fetching", () => {
      setupHook({ isLoading: true });
      renderPage();
      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });

    it("does not render session rows while loading", () => {
      setupHook({ isLoading: true });
      renderPage();
      expect(screen.queryByText("root")).not.toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it("shows 'No sessions found' when there are no sessions", () => {
      renderPage();
      expect(screen.getByText("No sessions found")).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders the error banner with role='alert'", () => {
      setupHook({ error: new Error("Server error. Please try again later.") });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("displays the error message in the banner", () => {
      setupHook({ error: new Error("You don't have permission to view sessions.") });
      renderPage();
      expect(screen.getByRole("alert")).toHaveTextContent(
        "You don't have permission to view sessions.",
      );
    });

    it("does not show the error banner when there is no error", () => {
      renderPage();
      expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    });
  });

  describe("session rows", () => {
    it("renders one row per session", () => {
      setupHook({
        sessions: [
          makeSession({ uid: "session-1", username: "root" }),
          makeSession({ uid: "session-2", username: "admin" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("root")).toBeInTheDocument();
      expect(screen.getByText("admin")).toBeInTheDocument();
    });

    it("renders the device name via DeviceChip", () => {
      setupHook({ sessions: [makeSession()], totalCount: 1 });
      renderPage();
      expect(screen.getByText("web-server-01")).toBeInTheDocument();
    });

    it("renders the truncated session uid", () => {
      setupHook({ sessions: [makeSession({ uid: "abcdef1234567890" })], totalCount: 1 });
      renderPage();
      expect(screen.getByText("abcdef1234")).toBeInTheDocument();
    });

    it("renders the IP address", () => {
      setupHook({ sessions: [makeSession({ ip_address: "10.0.0.1" })], totalCount: 1 });
      renderPage();
      expect(screen.getByText("10.0.0.1")).toBeInTheDocument();
    });

    it("navigates to session detail when a row is clicked", async () => {
      const user = userEvent.setup();
      setupHook({ sessions: [makeSession({ uid: "session-abc" })], totalCount: 1 });
      renderPage();

      await user.click(screen.getByText("root"));

      expect(mockNavigate).toHaveBeenCalledWith("/admin/sessions/session-abc");
    });
  });

  describe("active indicator", () => {
    it("renders a green dot for active sessions", () => {
      setupHook({ sessions: [makeSession({ active: true })], totalCount: 1 });
      renderPage();
      const dot = document.querySelector(".bg-accent-green");
      expect(dot).toBeInTheDocument();
    });

    it("renders a muted dot for inactive sessions", () => {
      setupHook({ sessions: [makeSession({ active: false })], totalCount: 1 });
      renderPage();
      const dot = document.querySelector(".bg-text-muted\\/40");
      expect(dot).toBeInTheDocument();
    });
  });

  describe("authentication indicator", () => {
    it("renders the 'Authenticated' shield for authenticated sessions", () => {
      setupHook({ sessions: [makeSession({ authenticated: true })], totalCount: 1 });
      renderPage();
      expect(screen.getByTitle("Authenticated")).toBeInTheDocument();
    });

    it("renders the 'Not authenticated' shield for unauthenticated sessions", () => {
      setupHook({ sessions: [makeSession({ authenticated: false })], totalCount: 1 });
      renderPage();
      expect(screen.getAllByTitle("Not authenticated").length).toBeGreaterThan(0);
    });

    it("shows the warning icon in the username cell for unauthenticated sessions", () => {
      setupHook({ sessions: [makeSession({ authenticated: false })], totalCount: 1 });
      renderPage();
      // ExclamationTriangleIcon has title "Not authenticated"
      expect(screen.getAllByTitle("Not authenticated").length).toBeGreaterThanOrEqual(2);
    });
  });

  describe("device fallback", () => {
    it("shows the truncated device_uid when device object is missing", () => {
      setupHook({
        sessions: [makeSession({ device: null, device_uid: "abcd1234efgh" })],
        totalCount: 1,
      });
      renderPage();
      expect(screen.getByText("abcd1234")).toBeInTheDocument();
    });
  });

  describe("pagination", () => {
    it("renders pagination when totalCount > perPage", () => {
      setupHook({
        sessions: Array.from({ length: 10 }, (_, i) =>
          makeSession({ uid: `session-${i}`, username: `user-${i}` }),
        ),
        totalCount: 25,
      });
      renderPage();
      // Pagination renders page info
      expect(screen.getByText(/25/)).toBeInTheDocument();
    });
  });
});
