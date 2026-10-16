import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestQueryClient, createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import {
  mockLegacyNamespace,
  mockNamespace,
  mockProvisioningKeyEvent,
  mockStats,
} from "@/tests/factories";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import type { DeviceStatus } from "@/client";
import Dashboard from "../Dashboard";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function serveDashboard(legacy: boolean, statuses: DeviceStatus[]) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(legacy ? mockLegacyNamespace() : mockNamespace()),
    ),
    http.get("*/api/stats", () =>
      HttpResponse.json(mockStats({ pending_devices: 2 })),
    ),
    http.get("*/api/sessions", () => jsonWithTotal([])),
    http.get("*/api/devices/tenant-registrations", () =>
      jsonWithTotal(
        statuses.map((status, i) =>
          mockProvisioningKeyEvent({
            id: `registration-${i}`,
            device_status: status,
          }),
        ),
      ),
    ),
  );
}

function renderPage(queryClient = createTestQueryClient()) {
  return render(
    <>
      <Dashboard />
      <NamespaceModeProbe />
    </>,
    {
      wrapper: createTestWrapper({
        initialEntries: ["/dashboard"],
        queryClient,
      }),
    },
  );
}

beforeEach(() => {
  seedAuthStore();
});

describe("Dashboard", () => {
  it.each<{
    mode: string;
    legacy: boolean;
    statuses: DeviceStatus[];
    to: string;
  }>([
    { mode: "legacy", legacy: true, statuses: [], to: "/devices/pending" },
    {
      mode: "new",
      legacy: false,
      statuses: ["pending", "pending"],
      to: "/devices/pending",
    },
    {
      mode: "new",
      legacy: false,
      statuses: ["accepted"],
      to: "/devices/add/fleet",
    },
    { mode: "new", legacy: false, statuses: [], to: "/devices/add/fleet" },
  ])(
    "sends a $mode namespace with 2 pending devices and tenant ID registrations $statuses to $to",
    async ({ mode, legacy, statuses, to }) => {
      serveDashboard(legacy, statuses);
      const queryClient = createTestQueryClient();
      renderPage(queryClient);
      await screen.findByText(`mode: ${mode}`);
      await waitFor(() => expect(queryClient.isFetching()).toBe(0));

      expect(
        screen.getByRole("link", { name: /review pending/i }),
      ).toHaveAttribute("href", to);
    },
  );

  it("sends an operator, who cannot list provisioning keys, to the tenant ID queue", async () => {
    seedAuthStore({ role: "operator" });
    serveDashboard(false, ["accepted"]);
    const queryClient = createTestQueryClient();
    renderPage(queryClient);
    await screen.findByText("mode: new");
    await waitFor(() => expect(queryClient.isFetching()).toBe(0));

    expect(
      screen.getByRole("link", { name: /review pending/i }),
    ).toHaveAttribute("href", "/devices/pending");
  });

  it("offers no review to a member who may not accept devices", async () => {
    seedAuthStore({ role: "observer" });
    serveDashboard(true, []);
    renderPage();

    await screen.findByText("mode: legacy");
    expect(screen.getByText("Pending Devices")).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: /review pending/i }),
    ).not.toBeInTheDocument();
  });
});
