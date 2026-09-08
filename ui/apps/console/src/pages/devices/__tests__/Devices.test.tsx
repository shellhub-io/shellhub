import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import type { UseActionDialogResult } from "@/hooks/useActionDialog";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice, mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import Devices from "../index";

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

vi.mock("@/components/common/PlatformBadge", () => ({
  default: ({ platform }: { platform: string }) => <span>{platform}</span>,
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

vi.mock("@/components/common/ActionDialog", () => ({
  default: () => null,
}));

const mockRequestAction = vi.fn();
const mockDeviceActionsController: UseActionDialogResult = {
  action: undefined,
  actionKey: "closed",
  requestAction: mockRequestAction,
  close: vi.fn(),
  handleSuccess: vi.fn(),
};
vi.mock("@/hooks/useActionDialog", () => ({
  useActionDialog: vi.fn(() => mockDeviceActionsController),
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function renderPage(initialEntries: string[] = ["/"]) {
  return render(<Devices />, {
    wrapper: createTestWrapper({ initialEntries }),
  });
}

let lastDevicesUrl: URL | null;

function setDevices(
  devices: ReturnType<typeof mockDevice>[],
  total?: number,
) {
  server.use(
    http.get("*/api/devices", ({ request }) => {
      lastDevicesUrl = new URL(request.url);
      return jsonWithTotal(devices, total ?? devices.length);
    }),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  lastDevicesUrl = null;
  seedAuthStore();
  setDevices([]);
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: "jwt-token", role: "owner" }),
    ),
    http.post("*/api/tags", () => new HttpResponse(null, { status: 204 })),
    http.post(
      "*/api/devices/:uid/tags/:name",
      () => new HttpResponse(null, { status: 204 }),
    ),
    http.delete(
      "*/api/devices/:uid/tags/:name",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
  mockNavigate.mockReset();
  mockManageTagsDrawer.mockReset();
  mockRequestAction.mockReset();
});

describe("Devices list", () => {
  describe("rendering", () => {
    it("renders the page heading", async () => {
      renderPage();
      expect(
        await screen.findByRole("heading", { name: "Devices" }),
      ).toBeInTheDocument();
    });

    it("renders the Accepted tab and the Install Keys link, not the pending/rejected tabs", async () => {
      renderPage();
      expect(
        await screen.findByRole("button", { name: "Accepted" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "Install Keys" }),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Pending" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: "Rejected" }),
      ).not.toBeInTheDocument();
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
      server.use(
        http.get("*/api/devices", () => new Promise(() => {})),
      );
      renderPage();
      expect(screen.getByText("Loading devices...")).toBeInTheDocument();
    });
  });

  describe("empty state", () => {
    it('renders "No devices found" when list is empty', async () => {
      renderPage();
      expect(await screen.findByText("No devices found")).toBeInTheDocument();
    });
  });

  describe("device rows", () => {
    it("renders a row for each device", async () => {
      setDevices(
        [
          mockDevice({ uid: "uid-1", name: "alpha" }),
          mockDevice({ uid: "uid-2", name: "beta" }),
        ],
        2,
      );
      renderPage();
      expect(await screen.findByText("alpha")).toBeInTheDocument();
      expect(screen.getByText("beta")).toBeInTheDocument();
    });

    it("navigates to device detail on row click", async () => {
      const user = userEvent.setup();
      setDevices(
        [mockDevice({ uid: "uid-abc", name: "clickable" })],
        1,
      );
      renderPage();
      await user.click(await screen.findByText("clickable"));
      expect(mockNavigate).toHaveBeenCalledWith("/devices/uid-abc");
    });
  });

  describe("error state", () => {
    it("renders an error message when the query fails", async () => {
      server.use(
        http.get("*/api/devices", () =>
          HttpResponse.json({}, { status: 500 }),
        ),
      );
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
        expect(lastDevicesUrl).not.toBeNull();
        expect(lastDevicesUrl!.searchParams.get("sort_by")).toBe("last_seen");
        expect(lastDevicesUrl!.searchParams.get("order_by")).toBe("desc");
      });
    });

    it("toggles sort when the Hostname header is clicked", async () => {
      const user = userEvent.setup();
      setDevices([mockDevice({ uid: "uid-1", name: "alpha" })], 1);
      renderPage();
      await screen.findByText("alpha");

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      await waitFor(() => {
        expect(lastDevicesUrl!.searchParams.get("sort_by")).toBe("name");
        expect(lastDevicesUrl!.searchParams.get("order_by")).toBe("asc");
      });

      await user.click(
        screen.getByRole("button", { name: "Sort by Hostname" }),
      );
      await waitFor(() => {
        expect(lastDevicesUrl!.searchParams.get("sort_by")).toBe("name");
        expect(lastDevicesUrl!.searchParams.get("order_by")).toBe("desc");
      });
    });
  });

  describe("URL hydration — URL params seed page state on mount", () => {
    it("enforces status=accepted regardless of URL", async () => {
      renderPage(["/?status=pending&tags=a&tags=b&page=2"]);
      await waitFor(() => {
        expect(lastDevicesUrl?.searchParams.get("status")).toBe("accepted");
      });
    });

    it("passes tags from URL as a filter to the SDK", async () => {
      renderPage(["/?tags=a&tags=b"]);
      await waitFor(() => {
        expect(lastDevicesUrl).not.toBeNull();
        const filter = lastDevicesUrl!.searchParams.get("filter") ?? "";
        const decoded = atob(filter);
        expect(decoded).toContain('"a"');
        expect(decoded).toContain('"b"');
      });
    });

    it("passes page from URL to the SDK", async () => {
      renderPage(["/?page=2"]);
      await waitFor(() => {
        expect(lastDevicesUrl?.searchParams.get("page")).toBe("2");
      });
    });

    it("falls back to status=accepted and page=1 when URL has no params", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(lastDevicesUrl).not.toBeNull();
        expect(lastDevicesUrl!.searchParams.get("status")).toBe("accepted");
        expect(lastDevicesUrl!.searchParams.get("page")).toBe("1");
      });
    });

    it("falls back to status=accepted for an invalid status value", async () => {
      renderPage(["/?status=invalid"]);
      await waitFor(() => {
        expect(lastDevicesUrl).not.toBeNull();
        expect(lastDevicesUrl!.searchParams.get("status")).toBe("accepted");
      });
    });

    it("passes no tag filter when no tags param is present", async () => {
      renderPage(["/"]);
      await waitFor(() => {
        expect(lastDevicesUrl).not.toBeNull();
        expect(lastDevicesUrl!.searchParams.get("filter")).toBeNull();
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
        const filter = lastDevicesUrl!.searchParams.get("filter") ?? "";
        const decoded = atob(filter);
        expect(decoded).toContain("myhost");
      });
    });
  });

  describe("tag mutation — onTagRenamed/onTagDeleted update URL tags array", () => {
    it("renames a tag in filter when onTagRenamed is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);
      await waitFor(() => expect(lastDevicesUrl).not.toBeNull());

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagRenamed?: (oldName: string, newName: string) => void;
      };
      expect(lastCall?.onTagRenamed).toBeDefined();

      await act(async () => {
        lastCall.onTagRenamed!("a", "alpha");
      });

      await waitFor(() => {
        const filter = lastDevicesUrl!.searchParams.get("filter") ?? "";
        const decoded = atob(filter);
        expect(decoded).toContain("alpha");
        expect(decoded).toContain('"b"');
      });
    });

    it("removes a tag from filter when onTagDeleted is called from ManageTagsDrawer", async () => {
      renderPage(["/?tags=a&tags=b"]);
      await waitFor(() => expect(lastDevicesUrl).not.toBeNull());

      const lastCall = mockManageTagsDrawer.mock.calls.at(-1)?.[0] as {
        onTagDeleted?: (name: string) => void;
      };
      expect(lastCall?.onTagDeleted).toBeDefined();

      await act(async () => {
        lastCall.onTagDeleted!("a");
      });

      await waitFor(() => {
        const filter = lastDevicesUrl!.searchParams.get("filter") ?? "";
        const decoded = atob(filter);
        expect(decoded).not.toContain('"a"');
        expect(decoded).toContain('"b"');
      });
    });

    it("hydrates tags from URL into the SDK filter", async () => {
      renderPage(["/?tags=existing"]);
      await waitFor(() => {
        const filter = lastDevicesUrl!.searchParams.get("filter") ?? "";
        const decoded = atob(filter);
        expect(decoded).toContain("existing");
      });
      expect(
        screen.getByPlaceholderText("Search by hostname..."),
      ).toBeInTheDocument();
    });
  });
});
