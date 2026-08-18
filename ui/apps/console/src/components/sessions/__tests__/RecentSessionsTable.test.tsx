import { type ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import RecentSessionsTable from "../RecentSessionsTable";
import type { Device, Session } from "@/client";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

vi.mock("@/client", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client")>();
  return {
    ...actual,
    getSessions: vi.fn(),
    getSessionsAdmin: vi.fn(),
  };
});

vi.mock("@/stores/authStore", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/stores/authStore")>();
  return {
    ...actual,
    useAuthStore: vi.fn((selector: (s: { isAdmin: boolean }) => unknown) =>
      selector({ isAdmin: true }),
    ),
  };
});

import { getSessions, getSessionsAdmin } from "@/client";

function makeSession(overrides: Partial<Session> = {}): Session {
  return {
    uid: "session-1",
    device_uid: "device-1",
    device: { uid: "device-1", name: "my-device", online: true, info: { id: "ubuntu" } } as Device,
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

function mockSdkResponse(sessions: Session[] = [], totalCount?: number) {
  return {
    data: sessions,
    response: {
      headers: new Headers({ "X-Total-Count": String(totalCount ?? sessions.length) }),
    },
  } as never;
}

function renderTable(isAdmin = false) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>{children}</MemoryRouter>
      </QueryClientProvider>
    );
  }

  return render(<RecentSessionsTable isAdmin={isAdmin} />, { wrapper: Wrapper });
}

describe("RecentSessionsTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(getSessions).mockResolvedValue(mockSdkResponse());
    vi.mocked(getSessionsAdmin).mockResolvedValue(mockSdkResponse());
  });

  describe("default (non-admin)", () => {
    it("renders 'View all' link to /sessions", async () => {
      renderTable();
      await waitFor(() => {
        expect(screen.getByRole("link", { name: /view all/i })).toHaveAttribute("href", "/sessions");
      });
    });

    it("navigates to /sessions/:uid on row click", async () => {
      const user = userEvent.setup();
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession({ uid: "s-1" })]),
      );
      renderTable();

      await waitFor(() => {
        expect(screen.getByText("root")).toBeInTheDocument();
      });
      await user.click(screen.getByText("root"));
      expect(mockNavigate).toHaveBeenCalledWith("/sessions/s-1");
    });

    it("renders device chip with link to /devices/:uid", async () => {
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession()]),
      );
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
        expect(screen.getByRole("link", { name: /view all/i })).toHaveAttribute("href", "/admin/sessions");
      });
    });

    it("navigates to /admin/sessions/:uid on row click", async () => {
      const user = userEvent.setup();
      vi.mocked(getSessionsAdmin).mockResolvedValue(
        mockSdkResponse([makeSession({ uid: "s-2" })]),
      );
      renderTable(true);

      await waitFor(() => {
        expect(screen.getByText("root")).toBeInTheDocument();
      });
      await user.click(screen.getByText("root"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/sessions/s-2");
    });

    it("renders device chip with link to /admin/devices/:uid", async () => {
      vi.mocked(getSessionsAdmin).mockResolvedValue(
        mockSdkResponse([makeSession()]),
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
      vi.mocked(getSessions).mockReturnValue(new Promise(() => {}) as never);
      renderTable();
      expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
    });
  });

  describe("error state", () => {
    it("renders error callout on fetch failure", async () => {
      vi.mocked(getSessions).mockRejectedValue(
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
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession({ authenticated: false })]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByTitle("Not authenticated")).toBeInTheDocument();
      });
    });
  });

  describe("session data rendering", () => {
    it("renders session username", async () => {
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession({ username: "admin" })]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("admin")).toBeInTheDocument();
      });
    });

    it("renders session type badge", async () => {
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession({ events: { types: ["shell"], seats: [] } })]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("shell")).toBeInTheDocument();
      });
    });

    it("renders device name", async () => {
      vi.mocked(getSessions).mockResolvedValue(
        mockSdkResponse([makeSession()]),
      );
      renderTable();
      await waitFor(() => {
        expect(screen.getByText("my-device")).toBeInTheDocument();
      });
    });
  });
});
