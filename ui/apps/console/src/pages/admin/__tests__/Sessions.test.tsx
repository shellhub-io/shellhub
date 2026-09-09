import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import AdminSessions from "../Sessions";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSession } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

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
    http.get("*/admin/api/sessions", () =>
      jsonWithTotal(sessions, total ?? sessions.length),
    ),
  );
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminSessions />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  useAuthStore.setState({ isAdmin: true });
  setSessions([]);
});

describe("AdminSessions", () => {
  it("shows a loading spinner while fetching", () => {
    server.use(http.get("*/admin/api/sessions", () => new Promise(() => {})));
    renderPage();
    expect(screen.getByText(/loading sessions/i)).toBeInTheDocument();
  });

  it("shows 'No sessions found' when there are no sessions", async () => {
    renderPage();
    expect(await screen.findByText("No sessions found")).toBeInTheDocument();
  });

  it("displays the console's own copy for the status in the error banner", async () => {
    server.use(
      http.get("*/admin/api/sessions", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    renderPage();
    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("You do not have permission to do this.");
  });

  describe("session rows", () => {
    it("renders the device name, the IP and a truncated session uid", async () => {
      setSessions([
        mockSession({ uid: "abcdef1234567890", ip_address: "10.0.0.1" }),
      ]);
      renderPage();
      expect(await screen.findByText("my-device")).toBeInTheDocument();
      expect(screen.getByText("10.0.0.1")).toBeInTheDocument();
      expect(screen.getByText("abcdef1234")).toBeInTheDocument();
    });

    it("shows the truncated device_uid when device object is missing", async () => {
      setSessions([mockSession({ device: null, device_uid: "abcd1234efgh" })]);
      renderPage();
      expect(await screen.findByText("abcd1234")).toBeInTheDocument();
    });

    it("navigates to session detail when a row is clicked", async () => {
      const user = userEvent.setup();
      setSessions([mockSession({ uid: "session-abc" })]);
      renderPage();

      await user.click(await screen.findByText("root"));

      expect(mockNavigate).toHaveBeenCalledWith("/admin/sessions/session-abc");
    });
  });

  it.each([
    [true, "Authenticated"],
    [false, "Not authenticated"],
  ])("authenticated=%s renders the '%s' shield", async (
    authenticated,
    title,
  ) => {
    setSessions([mockSession({ authenticated })]);
    renderPage();
    await screen.findByText("root");
    expect(screen.getAllByTitle(title).length).toBeGreaterThan(0);
  });
});
