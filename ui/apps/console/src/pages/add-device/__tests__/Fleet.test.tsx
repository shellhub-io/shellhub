import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockProvisioningKey } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { fullText } from "@/tests/fullText";
import Fleet from "../Fleet";

function renderPage() {
  return render(
    <ClipboardProvider>
      <Fleet />
    </ClipboardProvider>,
    {
      wrapper: createTestWrapper({
        initialEntries: ["/devices/add/fleet"],
      }),
    },
  );
}

async function keyRow(name: string) {
  return within(await screen.findByRole("row", { name: new RegExp(name) }));
}

beforeEach(() => {
  vi.clearAllMocks();
  seedAuthStore();
  server.use(
    http.get("*/api/namespaces/provisioning-key/:key/reveal", ({ params }) =>
      HttpResponse.json({ key: `secret-of-${String(params.key)}` }),
    ),
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
        mockProvisioningKey({
          id: "dated-digest",
          name: "lab-key",
          mode: "automatic",
          expires_at: "2099-10-24T00:00:00Z",
          ephemeral: true,
          ephemeral_timeout: 15,
        }),
        mockProvisioningKey({
          id: "lapsed-digest",
          name: "lapsed-key",
          mode: "automatic",
          expires_at: "2020-01-01T00:00:00Z",
        }),
        mockProvisioningKey({
          id: "paused-digest",
          name: "paused-key",
          mode: "automatic",
          disabled: true,
          expires_at: "2020-01-01T00:00:00Z",
        }),
        mockProvisioningKey({
          id: "revoked-digest",
          name: "old-key",
          mode: "automatic",
          revoked: true,
        }),
      ]),
    ),
  );
});

describe("Fleet", () => {
  it("opens no install command until a key is picked", async () => {
    renderPage();
    await keyRow("fleet-key");

    expect(screen.queryByText(/PROVISIONING_KEY=/)).not.toBeInTheDocument();
  });

  it("puts the key picked in the list on the install command", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("row", { name: /auto-key/ }));

    expect(
      await screen.findByText(/PROVISIONING_KEY=secret-of-auto-key/),
    ).toBeInTheDocument();
  });

  it("folds the install command on a second click", async () => {
    const user = userEvent.setup();
    renderPage();
    const row = await screen.findByRole("row", { name: /auto-key/ });

    await user.click(row);
    await screen.findByText(/PROVISIONING_KEY=secret-of-auto-key/);
    await user.click(row);

    expect(screen.queryByText(/PROVISIONING_KEY=/)).not.toBeInTheDocument();
  });

  it("tells a role that cannot read keys why there is no command", async () => {
    const user = userEvent.setup();
    let reveals = 0;
    server.use(
      http.get("*/api/namespaces/provisioning-key/:key/reveal", () => {
        reveals += 1;
        return HttpResponse.json({ key: "unused" });
      }),
    );
    seedAuthStore({ role: "operator" });
    renderPage();

    await user.click(await screen.findByRole("row", { name: /auto-key/ }));

    expect(
      await screen.findByText(/Your role cannot read this key/),
    ).toBeInTheDocument();
    expect(screen.queryByText(/PROVISIONING_KEY=/)).not.toBeInTheDocument();
    expect(reveals).toBe(0);
  });

  it("keeps a revoked key off the install command", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("row", { name: /old-key/ }));

    expect(screen.queryByText(/PROVISIONING_KEY=/)).not.toBeInTheDocument();
  });

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
      (await keyRow("fleet-key")).getByText(fullText("0 of unlimited devices")),
    ).toBeInTheDocument();
  });

  it("warns when accepting everything waiting would pass the key's limit", async () => {
    renderPage();

    const row = await keyRow("edge-fleet");
    expect(row.getByText(fullText("3 of 4 devices"))).toBeInTheDocument();
    expect(row.getByText(/2 over/)).toBeInTheDocument();
  });

  it("folds a key's mode, expiry and cleanup under its name", async () => {
    renderPage();

    expect(
      (await keyRow("lab-key")).getByText(
        "Automatic · expires Oct 24, 2099 · removed after 15m offline",
      ),
    ).toBeInTheDocument();
  });

  it("says a revoked key is revoked instead of naming its mode", async () => {
    renderPage();

    const row = await keyRow("old-key");
    expect(row.getByText("Revoked")).toBeInTheDocument();
    expect(row.queryByText(/Automatic/)).not.toBeInTheDocument();
  });

  it("flags a key in use that has expired", async () => {
    renderPage();

    const line = (await keyRow("lapsed-key")).getByText(/^Automatic · expired/);
    expect(line).toHaveClass("text-accent-red");
  });

  it("keeps a disabled key quiet even once it has expired", async () => {
    renderPage();

    const line = (await keyRow("paused-key")).getByText(/^Disabled · expired/);
    expect(line).not.toHaveClass("text-accent-red");
  });
});
