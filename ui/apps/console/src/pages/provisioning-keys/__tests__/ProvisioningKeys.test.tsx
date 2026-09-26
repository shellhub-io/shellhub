import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockProvisioningKey } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import ProvisioningKeys from "../index";

function renderPage() {
  return render(<ProvisioningKeys />, {
    wrapper: createTestWrapper({
      initialEntries: ["/settings/provisioning-keys"],
    }),
  });
}

async function keyRow(name: string) {
  return within(await screen.findByRole("row", { name: new RegExp(name) }));
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/provisioning-key", () =>
      jsonWithTotal([
        mockProvisioningKey({
          id: "waiting-digest",
          name: "fleet-key",
          pending_devices: 2,
        }),
        mockProvisioningKey({
          id: "settled-digest",
          name: "auto-key",
          mode: "automatic",
          pending_devices: 0,
        }),
        mockProvisioningKey({
          id: "capped-digest",
          name: "edge-fleet",
          usage_limit: 4,
          used_times: 3,
          pending_devices: 3,
        }),
      ]),
    ),
  );
});

describe("Provisioning keys", () => {
  it("says how many devices a key has waiting for a decision", async () => {
    renderPage();

    expect(
      (await keyRow("fleet-key")).getByText("2 waiting"),
    ).toBeInTheDocument();
  });

  it("says nothing about waiting when a key has nothing pending", async () => {
    renderPage();

    expect(
      (await keyRow("auto-key")).queryByText(/waiting/),
    ).not.toBeInTheDocument();
  });

  it("keeps the spend readable while devices wait", async () => {
    renderPage();

    expect(
      (await keyRow("fleet-key")).getByText("0 used"),
    ).toBeInTheDocument();
  });

  it("warns when accepting everything waiting would pass the key's limit", async () => {
    renderPage();

    const row = await keyRow("edge-fleet");
    expect(row.getByText("3 / 4")).toBeInTheDocument();
    expect(row.getByText(/2 over/)).toBeInTheDocument();
  });
});
