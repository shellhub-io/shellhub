import { describe, it, expect, beforeEach, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import {
  mockLegacyNamespace,
  mockNamespace,
  mockProvisioningKey,
  mockStats,
} from "@/tests/factories";
import { NamespaceModeProbe } from "@/tests/NamespaceModeProbe";
import Dashboard from "../Dashboard";

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

function serveDashboard(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/provisioning-key", () =>
      jsonWithTotal([
        mockProvisioningKey({
          id: "legacy-digest",
          name: "legacy",
          type: "legacy",
        }),
      ]),
    ),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(legacy ? mockLegacyNamespace() : mockNamespace()),
    ),
    http.get("*/api/stats", () =>
      HttpResponse.json(mockStats({ pending_devices: 2 })),
    ),
    http.get("*/api/sessions", () => jsonWithTotal([])),
  );
}

function renderPage() {
  return render(
    <>
      <Dashboard />
      <NamespaceModeProbe />
    </>,
    {
      wrapper: createTestWrapper({ initialEntries: ["/dashboard"] }),
    },
  );
}

beforeEach(() => {
  seedAuthStore();
});

describe("Dashboard", () => {
  it("sends a new namespace's pending count to Fleet", async () => {
    serveDashboard(false);
    renderPage();
    await screen.findByText("mode: new");

    expect(
      await screen.findByRole("link", { name: /review pending/i }),
    ).toHaveAttribute("href", "/devices/add/fleet");
  });

  it("sends a legacy namespace's pending count to the pending queue", async () => {
    serveDashboard(true);
    renderPage();

    const link = await screen.findByRole("link", { name: /review pending/i });
    await waitFor(() =>
      expect(link).toHaveAttribute(
        "href",
        "/devices/add/fleet/legacy-digest/activity",
      ),
    );
  });
});
