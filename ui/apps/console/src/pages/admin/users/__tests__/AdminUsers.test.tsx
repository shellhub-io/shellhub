import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  render,
  screen,
  fireEvent,
  act,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import AdminUsers from "../index";
import type { UserAdminResponse } from "@/client/model";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";

vi.mock("../AccountRequestsTab", () => ({
  default: () => null,
}));

vi.mock("../CreateUserDrawer", () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="create-drawer" /> : null,
}));

vi.mock("../EditUserDrawer", () => ({
  default: ({ open }: { open: boolean; user: unknown; onClose: () => void }) =>
    open ? <div data-testid="edit-drawer" /> : null,
}));

vi.mock("../DeleteUserDialog", () => ({
  default: ({ open }: { open: boolean; user: unknown; onClose: () => void }) =>
    open ? <div data-testid="delete-dialog" /> : null,
}));

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function mockAdminUser(
  overrides: Partial<UserAdminResponse> = {},
): UserAdminResponse {
  return {
    id: "user-id-1",
    name: "Alice Smith",
    email: "alice@example.com",
    username: "alice",
    status: "confirmed",
    admin: false,
    created_at: "2024-01-01T00:00:00Z",
    last_login: "2024-06-01T00:00:00Z",
    ...overrides,
  };
}

let lastRequestUrl: URL | null;

function setUsers(users: UserAdminResponse[], total?: number) {
  server.use(
    http.get("*/admin/api/users", ({ request }) => {
      lastRequestUrl = new URL(request.url);
      return jsonWithTotal(users, total ?? users.length);
    }),
  );
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminUsers />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminUsers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    lastRequestUrl = null;
    useAuthStore.setState({ isAdmin: true });
    setUsers([]);
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading users..." text', () => {
      server.use(
        http.get("*/admin/api/users", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading users...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No users found" when the user list is empty', async () => {
      renderPage();
      expect(await screen.findByText("No users found")).toBeInTheDocument();
    });
  });

  describe("user rows", () => {
    it("renders a row for each returned user", async () => {
      setUsers([
        mockAdminUser({ id: "id-1", name: "Alice Smith" }),
        mockAdminUser({ id: "id-2", name: "Bob Jones" }),
      ]);
      renderPage();
      expect(await screen.findByText("Alice Smith")).toBeInTheDocument();
      expect(screen.getByText("Bob Jones")).toBeInTheDocument();
    });

    it("navigates to user detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      setUsers([
        mockAdminUser({ id: "uid-abc", name: "Clickable User" }),
      ]);
      renderPage();
      await user.click(await screen.findByText("Clickable User"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/users/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      server.use(
        http.get("*/admin/api/users", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("debounce — search is debounced before reaching the query hook", () => {
    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("does not pass the new search to the API until the debounce delay elapses", async () => {
      renderPage(["/"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });

      act(() => {
        fireEvent.change(searchbox, { target: { value: "alice" } });
      });

      const hasFilter = () =>
        lastRequestUrl !== null &&
        lastRequestUrl.searchParams.get("filter") !== null;
      expect(hasFilter()).toBe(false);

      act(() => {
        vi.advanceTimersByTime(350);
      });

      await waitFor(() => {
        expect(hasFilter()).toBe(true);
      });
    });
  });
});
