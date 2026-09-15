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
import AdminNamespaces from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

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

const mockNavigate = vi.hoisted(() => vi.fn());

vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

let lastRequestUrl: URL | null;

function setNamespaces(
  namespaces: ReturnType<typeof mockNamespace>[],
  total?: number,
) {
  server.use(
    http.get("*/admin/api/namespaces", ({ request }) => {
      lastRequestUrl = new URL(request.url);
      return jsonWithTotal(namespaces, total ?? namespaces.length);
    }),
  );
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <AdminNamespaces />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

describe("AdminNamespaces", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    lastRequestUrl = null;
    useAuthStore.setState({ isAdmin: true });
    setNamespaces([]);
  });

  describe("loading state", () => {
    it('renders the loading spinner with "Loading namespaces..." text', () => {
      server.use(
        http.get("*/admin/api/namespaces", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByRole("status")).toBeInTheDocument();
      expect(screen.getByText("Loading namespaces...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No namespaces found" when the list is empty', async () => {
      renderPage();
      expect(
        await screen.findByText("No namespaces found"),
      ).toBeInTheDocument();
    });
  });

  describe("namespace rows", () => {
    it("renders a row for each returned namespace", async () => {
      setNamespaces([
        mockNamespace({ tenant_id: "t-1", name: "namespace-alpha" }),
        mockNamespace({ tenant_id: "t-2", name: "namespace-beta" }),
      ]);
      renderPage();
      expect(await screen.findByText("namespace-alpha")).toBeInTheDocument();
      expect(screen.getByText("namespace-beta")).toBeInTheDocument();
    });

    it("navigates to namespace detail page when a row is clicked", async () => {
      const user = userEvent.setup();
      setNamespaces([
        mockNamespace({ tenant_id: "tenant-xyz", name: "clickable-ns" }),
      ]);
      renderPage();
      await user.click(await screen.findByText("clickable-ns"));
      expect(mockNavigate).toHaveBeenCalledWith("/admin/namespaces/tenant-xyz");
    });
  });

  describe("error state", () => {
    it("renders an error alert when the SDK returns an error", async () => {
      server.use(
        http.get("*/admin/api/namespaces", () =>
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
        name: "Search namespaces by name",
      });

      act(() => {
        fireEvent.change(searchbox, { target: { value: "dev" } });
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
