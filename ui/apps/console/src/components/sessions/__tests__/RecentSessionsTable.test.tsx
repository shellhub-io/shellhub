import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import RecentSessionsTable from "../RecentSessionsTable";
import type { Device, Session } from "@/client/model";

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
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

function mockSessionsHandler(sessions: Session[] = [], total?: number) {
  return () => jsonWithTotal(sessions, total);
}

function renderTable(isAdmin = false) {
  return render(<RecentSessionsTable isAdmin={isAdmin} />, {
    wrapper: createTestWrapper({ initialEntries: ["/"] }),
  });
}

describe("RecentSessionsTable", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ isAdmin: true });
    server.use(
      http.get("*/api/sessions", mockSessionsHandler()),
      http.get("*/admin/api/sessions", mockSessionsHandler()),
    );
  });

  it.each([
    [false, "*/api/sessions", ""],
    [true, "*/admin/api/sessions", "/admin"],
  ] as const)(
    "isAdmin=%s links View all, the row and the device chip under '%s'",
    async (isAdmin, endpoint, prefix) => {
      const user = userEvent.setup();
      server.use(
        http.get(endpoint, () => jsonWithTotal([makeSession({ uid: "s-1" })])),
      );
      renderTable(isAdmin);

      await waitFor(() => {
        expect(screen.getByText("root")).toBeInTheDocument();
      });

      expect(screen.getByRole("link", { name: /view all/i })).toHaveAttribute(
        "href",
        `${prefix}/sessions`,
      );
      expect(screen.getByText("my-device").closest("a")).toHaveAttribute(
        "href",
        `${prefix}/devices/device-1`,
      );

      await user.click(screen.getByText("root"));
      expect(mockNavigate).toHaveBeenCalledWith(`${prefix}/sessions/s-1`);
    },
  );

  it("shows loading message while fetching", () => {
    server.use(http.get("*/api/sessions", () => new Promise(() => {})));
    renderTable();
    expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
  });

  it("renders error callout on fetch failure", async () => {
    server.use(
      http.get("*/api/sessions", () => HttpResponse.json({}, { status: 500 })),
    );
    renderTable();
    await waitFor(() => {
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });
  });

  it("shows empty message when no sessions", async () => {
    renderTable();
    await waitFor(() => {
      expect(screen.getByText("No recent sessions")).toBeInTheDocument();
    });
  });

  it("renders warning icon for unauthenticated sessions", async () => {
    server.use(
      http.get("*/api/sessions", () =>
        jsonWithTotal([makeSession({ authenticated: false })]),
      ),
    );
    renderTable();
    await waitFor(() => {
      expect(screen.getByTitle("Not authenticated")).toBeInTheDocument();
    });
  });

  it("renders session type badge", async () => {
    server.use(
      http.get("*/api/sessions", () =>
        jsonWithTotal([
          makeSession({ events: { types: ["shell"], seats: [] } }),
        ]),
      ),
    );
    renderTable();
    await waitFor(() => {
      expect(screen.getByText("shell")).toBeInTheDocument();
    });
  });
});
