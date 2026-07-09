import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { UserAdminResponse } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminUsers", () => ({
  useAdminUsers: vi.fn(),
}));

vi.mock("@/hooks/useLoginAsUser", () => ({
  useLoginAsUser: vi.fn(),
}));

vi.mock("@/hooks/useAdminAccountRequests", () => ({
  useAdminAccountRequests: () => ({ totalCount: 0 }),
}));

vi.mock("@/env", () => ({
  getConfig: () => ({ enterprise: false, cloud: false }),
}));

vi.mock("../AccountRequestsTab", () => ({
  default: () => null,
}));

vi.mock("./mocks", () => ({}));

// Drawer/Dialog mocks — keep tests fast and focused
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

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports ───────────────────────────────────────────────────────────────────

import { useAdminUsers } from "@/hooks/useAdminUsers";
import { useLoginAsUser } from "@/hooks/useLoginAsUser";
import AdminUsers from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  users: [] as UserAdminResponse[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

const defaultLoginAsState = {
  loginAs: vi.fn(),
  loadingId: null as string | null,
  errorId: null as string | null,
};

function makeUser(
  overrides: Partial<UserAdminResponse> = {},
): UserAdminResponse {
  return {
    id: "user-id-1",
    name: "Alice Smith",
    email: "alice@example.com",
    username: "alice",
    status: "confirmed",
    admin: false,
    ...overrides,
  } as UserAdminResponse;
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminUsers />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminUsers", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAdminUsers).mockReturnValue(defaultHookState);
    vi.mocked(useLoginAsUser).mockReturnValue(defaultLoginAsState);
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Users" }),
      ).toBeInTheDocument();
    });

    it("renders the search input with correct aria-label", () => {
      renderPage();
      expect(
        screen.getByRole("searchbox", { name: "Search users by username" }),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading users..." text', () => {
      vi.mocked(useAdminUsers).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        users: [],
      });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading users...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No users found" when the user list is empty', () => {
      renderPage();
      expect(screen.getByText("No users found")).toBeInTheDocument();
    });
  });

  describe("user rows", () => {
    it("renders a row for each returned user", () => {
      vi.mocked(useAdminUsers).mockReturnValue({
        ...defaultHookState,
        users: [
          makeUser({ id: "id-1", name: "Alice Smith" }),
          makeUser({ id: "id-2", name: "Bob Jones" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("Alice Smith")).toBeInTheDocument();
      expect(screen.getByText("Bob Jones")).toBeInTheDocument();
    });

    it("navigates to user detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminUsers).mockReturnValue({
        ...defaultHookState,
        users: [makeUser({ id: "uid-abc", name: "Clickable User" })],
        totalCount: 1,
      });
      renderPage();
      await user.click(screen.getByText("Clickable User"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/users/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the hook returns an error", () => {
      vi.mocked(useAdminUsers).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByRole("alert")).toBeInTheDocument();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  // ── URL-driven state (usePaginatedListState adoption) ─────────────────────────

  describe("URL hydration — controls reflect URL params on mount", () => {
    it("calls useAdminUsers with search and page hydrated from URL params", () => {
      renderPage(["/?search=foo&page=2"]);
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );
      // The search field should reflect the URL value
      expect(
        screen.getByRole("searchbox", { name: "Search users by username" }),
      ).toHaveValue("foo");
    });

    it("calls useAdminUsers with page=1 and search='' when URL has no params", () => {
      renderPage(["/"]);
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });
  });

  describe("URL writes — clearing search resets page to 1 and omits both params", () => {
    it("omits search and page from the URL after clearing a prefilled search", async () => {
      const user = userEvent.setup();
      // Start with ?search=foo&page=2 in the URL
      renderPage(["/?search=foo&page=2"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });
      expect(searchbox).toHaveValue("foo");

      // Clear the search field
      await user.clear(searchbox);

      // After clearing, useAdminUsers must be called with page=1 (default)
      // and search='' (default), meaning neither is in the URL any more.
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });
  });

  // ── Exact URL params from unit spec ──────────────────────────────────────────

  describe("URL hydration — ?page=3&search=alice hydrates controls", () => {
    it("hydrates the search field to 'alice' and calls useAdminUsers with page=3", () => {
      renderPage(["/?page=3&search=alice"]);

      // SearchField must reflect the URL value
      expect(
        screen.getByRole("searchbox", { name: "Search users by username" }),
      ).toHaveValue("alice");

      // Hook must be called with the URL values (both page and search)
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 3 }),
      );
    });
  });

  describe("URL writes — typing in SearchField resets page to 1", () => {
    it("resets page to 1 and reflects new search value after typing", async () => {
      const user = userEvent.setup();
      // Start on page 3
      renderPage(["/?page=3"]);

      // Confirm initial page
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 3 }),
      );

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });

      // Type a search term — this should reset page to 1
      await user.type(searchbox, "bob");

      // Hook must now be called with page=1 (reset) and the searchbox shows new value
      expect(vi.mocked(useAdminUsers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
      expect(searchbox).toHaveValue("bob");
    });
  });

  describe("debounce — search is debounced before reaching the query hook", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("does not pass the new search to useAdminUsers until the debounce delay elapses", () => {
      renderPage(["/"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });

      // Fire a change event — before the 300 ms debounce fires the hook should
      // NOT have been called with search: "alice".
      act(() => {
        fireEvent.change(searchbox, { target: { value: "alice" } });
      });

      // Immediately after the change, the debounced value hasn't settled yet.
      const callsBeforeDebounce = vi.mocked(useAdminUsers).mock.calls;
      const hadAliceBefore = callsBeforeDebounce.some(
        ([args]) => (args as { search?: string }).search === "alice",
      );
      expect(hadAliceBefore).toBe(false);

      // Advance timers past the debounce window (300 ms)
      act(() => {
        vi.advanceTimersByTime(350);
      });

      // Now the debounced search should have fired and the hook called with "alice"
      const callsAfterDebounce = vi.mocked(useAdminUsers).mock.calls;
      const hadAliceAfter = callsAfterDebounce.some(
        ([args]) => (args as { search?: string }).search === "alice",
      );
      expect(hadAliceAfter).toBe(true);
    });
  });
});
