import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { paginatedResponse } from "@/tests/sdk";
import RecentSessionsTable from "../RecentSessionsTable";
import type { Device, Session } from "@/client";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getSessions: vi.fn(),
    getSessionsAdmin: vi.fn(),
  }),
);

vi.mock("@/stores/authStore", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/stores/authStore")>();
  return {
    ...actual,
    useAuthStore: vi.fn((selector: (s: { isAdmin: boolean }) => unknown) =>
      selector({ isAdmin: true }),
    ),
  };
});

function makeSession(overrides: Partial<Session> = {}): Session {
  return {
    uid: "session-1",
    device_uid: "device-1",
    device: {
      uid: "device-1",
      name: "my-device",
      online: true,
      info: { id: "ubuntu" },
    } as Device,
    tenant_id: "tenant-1",
    username: "root",
    ip_address: "192.168.1.1",
    started_at: "2024-01-01T00:00:00Z",
    last_seen: "2024-01-01T00:01:00Z",
    active: false,
    authenticated: true,
    recorded: false,
    type: "term",
    term: "xterm",
    position: { latitude: 0, longitude: 0 },
    events: { types: ["term"], seats: [] },
    web: true,
    ...overrides,
  };
}

function mockSessionsResponse(sessions: Session[] = [], totalCount?: number) {
  return paginatedResponse(sessions, totalCount);
}

function renderTable(isAdmin = false) {
  return render(<RecentSessionsTable isAdmin={isAdmin} />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

describe("RecentSessionsTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sdk.getSessions.mockResolvedValue(mockSessionsResponse());
    sdk.getSessionsAdmin.mockResolvedValue(mockSessionsResponse());
  });

  describe("default (non-admin)", () => {
    it("renders 'View all' link to /sessions", async () => {
      renderTable();
      await waitFor(() => {
        expect(screen.getByRole("link", { name: /view all/i })).toHaveAttribute(
          "href",
          "/sessions",
        );
      });
    });

    it("navigates to /sessions/:uid on row click", async () => {
      const user = userEvent.setup();
      sdk.getSessions.mockResolvedValue(
        mockSessionsResponse([makeSession({ uid: "s-1" })]),
      );
      renderTable();

      await waitFor(() => {
        expect(screen.getByText("root")).toBeInTheDocument();
      });
      await user.click(screen.getByText("root"));
      expect(mockNavigate).toHaveBeenCalledWith("/sessions/s-1");
    });

    it("renders device chip with link to /devices/:uid", async () => {
      sdk.getSessions.mockResolvedValue(mockSessionsResponse([makeSession()]));
      renderTable();

      await waitFor(() => {
        expect(screen.getByText("my-device")).toBeInTheDocument();
      });
      const chip = screen.getByText("my-device").closest("a");
      expect(chip).toHaveAttribute("href", "/devices/device-1");
    });
  });

  describe("admin mode", () => {
    it("renders 'View all' link to /admin/sessions", async () => {
      renderTable(true);
      await waitFor(() => {
        expect(screen.getByRole("link", { name: /view all/i })).toHaveAttribute(
          "href",
          "/admin/sessions",
        );
      });
    });

    it("navigates to /admin/sessions/:uid on row click", async () => {
      const user = userEvent.setup();
      sdk.getSessionsAdmin.mockResolvedValue(
        mockSessionsResponse([makeSession({ uid: "s-2" })]),
      );
      renderTable(true);

      await waitFor(() => {
        expect(screen.getByText("root")).toBeInTheDocument();
      });
      await user.click(screen.getByText("root"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/sessions/s-2");
    });

    it("renders device chip with link to /admin/devices/:uid", async () => {
      sdk.getSessionsAdmin.mockResolvedValue(
        mockSessionsResponse([makeSession()]),
      );
      renderTable(true);

      await waitFor(() => {
        expect(screen.getByText("my-device")).toBeInTheDocument();
      });
      const chip = screen.getByText("my-device").closest("a");
      expect(chip).toHaveAttribute("href", "/admin/devices/device-1");
    });
  });

  describe("loading state", () => {
    it("shows loading message while fetching", () => {
      sdk.getSessions.mockReturnValue(new Promise(() => {}));
      renderTable();
      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error callout on fetch failure", async () => {
      sdk.getSessions.mockRejectedValue(
        Object.assign(new Error(), { status: 500, headers: new Headers() }),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
    });
  });

  describe("empty state", () => {
    it("shows empty message when no sessions", async () => {
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("No recent sessions")).toBeInTheDocument();
      });
    });
  });

  describe("unauthenticated session", () => {
    it("renders warning icon for unauthenticated sessions", async () => {
      sdk.getSessions.mockResolvedValue(
        mockSessionsResponse([makeSession({ authenticated: false })]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByTitle("Not authenticated")).toBeInTheDocument();
      });
    });
  });

  describe("session data rendering", () => {
    it("renders session username", async () => {
      sdk.getSessions.mockResolvedValue(
        mockSessionsResponse([makeSession({ username: "admin" })]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("admin")).toBeInTheDocument();
      });
    });

    it("renders session type badge", async () => {
      sdk.getSessions.mockResolvedValue(
        mockSessionsResponse([
          makeSession({ events: { types: ["shell"], seats: [] } }),
        ]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("shell")).toBeInTheDocument();
      });
    });

    it("renders device name", async () => {
      sdk.getSessions.mockResolvedValue(mockSessionsResponse([makeSession()]));
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("my-device")).toBeInTheDocument();
      });
    });
  });
});
