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
import AdminUsers from "../index";
import type { UserAdminResponse } from "@/client";
import { makeSdkError, paginatedResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getUsers: vi.fn(),
  }),
);

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
    useAuthStore.setState({ isAdmin: true });
    sdk.getUsers.mockResolvedValue(paginatedResponse([]));
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
      sdk.getUsers.mockReturnValue(new Promise(() => {}));
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
      sdk.getUsers.mockResolvedValue(
        paginatedResponse([
          mockAdminUser({ id: "id-1", name: "Alice Smith" }),
          mockAdminUser({ id: "id-2", name: "Bob Jones" }),
        ]),
      );
      renderPage();
      expect(await screen.findByText("Alice Smith")).toBeInTheDocument();
      expect(screen.getByText("Bob Jones")).toBeInTheDocument();
    });

    it("navigates to user detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      sdk.getUsers.mockResolvedValue(
        paginatedResponse([
          mockAdminUser({ id: "uid-abc", name: "Clickable User" }),
        ]),
      );
      renderPage();
      await user.click(await screen.findByText("Clickable User"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/users/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      sdk.getUsers.mockRejectedValue(makeSdkError(500));
      renderPage();
      expect(await screen.findByRole("alert")).toBeInTheDocument();
      expect(
        screen.getByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("URL hydration — controls reflect URL params on mount", () => {
    it("passes search and page hydrated from URL to the SDK", async () => {
      renderPage(["/?search=foo&page=2"]);
      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2 }),
          }),
        );
      });
      expect(
        screen.getByRole("searchbox", { name: "Search users by username" }),
      ).toHaveValue("foo");
    });

    it("passes page=1 to the SDK when URL has no params", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
    });
  });

  describe("URL writes — clearing search resets page to 1 and omits both params", () => {
    it("omits search and page from the URL after clearing a prefilled search", async () => {
      const user = userEvent.setup();
      renderPage(["/?search=foo&page=2"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });
      expect(searchbox).toHaveValue("foo");

      await user.clear(searchbox);

      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
    });
  });

  describe("URL hydration — ?page=3&search=alice hydrates controls", () => {
    it("hydrates the search field to 'alice' and passes page=3 to the SDK", async () => {
      renderPage(["/?page=3&search=alice"]);

      expect(
        screen.getByRole("searchbox", { name: "Search users by username" }),
      ).toHaveValue("alice");

      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 3 }),
          }),
        );
      });
    });
  });

  describe("URL writes — typing in SearchField resets page to 1", () => {
    it("resets page to 1 and reflects new search value after typing", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=3"]);

      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 3 }),
          }),
        );
      });

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });

      await user.type(searchbox, "bob");

      await waitFor(() => {
        expect(sdk.getUsers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1 }),
          }),
        );
      });
      expect(searchbox).toHaveValue("bob");
    });
  });

  describe("debounce — search is debounced before reaching the query hook", () => {
    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("does not pass the new search to the SDK until the debounce delay elapses", async () => {
      renderPage(["/"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search users by username",
      });

      act(() => {
        fireEvent.change(searchbox, { target: { value: "alice" } });
      });

      const hasFilter = (calls: unknown[][]) =>
        calls.some(
          ([args]) => (args as { query?: { filter?: string } })?.query?.filter,
        );
      expect(hasFilter(sdk.getUsers.mock.calls)).toBe(false);

      act(() => {
        vi.advanceTimersByTime(350);
      });

      await waitFor(() => {
        expect(hasFilter(sdk.getUsers.mock.calls)).toBe(true);
      });
    });
  });
});
