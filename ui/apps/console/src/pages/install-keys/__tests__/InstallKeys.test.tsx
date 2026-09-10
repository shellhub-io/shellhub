import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { createTestWrapper } from "@/tests/wrapper";
import { mockInstallKey } from "@/tests/factories";
import { paginatedResponse } from "@/tests/sdk";
import { seedAuthStore } from "@/tests/seedAuthStore";
import InstallKeys from "../index";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    installKeyList: vi.fn(),
  }),
);

function renderPage() {
  return render(<InstallKeys />, {
    wrapper: createTestWrapper({ initialEntries: ["/install-keys"] }),
  });
}

async function keyRow(name: string) {
  return within(await screen.findByRole("row", { name: new RegExp(name) }));
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  sdk.installKeyList.mockResolvedValue(
    paginatedResponse([
      mockInstallKey({
        id: "waiting-digest",
        name: "fleet-key",
        pending_devices: 2,
      }),
      mockInstallKey({
        id: "settled-digest",
        name: "auto-key",
        mode: "automatic",
        pending_devices: 0,
      }),
      mockInstallKey({
        id: "capped-digest",
        name: "edge-fleet",
        usage_limit: 4,
        used_times: 3,
        pending_devices: 3,
      }),
    ]),
  );
});

describe("Install keys", () => {
  it("says how many devices a key has waiting for a decision", async () => {
    renderPage();

    expect((await keyRow("fleet-key")).getByText("2 waiting")).toBeInTheDocument();
  });

  it("says nothing about waiting when a key has nothing pending", async () => {
    renderPage();

    expect((await keyRow("auto-key")).queryByText(/waiting/)).not.toBeInTheDocument();
  });

  it("keeps the spend readable while devices wait", async () => {
    renderPage();

    expect((await keyRow("fleet-key")).getByText(/0 \/ ∞ used/)).toBeInTheDocument();
  });

  it("warns when accepting everything waiting would pass the key's limit", async () => {
    renderPage();

    expect((await keyRow("edge-fleet")).getByText(/3 \/ 4 used · 2 over/)).toBeInTheDocument();
  });
});
