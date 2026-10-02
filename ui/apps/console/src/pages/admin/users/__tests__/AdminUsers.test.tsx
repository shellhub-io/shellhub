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
import type { UserAdminResponse } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { mockAdminUser } from "@/tests/factories";
import { decodeB64url } from "@/tests/decodeB64url";
import { setEdition } from "@/tests/edition";

vi.mock("../CreateUserModal", () => ({
  default: ({ open }: { open: boolean }) =>
    open ? <div data-testid="create-modal" /> : null,
}));

vi.mock("../EditUserModal", () => ({
  default: ({ open }: { open: boolean; user: unknown; onClose: () => void }) =>
    open ? <div data-testid="edit-modal" /> : null,
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
    setEdition("community");
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

    it("opens the reject dialog, not the delete one, for an account awaiting approval", async () => {
      setEdition("enterprise");
      setUsers([
        mockAdminUser({ email: "maria@acme.io", awaiting_approval: true }),
      ]);
      renderPage(["/?subset=awaiting_approval"]);

      await userEvent.click(
        await screen.findByRole("button", {
          name: "Reject account for maria@acme.io",
        }),
      );

      expect(
        await screen.findByRole("dialog", { name: "Reject user" }),
      ).toBeInTheDocument();
      expect(screen.queryByTestId("delete-dialog")).not.toBeInTheDocument();
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

  describe("subset tabs", () => {
    function sentFilter() {
      const raw = lastRequestUrl?.searchParams.get("filter");
      return raw ? decodeB64url(raw) : null;
    }

    it.each([
      [
        "admin",
        [
          {
            type: "property",
            params: { name: "admin", operator: "bool", value: true },
          },
        ],
      ],
      [
        "not_confirmed",
        [
          {
            type: "property",
            params: { name: "status", operator: "eq", value: "not-confirmed" },
          },
        ],
      ],
    ])("sends the %s filter from the URL", async (subset, expected) => {
      renderPage([`/?subset=${subset}`]);
      await waitFor(() => expect(sentFilter()).toEqual(expected));
    });

    it("searches name, username and email when no subset is picked", async () => {
      renderPage(["/?search=al"]);
      await waitFor(() =>
        expect(sentFilter()).toEqual(
          ["name", "username", "email"].map((name) => ({
            type: "property",
            params: { name, operator: "contains", value: "al" },
          })),
        ),
      );
      expect(
        screen.getByRole("searchbox", {
          name: "Search users by name, username or email",
        }),
      ).toBeInTheDocument();
    });

    it("combines a subset with the username search under an and operator", async () => {
      renderPage(["/?subset=admin&search=al"]);
      await waitFor(() =>
        expect(sentFilter()).toEqual([
          { type: "operator", params: { name: "and" } },
          {
            type: "property",
            params: { name: "username", operator: "contains", value: "al" },
          },
          {
            type: "property",
            params: { name: "admin", operator: "bool", value: true },
          },
        ]),
      );
    });

    it("sends the awaiting approval filter in enterprise", async () => {
      setEdition("enterprise");
      renderPage(["/?subset=awaiting_approval"]);
      await waitFor(() =>
        expect(sentFilter()).toEqual([
          {
            type: "property",
            params: {
              name: "awaiting_approval",
              operator: "bool",
              value: true,
            },
          },
        ]),
      );
    });

    it("narrows the search box to username under a tab", async () => {
      renderPage(["/?subset=admin"]);
      expect(
        await screen.findByRole("searchbox", {
          name: "Search users by username",
        }),
      ).toBeInTheDocument();
    });

    it("switches the subset when a tab is clicked", async () => {
      const user = userEvent.setup();
      renderPage();

      await user.click(await screen.findByRole("tab", { name: "Admins" }));
      await waitFor(() =>
        expect(sentFilter()).toEqual([
          {
            type: "property",
            params: { name: "admin", operator: "bool", value: true },
          },
        ]),
      );
      expect(screen.getByRole("tab", { name: "Admins" })).toHaveAttribute(
        "aria-selected",
        "true",
      );
    });

    it("keeps only the selected tab in the tab order", async () => {
      renderPage(["/?subset=admin"]);

      expect(
        await screen.findByRole("tab", { name: "Admins" }),
      ).toHaveAttribute("tabindex", "0");
      expect(screen.getByRole("tab", { name: "All" })).toHaveAttribute(
        "tabindex",
        "-1",
      );
    });

    it.each([
      {
        case: "ArrowRight moves to the next tab",
        from: "/?subset=admin",
        start: "Admins",
        key: "{ArrowRight}",
        expected: "Unconfirmed",
      },
      {
        case: "ArrowLeft from the first tab wraps to the last",
        from: "/",
        start: "All",
        key: "{ArrowLeft}",
        expected: "Unconfirmed",
      },
      {
        case: "End selects the last tab",
        from: "/",
        start: "All",
        key: "{End}",
        expected: "Unconfirmed",
      },
      {
        case: "Home selects the first tab",
        from: "/?subset=admin",
        start: "Admins",
        key: "{Home}",
        expected: "All",
      },
    ])("$case", async ({ from, start, key, expected }) => {
      const user = userEvent.setup();
      renderPage([from]);

      (await screen.findByRole("tab", { name: start })).focus();
      await user.keyboard(key);

      const target = screen.getByRole("tab", { name: expected });
      await waitFor(() => expect(target).toHaveFocus());
      expect(target).toHaveAttribute("aria-selected", "true");
    });

    it("ignores awaiting approval outside enterprise", async () => {
      setEdition("cloud");
      renderPage(["/?subset=awaiting_approval"]);
      await screen.findByText("No users found");
      expect(sentFilter()).toBeNull();
      expect(
        screen.queryByRole("tab", { name: "Awaiting approval" }),
      ).not.toBeInTheDocument();
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
        name: "Search users by name, username or email",
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
