import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import {
  mockDevice,
  mockLegacyNamespace,
  mockNamespace,
  mockProvisioningKeyEvent,
  mockStats,
} from "@/tests/factories";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
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

vi.mock("@/components/ManageTagsModal", () => ({
  default: () => <div />,
}));

vi.mock("@/components/ConnectModal", () => ({
  default: () => <div />,
}));

vi.mock("@/components/common/TagsPopover", () => ({
  default: ({ tags }: { tags: string[] }) => (
    <span>{tags.length > 0 ? tags.join(", ") : "No tags"}</span>
  ),
}));

vi.mock("@/components/common/RestrictedAction", () => ({
  default: ({ children }: { children: React.ReactNode }) => <>{children}</>,
}));

const mockNavigate = vi.fn();
vi.mock("react-router-dom", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router-dom")>();
  return { ...actual, useNavigate: () => mockNavigate };
});

function renderPage(
  initialEntries: string[] = ["/"],
  queryClient = createTestQueryClient(),
) {
  return render(
    <>
      <Devices />
      <NamespaceModeProbe />
    </>,
    {
      wrapper: createTestWrapper({ initialEntries, queryClient }),
    },
  );
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
    http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
    http.get("*/api/stats", () => HttpResponse.json(mockStats())),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json({ token: VALID_JWT, role: "owner" }),
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
});

describe("Devices list", () => {
  it("shows the accepted label, not the pending/rejected tabs", async () => {
    renderPage();
    expect(await screen.findByText("Accepted")).toBeInTheDocument();
    expect(screen.queryByText("Pending")).not.toBeInTheDocument();
    expect(screen.queryByText("Rejected")).not.toBeInTheDocument();
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

describe("Devices empty state", () => {
  it("offers to continue the first-run setup to a member who can accept devices", async () => {
    seedAuthStore({ role: "owner" });
    renderPage();

    expect(
      await screen.findByRole("link", { name: /continue setup/i }),
    ).toHaveAttribute("href", "/dashboard");
  });

  it("leaves the setup link out when the user has several namespaces", async () => {
    seedAuthStore({ role: "owner" });
    server.use(
      http.get("*/api/namespaces", () =>
        jsonWithTotal([
          mockNamespace({ tenant_id: "a" }),
          mockNamespace({ tenant_id: "b" }),
        ]),
      ),
    );
    renderPage();

    expect(await screen.findByText(/no devices found/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /continue setup/i }),
    ).not.toBeInTheDocument();
  });

  it("leaves the setup link out for a member who cannot accept devices", async () => {
    seedAuthStore({ role: "observer" });
    renderPage();

    expect(await screen.findByText(/no devices found/i)).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /continue setup/i }),
    ).not.toBeInTheDocument();
  });
});

describe("Devices pending queue", () => {
  function serveRegistrations(statuses: ("pending" | "rejected")[]) {
    server.use(
      http.get("*/api/devices/tenant-registrations", () =>
        jsonWithTotal(
          statuses.map((status, i) =>
            mockProvisioningKeyEvent({
              id: `registration-${i}`,
              device_uid: `uid-${i}`,
              device_status: status,
            }),
          ),
        ),
      ),
    );
  }

  it("is not offered by a new namespace no device registered with the tenant ID", async () => {
    serveRegistrations([]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await screen.findByText("mode: new");
    await waitFor(() => expect(queryClient.isFetching()).toBe(0));
    expect(
      screen.queryByRole("link", { name: /pending$/ }),
    ).not.toBeInTheDocument();
  });

  it("is linked from a new namespace with the count of devices waiting", async () => {
    serveRegistrations(["pending", "pending", "rejected"]);
    renderPage();

    expect(
      await screen.findByRole("link", { name: "2 pending" }),
    ).toHaveAttribute("href", "/devices/pending");
  });

  it("stays linked while only a rejected registration remains to reconsider", async () => {
    serveRegistrations(["rejected"]);
    renderPage();

    expect(
      await screen.findByRole("link", { name: "0 pending" }),
    ).toHaveAttribute("href", "/devices/pending");
  });

  it("is linked from a legacy namespace with its count", async () => {
    server.use(
      http.get("*/api/namespaces/:tenant", () =>
        HttpResponse.json(mockLegacyNamespace()),
      ),
    );
    serveRegistrations(["pending", "pending", "pending"]);
    renderPage();

    expect(
      await screen.findByRole("link", { name: "3 pending" }),
    ).toHaveAttribute("href", "/devices/pending");
  });

  it("is not offered to a member who may not accept devices", async () => {
    seedAuthStore({ role: "observer" });
    server.use(
      http.get("*/api/namespaces/:tenant", () =>
        HttpResponse.json(mockLegacyNamespace()),
      ),
    );
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await screen.findByText("mode: legacy");
    await waitFor(() => expect(queryClient.isFetching()).toBe(0));
    expect(
      screen.queryByRole("link", { name: /pending$/ }),
    ).not.toBeInTheDocument();
  });
});
