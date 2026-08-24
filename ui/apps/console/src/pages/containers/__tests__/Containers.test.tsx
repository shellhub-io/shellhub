import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockContainer, mockNamespace } from "@/tests/factories";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { seedAuthStore } from "@/tests/seedAuthStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getContainers: vi.fn(),
    createTag: vi.fn(),
    pushTagToContainer: vi.fn(),
    pullTagFromContainer: vi.fn(),
    getNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
  }),
);

vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

vi.mock("@/stores/terminalStore", () => ({
  useTerminalStore: (sel: (s: { sessions: [] }) => unknown) =>
    sel({ sessions: [] }),
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
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

import Containers from "../index";

function renderPage(initialEntries: string[] = ["/"]) {
  return render(<Containers />, {
    wrapper: createTestWrapper({ initialEntries }),
  });
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.getContainers.mockResolvedValue(paginatedResponse([]));
  sdk.getNamespace.mockResolvedValue(mockSdkResponse(mockNamespace()));
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
  sdk.createTag.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pushTagToContainer.mockResolvedValue(mockSdkResponse(undefined));
  sdk.pullTagFromContainer.mockResolvedValue(mockSdkResponse(undefined));
  mockNavigate.mockReset();
  mockManageTagsDrawer.mockReset();
  mockRequestAction.mockReset();
});

describe("Containers list", () => {
  describe("rendering", () => {
    it("renders the page heading", async () => {
      renderPage();
      expect(
        await screen.findByRole("heading", { name: "Containers" }),
      ).toBeInTheDocument();
    });

    it("renders all status filter tabs", async () => {
      renderPage();
      expect(
        await screen.findByRole("tab", { name: "Accepted" }),
      ).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Pending" })).toBeInTheDocument();
      expect(screen.getByRole("tab", { name: "Rejected" })).toBeInTheDocument();
    });

    it("renders the search input", async () => {
      renderPage();
      expect(
        await screen.findByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });

  describe("loading state", () => {
    it("renders the loading message", () => {
      sdk.getContainers.mockReturnValue(new Promise(() => {}));
      renderPage();
      expect(screen.getByText("Loading containers...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No containers found" when list is empty', async () => {
      renderPage();
      expect(
        await screen.findByText("No containers found"),
      ).toBeInTheDocument();
    });
  });

  describe("container rows", () => {
    it("renders a row for each container", async () => {
      sdk.getContainers.mockResolvedValue(
        paginatedResponse(
          [
            mockContainer({ uid: "uid-1", name: "alpha" }),
            mockContainer({ uid: "uid-2", name: "beta" }),
          ],
          2,
        ),
      );
      renderPage();
      expect(await screen.findByText("alpha")).toBeInTheDocument();
      expect(screen.getByText("beta")).toBeInTheDocument();
    });

    it("navigates to container detail on row click", async () => {
      const user = userEvent.setup();
      sdk.getContainers.mockResolvedValue(
        paginatedResponse(
          [mockContainer({ uid: "uid-abc", name: "clickable" })],
          1,
        ),
      );
      renderPage();
      await user.click(await screen.findByText("clickable"));
      expect(mockNavigate).toHaveBeenCalledWith("/containers/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error message when the query fails", async () => {
      sdk.getContainers.mockRejectedValue({ status: 500 });
      renderPage();
      expect(
        await screen.findByText("Something went wrong on our side. Try again."),
      ).toBeInTheDocument();
    });
  });

  describe("sorting", () => {
    it("requests last_seen/desc sort by default", async () => {
      renderPage();
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({
              sort_by: "last_seen",
              order_by: "desc",
            }),
          }),
        );
      });
    });

    it("toggles sort when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      sdk.getContainers.mockResolvedValue(
        paginatedResponse([mockContainer({ uid: "uid-1", name: "alpha" })], 1),
      );
      renderPage();
      await screen.findByText("alpha");

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({
              sort_by: "name",
              order_by: "asc",
            }),
          }),
        );
      });

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({
              sort_by: "name",
              order_by: "desc",
            }),
          }),
        );
      });
    });
  });

  describe("URL hydration — URL params seed page state on mount", () => {
    it("passes status=pending from URL to the SDK", async () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ status: "pending" }),
          }),
        );
      });
    });

    it("passes tags from URL as a filter to the SDK", async () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      await waitFor(() => {
        const call = sdk.getContainers.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).toContain('"a"');
        expect(decoded).toContain('"b"');
      });
    });

    it("passes page=2 from URL to the SDK", async () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2 }),
          }),
        );
      });
    });

    it("falls back to status=accepted and page=1 when URL has no params", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ status: "accepted", page: 1 }),
          }),
        );
      });
    });

    it("falls back to status=accepted for an invalid status value", async () => {
      renderPage(["/?status=invalid"]);
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ status: "accepted" }),
          }),
        );
      });
    });

    it("passes no tag filter when no tags param is present", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        const call = sdk.getContainers.mock.calls[0]?.[0] as {
          query?: { filter?: string };
        };
        expect(call?.query?.filter).toBeUndefined();
      });
    });
  });

  describe("search — whitespace is trimmed before passing to the SDK", () => {
    it("passes trimmed search when input has surrounding spaces", async () => {
      const user = userEvent.setup();
      renderPage();
      await screen.findByPlaceholderText("Search by hostname...");
      await user.type(
        screen.getByPlaceholderText("Search by hostname..."),
        "  myhost  ",
      );
      await waitFor(() => {
        const call = sdk.getContainers.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).toContain("myhost");
      });
    });
  });

  describe("tag mutation — onTagRenamed/onTagDeleted update URL tags array", () => {
    it("renames a tag in filter when onTagRenamed is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);
      await waitFor(() => expect(sdk.getContainers).toHaveBeenCalled());

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagRenamed?: (oldName: string, newName: string) => void;
      };
      expect(lastCall?.onTagRenamed).toBeDefined();

      await act(async () => {
        lastCall.onTagRenamed!("a", "alpha");
      });

      await waitFor(() => {
        const call = sdk.getContainers.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).toContain("alpha");
        expect(decoded).toContain('"b"');
      });
    });

    it("removes a tag from filter when onTagDeleted is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);
      await waitFor(() => expect(sdk.getContainers).toHaveBeenCalled());

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagDeleted?: (name: string) => void;
      };
      expect(lastCall?.onTagDeleted).toBeDefined();

      await act(async () => {
        lastCall.onTagDeleted!("a");
      });

      await waitFor(() => {
        const call = sdk.getContainers.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).not.toContain('"a"');
        expect(decoded).toContain('"b"');
      });
    });

    it("hydrates tags from URL into the SDK filter", async () => {
      renderPage(["/?tags=existing"]);
      await waitFor(() => {
        const call = sdk.getContainers.mock.calls.at(-1)?.[0] as {
          query?: { filter?: string };
        };
        const decoded = atob(call?.query?.filter ?? "");
        expect(decoded).toContain("existing");
      });
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });

  describe("status change — clicking a status tab resets page to 1", () => {
    it("resets page to 1 when status tab is clicked while on page 2", async () => {
      const user = userEvent.setup();
      renderPage(["/?page=2"]);
      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 2, status: "accepted" }),
          }),
        );
      });

      await user.click(screen.getByRole("tab", { name: "Pending" }));

      await waitFor(() => {
        expect(sdk.getContainers).toHaveBeenCalledWith(
          expect.objectContaining({
            query: expect.objectContaining({ page: 1, status: "pending" }),
          }),
        );
      });
    });
  });

  describe("action delegation — action buttons use useContainerActions", () => {
    it("calls requestAction(container, 'accept') when Accept is clicked in pending view", async () => {
      const user = userEvent.setup();
      sdk.getContainers.mockResolvedValue(
        paginatedResponse(
          [
            mockContainer({
              uid: "uid-pending",
              name: "pending-box",
              status: "pending",
              online: false,
            }),
          ],
          1,
        ),
      );
      renderPage(["/?status=pending"]);
      await user.click(await screen.findByRole("button", { name: "Accept" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({ uid: "uid-pending", name: "pending-box" }),
        "accept",
      );
    });

    it("calls requestAction(container, 'reject') when Reject is clicked in pending view", async () => {
      const user = userEvent.setup();
      sdk.getContainers.mockResolvedValue(
        paginatedResponse(
          [
            mockContainer({
              uid: "uid-pending-2",
              name: "pending-box-2",
              status: "pending",
              online: false,
            }),
          ],
          1,
        ),
      );
      renderPage(["/?status=pending"]);
      await user.click(await screen.findByRole("button", { name: "Reject" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({
          uid: "uid-pending-2",
          name: "pending-box-2",
        }),
        "reject",
      );
    });

    it("calls requestAction(container, 'remove') when Remove is clicked in rejected view", async () => {
      const user = userEvent.setup();
      sdk.getContainers.mockResolvedValue(
        paginatedResponse(
          [
            mockContainer({
              uid: "uid-rejected",
              name: "rejected-box",
              status: "rejected",
              online: false,
            }),
          ],
          1,
        ),
      );
      renderPage(["/?status=rejected"]);
      await user.click(await screen.findByRole("button", { name: "Remove" }));
      expect(mockRequestAction).toHaveBeenCalledWith(
        expect.objectContaining({
          uid: "uid-rejected",
          name: "rejected-box",
        }),
        "remove",
      );
    });
  });
});
