import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
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

vi.mock("@/components/ManageTagsDrawer", () => ({
  default: () => <div />,
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

function setDevices(devices: ReturnType<typeof mockDevice>[], total?: number) {
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
  mockRequestAction.mockReset();
});

describe("Devices list", () => {
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

  it("navigates to device detail on row click", async () => {
    const user = userEvent.setup();
    setDevices([mockDevice({ uid: "uid-abc", name: "clickable" })], 1);
    renderPage();
    await user.click(await screen.findByText("clickable"));
    expect(mockNavigate).toHaveBeenCalledWith("/devices/uid-abc");
  });

  it("forces status=accepted even when the URL asks for another status", async () => {
    renderPage(["/?status=pending"]);
    await waitFor(() => {
      expect(lastDevicesUrl?.searchParams.get("status")).toBe("accepted");
    });
  });

  it("pages forward, and sorting a column re-requests it from the first page", async () => {
    const user = userEvent.setup();
    setDevices([mockDevice({ uid: "uid-abc", name: "paged" })], 25);
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Next page" }));
    await waitFor(() => {
      expect(lastDevicesUrl?.searchParams.get("page")).toBe("2");
    });

    await user.click(screen.getByRole("button", { name: "Sort by Hostname" }));
    await waitFor(() => {
      expect(lastDevicesUrl?.searchParams.get("sort_by")).toBe("name");
    });
    expect(lastDevicesUrl?.searchParams.get("order_by")).toBe("asc");
    expect(lastDevicesUrl?.searchParams.get("page")).toBe("1");
  });
});
