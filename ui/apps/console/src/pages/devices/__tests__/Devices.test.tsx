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
  mockProvisioningKey,
  mockProvisioningKeyEvent,
  mockStats,
  mockTag,
} from "@/tests/factories";
import { connectDialogHandlers } from "@/tests/handlers";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import TerminalManager from "@/components/terminal/TerminalManager";
import { useTerminalStore } from "@/stores/terminalStore";
import Devices from "../index";

vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
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

vi.mock("@/components/ManageTagsModal", () => ({
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

function serveLegacyNamespace() {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockLegacyNamespace()),
    ),
  );
}

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

function servePendingDevices(count: number) {
  server.use(
    http.get("*/api/stats", () =>
      HttpResponse.json(mockStats({ pending_devices: count })),
    ),
  );
}

function serveKeys(keys: ReturnType<typeof mockProvisioningKey>[]) {
  server.use(
    http.get("*/api/namespaces/provisioning-key", () => jsonWithTotal(keys)),
  );
}

function requestedTags(): string[] {
  const filter = lastDevicesUrl?.searchParams.get("filter");
  if (!filter) return [];
  const parsed = JSON.parse(atob(filter)) as {
    params: { name: string; value?: unknown };
  }[];
  const tagFilter = parsed.find((f) => f.params.name === "tags.name");
  return (tagFilter?.params.value as string[] | undefined) ?? [];
}

function linkTargets() {
  return screen.queryAllByRole("link").map((link) => link.getAttribute("href"));
}

async function settle(
  queryClient: ReturnType<typeof createTestQueryClient>,
  mode: string,
) {
  await screen.findByText(`mode: ${mode}`);
  await waitFor(() => expect(queryClient.isFetching()).toBe(0));
}

beforeEach(() => {
  vi.clearAllMocks();
  lastDevicesUrl = null;
  seedAuthStore();
  useTerminalStore.setState({ sessions: [], reconnectTarget: null });
  setDevices([]);
  server.use(
    http.get("*/api/namespaces/provisioning-key", () => jsonWithTotal([])),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
    http.get("*/api/namespaces", () => jsonWithTotal([mockNamespace()])),
    http.get("*/api/stats", () => HttpResponse.json(mockStats())),
    http.get("*/api/devices/tenant-registrations", () => jsonWithTotal([])),
    http.get("*/api/tags", () =>
      jsonWithTotal([mockTag({ name: "edge" }), mockTag({ name: "prod" })]),
    ),
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
  it("labels no status tab in a new namespace", async () => {
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "new");
    expect(screen.queryByText("Accepted")).not.toBeInTheDocument();
  });

  it("labels no status tab in a legacy namespace", async () => {
    serveLegacyNamespace();
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "legacy");
    expect(screen.queryByText("Accepted")).not.toBeInTheDocument();
  });

  it("navigates to device detail on row click", async () => {
    const user = userEvent.setup();
    setDevices([mockDevice({ uid: "uid-abc", name: "clickable" })], 1);
    renderPage();
    await user.click(await screen.findByText("clickable"));
    expect(mockNavigate).toHaveBeenCalledWith("/devices/uid-abc");
  });

  it("opens the connect dialog from a row's Connect", async () => {
    const user = userEvent.setup();
    server.use(...connectDialogHandlers);
    setDevices([mockDevice({ name: "edge-1", online: true })], 1);
    render(
      <>
        <Devices />
        <TerminalManager />
      </>,
      { wrapper: createTestWrapper({ initialEntries: ["/"] }) },
    );

    await user.click(await screen.findByRole("button", { name: "Connect" }));

    expect(
      await screen.findByRole("dialog", { name: "Connect" }),
    ).toBeInTheDocument();
    expect(mockNavigate).not.toHaveBeenCalled();
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

describe("Devices waiting for approval in a new namespace", () => {
  it("counts the keys' waiting devices and opens onto each key holding one", async () => {
    const user = userEvent.setup();
    servePendingDevices(9);
    serveKeys([
      mockProvisioningKey({
        id: "tenant-only",
        type: "legacy",
        name: "legacy",
        pending_devices: 1,
      }),
      mockProvisioningKey({
        id: "rollout",
        name: "rollout-lab",
        pending_devices: 3,
      }),
      mockProvisioningKey({ id: "idle", name: "idle-key", pending_devices: 0 }),
    ]);
    renderPage();

    await user.click(
      await screen.findByRole("button", { name: /waiting for approval 4/i }),
    );

    expect(
      screen.getByRole("link", { name: /rollout-lab.*3 waiting/i }),
    ).toHaveAttribute("href", "/devices/add/fleet/rollout/activity");
    expect(
      screen.getByRole("link", {
        name: /tenant-only registration.*1 waiting/i,
      }),
    ).toHaveAttribute("href", "/devices/add/fleet/tenant-only/activity");
    expect(screen.queryByText("idle-key")).not.toBeInTheDocument();
  });

  it("never links the tenant ID queue", async () => {
    servePendingDevices(2);
    serveRegistrations(["pending", "pending"]);
    serveKeys([
      mockProvisioningKey({
        id: "tenant-only",
        type: "legacy",
        name: "legacy",
        pending_devices: 2,
      }),
    ]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "new");
    expect(linkTargets()).not.toContain("/devices/pending");
  });

  it("offers nothing while no device waits", async () => {
    servePendingDevices(0);
    serveKeys([mockProvisioningKey({ pending_devices: 0 })]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "new");
    expect(
      screen.queryByRole("button", { name: /waiting for approval/i }),
    ).not.toBeInTheDocument();
  });

  it("offers nothing to a member who may not accept devices", async () => {
    seedAuthStore({ role: "observer" });
    servePendingDevices(3);
    serveKeys([mockProvisioningKey({ pending_devices: 3 })]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "new");
    expect(
      screen.queryByRole("button", { name: /waiting for approval/i }),
    ).not.toBeInTheDocument();
  });
});

describe("Devices waiting for approval in a legacy namespace", () => {
  it("links the tenant ID queue with its count", async () => {
    serveLegacyNamespace();
    serveRegistrations(["pending", "pending", "pending", "rejected"]);
    renderPage();

    expect(
      await screen.findByRole("link", { name: /waiting for approval 3/i }),
    ).toHaveAttribute("href", "/devices/pending");
  });

  it("offers nothing while no registration waits", async () => {
    serveLegacyNamespace();
    serveRegistrations(["rejected"]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "legacy");
    expect(linkTargets()).not.toContain("/devices/pending");
  });

  it("offers nothing to a member who may not accept devices", async () => {
    seedAuthStore({ role: "observer" });
    serveLegacyNamespace();
    serveRegistrations(["pending"]);
    const queryClient = createTestQueryClient();
    renderPage(["/"], queryClient);

    await settle(queryClient, "legacy");
    expect(linkTargets()).not.toContain("/devices/pending");
  });
});

describe("Devices toolbar", () => {
  it("shows each tag filter from the URL as a chip and counts them on Filters", async () => {
    renderPage(["/?tags=prod&tags=edge"]);

    expect(
      await screen.findByRole("button", { name: "Remove the prod filter" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Remove the edge filter" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: /filters 2/i }),
    ).toBeInTheDocument();
  });

  it("filters by a tag chosen in Filters", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: /^filters/i }));
    await user.click(await screen.findByRole("button", { name: "prod" }));

    await waitFor(() => {
      expect(requestedTags()).toEqual(["prod"]);
    });
    await user.keyboard("{Escape}");
    expect(
      screen.getByRole("button", { name: "Remove the prod filter" }),
    ).toBeInTheDocument();
  });

  it("drops a tag from the request when its chip is removed", async () => {
    const user = userEvent.setup();
    renderPage(["/?tags=prod&tags=edge"]);

    await user.click(
      await screen.findByRole("button", { name: "Remove the prod filter" }),
    );

    await waitFor(() => {
      expect(requestedTags()).toEqual(["edge"]);
    });
    expect(
      screen.queryByRole("button", { name: "Remove the prod filter" }),
    ).not.toBeInTheDocument();
  });

  it("drops every tag when Clear all is chosen", async () => {
    const user = userEvent.setup();
    renderPage(["/?tags=prod&tags=edge"]);

    await user.click(await screen.findByRole("button", { name: "Clear all" }));

    await waitFor(() => {
      expect(requestedTags()).toEqual([]);
    });
  });

  it("links Add Device to the install page", async () => {
    renderPage();

    expect(
      await screen.findByRole("link", { name: /add device/i }),
    ).toHaveAttribute("href", "/devices/add");
  });
});
