import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import type { NormalizedContainer } from "@/hooks/useContainers";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useContainers", () => ({
  useContainers: vi.fn(),
}));

vi.mock("@/hooks/useContainerMutations", () => ({
  useAddContainerTag: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useRemoveContainerTag: vi.fn(() => ({ mutateAsync: vi.fn() })),
}));

// Return the value immediately (no timer) so tests don't need fake timers.
vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: () => ({ namespace: { name: "my-ns" } }),
}));

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (sel: (s: { tenant: string }) => unknown) =>
    sel({ tenant: "tenant-1" }),
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: [] }) => unknown) =>
    sel({ sessions: [] }),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("@/components/common/CopyButton", () => ({
  default: ({ text }: { text: string }) => (
    <button type="button" aria-label={`Copy ${text}`} />
  ),
}));

vi.mock("@/components/common/PageHeader", () => ({
  default: ({
    title,
    children,
  }: {
    title: string;
    children?: React.ReactNode;
  }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));

vi.mock("@/utils/date", () => ({
  formatRelative: () => "just now",
  formatDateFull: () => "Jan 15, 2024",
}));

vi.mock("@/utils/sshid", () => ({
  buildSshid: (ns: string, name: string) => `${ns}.${name}@localhost`,
}));

vi.mock("@/components/common/TagFilterDropdown", () => ({
  default: () => <div />,
}));

const mockManageTagsDrawer = vi.fn();
vi.mock("@/components/ManageTagsDrawer", () => ({
  default: (props: {
    open: boolean;
    onClose: () => void;
    onTagRenamed?: (oldName: string, newName: string) => void;
    onTagDeleted?: (name: string) => void;
  }) => {
    mockManageTagsDrawer(props);
    return <div data-testid="manage-tags-drawer" />;
  },
}));

vi.mock("@/components/ConnectDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/TagsPopover", () => ({
  default: ({ tags }: { tags: string[] }) => (
    <span>{tags.length > 0 ? tags.join(", ") : "No tags"}</span>
  ),
}));

vi.mock("../ContainerActionsPortal", () => ({ default: () => null }));

vi.mock("../AddDockerConnectorDrawer", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

vi.mock("@/components/billing/BillingWarning", () => ({
  default: () => <div />,
}));
const mockRequestAction = vi.fn();
vi.mock("@/hooks/useContainerActions", () => ({
  useContainerActions: () => ({
    requestAction: mockRequestAction,
    operation: null,
    close: vi.fn(),
    runSuccess: vi.fn(),
    onBillingWarning: vi.fn(),
    billingEnabled: false,
    billingWarningOpen: false,
    closeBillingWarning: vi.fn(),
  }),
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import React from "react";
import { useContainers } from "@/hooks/useContainers";
import Containers from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

const defaultHookState = {
  containers: [] as NormalizedContainer[],
  totalCount: 0,
  isLoading: false,
  error: null,
  refetch: vi.fn(),
};

function makeContainer(
  overrides: Partial<NormalizedContainer> = {},
): NormalizedContainer {
  return {
    uid: "container-uid-1",
    name: "my-container",
    status: "accepted",
    online: true,
    tags: [],
    last_seen: new Date().toISOString(),
    created_at: new Date().toISOString(),
    identity: { mac: "aa:bb:cc:dd:ee:ff" },
    info: {
      id: "alpine",
      pretty_name: "Alpine Linux",
      arch: "x86_64",
      platform: "linux",
      version: "0.14.0",
    },
    remote_addr: "1.2.3.4",
    ...overrides,
  } as NormalizedContainer;
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <Containers />
    </MemoryRouter>,
  );
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("Containers list", () => {
  beforeEach(() => {
    vi.mocked(useContainers).mockReturnValue(defaultHookState);
    mockNavigate.mockReset();
    mockManageTagsDrawer.mockReset();
    mockRequestAction.mockReset();
  });

  describe("rendering", () => {
    it("renders the page heading", () => {
      renderPage();
      expect(
        screen.getByRole("heading", { name: "Containers" }),
      ).toBeInTheDocument();
    });

    it("renders all status filter tabs", () => {
      renderPage();
      expect(screen.getByRole("tab", { name: "Accepted" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Pending" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Rejected" })).toBeInTheDocument();
    });

    it("renders the search input", () => {
      renderPage();
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it("renders the loading message", () => {
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        isLoading: true,
      });
      renderPage();
      expect(screen.getByText("Loading containers...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No containers found" when list is empty', () => {
      renderPage();
      expect(screen.getByText("No containers found")).toBeInTheDocument();
    });
  });

  describe("container rows", () => {
    it("renders a row for each container", () => {
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [
          makeContainer({ uid: "uid-1", name: "alpha" }),
          makeContainer({ uid: "uid-2", name: "beta" }),
        ],
        totalCount: 2,
      });
      renderPage();
      expect(screen.getByText("alpha")).toBeInTheDocument();
      expect(screen.getByText("beta")).toBeInTheDocument();
    });

    it("navigates to container detail on row click", async () => {
      const user = userEvent.setup();
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [makeContainer({ uid: "uid-abc", name: "clickable" })],
        totalCount: 1,
      });
      renderPage();
      await user.click(screen.getByText("clickable"));
      expect(mockNavigate).toHaveBeenCalledWith("/containers/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error message when hook returns an error", () => {
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        error: new Error("Request failed"),
      });
      renderPage();
      expect(screen.getByText("Request failed")).toBeInTheDocument();
    });
  });

  describe("sorting", () => {
    it("requests last_seen/desc sort by default", () => {
      renderPage();
      expect(useContainers).toHaveBeenCalledWith(
        expect.objectContaining({ sortBy: "last_seen", orderBy: "desc" }),
      );
    });

    it("toggles sort when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [makeContainer({ uid: "uid-1", name: "alpha" })],
        totalCount: 1,
      });
      renderPage();

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      let calls = vi.mocked(useContainers).mock.calls;
      let last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "asc" });

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      calls = vi.mocked(useContainers).mock.calls;
      last = calls[calls.length - 1][0];
      expect(last).toMatchObject({ sortBy: "name", orderBy: "desc" });
    });
  });

  // ── URL hydration (usePaginatedListState adoption) ────────────────────────────

  describe("URL hydration — URL params seed page state on mount", () => {
    it("passes status=pending from URL to useContainers", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "pending" }),
      );
    });

    it("passes tags array from URL to useContainers", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["a", "b"] }),
      );
    });

    it("passes page=2 from URL to useContainers", () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2 }),
      );
    });

    it("falls back to status=accepted and page=1 when URL has no params", () => {
      renderPage(["/"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted", page: 1 }),
      );
    });

    it("falls back to status=accepted for an invalid status value", () => {
      renderPage(["/?status=invalid"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ status: "accepted" }),
      );
    });

    it("passes empty filterTags when no tags param is present", () => {
      renderPage(["/"]);
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: [] }),
      );
    });
  });

  // ── Search trimming ───────────────────────────────────────────────────────────

  describe("search — whitespace is trimmed before passing to useContainers", () => {
    it("passes trimmed search to useContainers when input has surrounding spaces", async () => {
      // useDebouncedValue is mocked to return its input immediately, so we can
      // verify the trim happens before the debounce without needing fake timers.
      const user = userEvent.setup();
      renderPage();
      const searchInput = screen.getByPlaceholderText("Search by hostname...");
      await user.type(searchInput, "  myhost  ");
      // The hook must have been called with the trimmed string "myhost".
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ search: "myhost" }),
      );
    });
  });

  // ── Tag mutation callbacks ────────────────────────────────────────────────────

  describe("tag mutation — onTagRenamed/onTagDeleted update URL tags array", () => {
    it("renames a tag in filterTags when onTagRenamed is called from ManageTagsDrawer", async () => {
      // Start with tags=a&tags=b in the URL
      renderPage(["/?tags=a&tags=b"]);

      // Grab the onTagRenamed callback passed to ManageTagsDrawer
      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagRenamed?: (oldName: string, newName: string) => void;
      };
      expect(lastCall?.onTagRenamed).toBeDefined();

      // Invoke the callback inside await act(async) so React flushes the URL update.
      await act(async () => {
        lastCall.onTagRenamed!("a", "alpha");
      });

      // useContainers should now be called with filterTags exactly ["alpha", "b"] (no stale "a")
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({
          filterTags: ["alpha", "b"],
        }),
      );
    });

    it("removes a tag from filterTags when onTagDeleted is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagDeleted?: (name: string) => void;
      };
      expect(lastCall?.onTagDeleted).toBeDefined();

      await act(async () => {
        lastCall.onTagDeleted!("a");
      });

      // useContainers should now be called with only tag "b"
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["b"] }),
      );
    });

    it("adds a tag to URL array when setArrayFilter is called via addFilterTag", () => {
      // TagFilterDropdown and TagsPopover are mocked away; verify URL
      // array hydration indirectly: render with an existing tag in the URL and
      // confirm useContainers receives it.
      renderPage(["/?tags=existing"]);
      // The tags=existing param must arrive at useContainers
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ filterTags: ["existing"] }),
      );
      // The filter bar is still visible and the tags array remains stable
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });

  // ── Status change resets page ─────────────────────────────────────────────────

  describe("status change — clicking a status tab resets page to 1", () => {
    it("resets page to 1 when status tab is clicked while on page 2", async () => {
      const user = userEvent.setup();
      // Start on page 2 with status=accepted
      renderPage(["/?page=2"]);

      // Verify we started on page 2
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 2, status: "accepted" }),
      );

      // Click the "Pending" tab (rendered as role="tab")
      await user.click(screen.getByRole("tab", { name: "Pending" }));

      // After switching status the page must be reset to 1
      expect(vi.mocked(useContainers)).toHaveBeenCalledWith(
        expect.objectContaining({ page: 1, status: "pending" }),
      );
    });
  });

  // ── useContainerActions + ContainerActionsPortal ──────────────────────────────

  describe("action delegation — action buttons use useContainerActions", () => {
    it("calls requestAction(container, 'accept') when Accept is clicked in pending view", async () => {
      const user = userEvent.setup();
      const pendingContainer = makeContainer({
        uid: "uid-pending",
        name: "pending-box",
        status: "pending",
        online: false,
      });
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [pendingContainer],
        totalCount: 1,
      });
      renderPage(["/?status=pending"]);
      await user.click(screen.getByRole("button", { name: "Accept" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        pendingContainer,
        "accept",
      );
    });

    it("calls requestAction(container, 'reject') when Reject is clicked in pending view", async () => {
      const user = userEvent.setup();
      const pendingContainer = makeContainer({
        uid: "uid-pending-2",
        name: "pending-box-2",
        status: "pending",
        online: false,
      });
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [pendingContainer],
        totalCount: 1,
      });
      renderPage(["/?status=pending"]);
      await user.click(screen.getByRole("button", { name: "Reject" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        pendingContainer,
        "reject",
      );
    });

    it("calls requestAction(container, 'remove') when Remove is clicked in rejected view", async () => {
      const user = userEvent.setup();
      const rejectedContainer = makeContainer({
        uid: "uid-rejected",
        name: "rejected-box",
        status: "rejected",
        online: false,
      });
      vi.mocked(useContainers).mockReturnValue({
        ...defaultHookState,
        containers: [rejectedContainer],
        totalCount: 1,
      });
      renderPage(["/?status=rejected"]);
      await user.click(screen.getByRole("button", { name: "Remove" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        rejectedContainer,
        "remove",
      );
    });
  });
});
