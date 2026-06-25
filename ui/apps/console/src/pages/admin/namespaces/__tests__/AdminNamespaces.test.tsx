import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { Namespace } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useAdminNamespaces", () => ({
  useAdminNamespaces: vi.fn(),
}));

// Drawer/Dialog mocks — keep tests fast and focused
vi.mock("../EditNamespaceDrawer", () => ({
  default: ({
    open,
  }: {
    open: boolean;
    namespace: unknown;
    onClose: () => void;
  }) => (open ? <div data-testid="edit-drawer" /> : null),
}));

vi.mock("../DeleteNamespaceDialog", () => ({
  default: ({
    open,
  }: {
    open: boolean;
    namespace: unknown;
    onClose: () => void;
  }) => (open ? <div data-testid="delete-dialog" /> : null),
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports ───────────────────────────────────────────────────────────────────

import { useAdminNamespaces } from "@/hooks/useAdminNamespaces";
import AdminNamespaces from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  namespaces: [] as Namespace[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeNamespace(overrides: Partial<Namespace> = {}): Namespace {
  return {
    tenant_id: "tenant-1",
    name: "my-namespace",
    owner: "owner-1",
    members: [{ id: "owner-1", email: "owner@example.com", role: 0 }],
    settings: {
      session_record: true,
      connection_announcement: "",
      device_auto_accept: false,
    },
    max_devices: 10,
    devices_accepted_count: 2,
    created_at: "2024-01-01T00:00:00Z",
    billing: null,
    ...overrides,
  } as Namespace;
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminNamespaces />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("AdminNamespaces", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAdminNamespaces).mockReturnValue(defaultHookState);
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Namespaces" }),
      ).toBeInTheDocument();
    });

    it("renders the search input with correct aria-label", () => {
      renderPage();
      expect(
        screen.getByRole("searchbox", { name: "Search namespaces by name" }),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading namespaces..." text', () => {
      vi.mocked(useAdminNamespaces).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
        namespaces: [],
      });
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading namespaces...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No namespaces found" when the list is empty', () => {
      renderPage();
      expect(screen.getByText("No namespaces found")).toBeInTheDocument();
    });
  });

  describe("namespace rows", () => {
    it("renders a row for each returned namespace", () => {
      vi.mocked(useAdminNamespaces).mockReturnValue({
        ...defaultHookState,
        namespaces: [
          makeNamespace({ tenant_id: "t-1", name: "namespace-alpha" }),
          makeNamespace({ tenant_id: "t-2", name: "namespace-beta" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("namespace-alpha")).toBeInTheDocument();
      expect(screen.getByText("namespace-beta")).toBeInTheDocument();
    });

    it("navigates to namespace detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useAdminNamespaces).mockReturnValue({
        ...defaultHookState,
        namespaces: [
          makeNamespace({ tenant_id: "tenant-xyz", name: "clickable-ns" }),
        ],
        totalCount: 1,
      });
      renderPage();
      await user.click(screen.getByText("clickable-ns"));
      expect(mockNavigate).toHaveBeenCalledWith(
        "/admin/namespaces/tenant-xyz",
      );
    });
  });

  describe("error state", () => {
    it("renders an error alert when the hook returns an error", () => {
      vi.mocked(useAdminNamespaces).mockReturnValue({
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
    it("calls useAdminNamespaces with search and page hydrated from URL params", () => {
      renderPage(["/?search=myns&page=3"]);
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 3 }),
      );
      // The search field should reflect the URL value
      expect(
        screen.getByRole("searchbox", { name: "Search namespaces by name" }),
      ).toHaveValue("myns");
    });

    it("calls useAdminNamespaces with page=1 and search='' when URL has no params", () => {
      renderPage(["/"]);
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });
  });

  describe("URL writes — clearing search resets page to 1 and omits both params", () => {
    it("omits search and page from the URL after clearing a prefilled search", async () => {
      const user = userEvent.setup();
      // Start with ?search=myns&page=3 in the URL
      renderPage(["/?search=myns&page=3"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search namespaces by name",
      });
      expect(searchbox).toHaveValue("myns");

      // Clear the search field
      await user.clear(searchbox);

      // After clearing, useAdminNamespaces must be called with page=1 (default)
      // and search='' (default), meaning neither is in the URL any more.
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
    });
  });

  // ── Exact URL params from unit spec ──────────────────────────────────────────

  describe("URL hydration — ?page=2&search=dev hydrates controls", () => {
    it("hydrates the search field to 'dev' and calls useAdminNamespaces with page=2", () => {
      renderPage(["/?page=2&search=dev"]);

      // SearchField must reflect the URL value immediately
      expect(
        screen.getByRole("searchbox", { name: "Search namespaces by name" }),
      ).toHaveValue("dev");

      // Hook must be called with the URL values
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );
    });
  });

  // ── URL writes — typing in SearchField resets page to 1 ─────────────────────

  describe("URL writes — typing in SearchField resets page to 1", () => {
    it("resets page to 1 and reflects new search value after typing", async () => {
      const user = userEvent.setup();
      // Start on page 2
      renderPage(["/?page=2"]);

      // Confirm initial page
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );

      const searchbox = screen.getByRole("searchbox", {
        name: "Search namespaces by name",
      });

      // Type a search term — this should reset page to 1
      await user.type(searchbox, "dev");

      // Hook must now be called with page=1 (reset) and the searchbox shows new value
      expect(vi.mocked(useAdminNamespaces)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1 }),
      );
      expect(searchbox).toHaveValue("dev");
    });
  });

  // ── Debounce — search is debounced before reaching the query hook ─────────────

  describe("debounce — search is debounced before reaching the query hook", () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("does not pass the new search to useAdminNamespaces until the debounce delay elapses", () => {
      renderPage(["/"]);

      const searchbox = screen.getByRole("searchbox", {
        name: "Search namespaces by name",
      });

      // Fire a change event — before the 300 ms debounce fires the hook should
      // NOT have been called with search: "dev".
      act(() => {
        fireEvent.change(searchbox, { target: { value: "dev" } });
      });

      // Immediately after the change, the debounced value hasn't settled yet.
      const callsBeforeDebounce = vi.mocked(useAdminNamespaces).mock.calls;
      const hadDevBefore = callsBeforeDebounce.some(
        ([args]) => (args as { search?: string }).search === "dev",
      );
      expect(hadDevBefore).toBe(false);

      // Advance timers past the debounce window (300 ms)
      act(() => {
        vi.advanceTimersByTime(350);
      });

      // Now the debounced search should have fired and the hook called with "dev"
      const callsAfterDebounce = vi.mocked(useAdminNamespaces).mock.calls;
      const hadDevAfter = callsAfterDebounce.some(
        ([args]) => (args as { search?: string }).search === "dev",
      );
      expect(hadDevAfter).toBe(true);
    });
  });
});
