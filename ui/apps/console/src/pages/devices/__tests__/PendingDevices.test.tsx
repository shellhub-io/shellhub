import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockProvisioningKeyEvent } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import PendingDevices from "../PendingDevices";

let accepted: string[];

beforeEach(() => {
  accepted = [];
  server.use(
    http.get("*/api/devices/tenant-registrations", () =>
      jsonWithTotal([
        mockProvisioningKeyEvent({
          id: "waiting",
          device_uid: "uid-waiting",
          hostname: "waiting-box",
          source_ip: "203.0.113.7",
          device_status: "pending",
        }),
        mockProvisioningKeyEvent({
          id: "rejected",
          device_uid: "uid-rejected",
          hostname: "rejected-box",
          device_status: "rejected",
          decided_status: "rejected",
        }),
      ]),
    ),
    http.patch("*/api/devices/:uid/accept", ({ params }) => {
      accepted.push(String(params.uid));
      return new HttpResponse(null, { status: 200 });
    }),
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(mockNamespace()),
    ),
  );
});

function renderPage() {
  return render(<PendingDevices />, {
    wrapper: createTestWrapper({ initialEntries: ["/devices/pending"] }),
  });
}

describe("PendingDevices", () => {
  it("shows the tenant ID registrations, the waiting one with what it reported", async () => {
    seedAuthStore({ role: "owner" });
    const user = userEvent.setup();
    renderPage();

    expect(await screen.findByText("waiting-box")).toBeInTheDocument();
    expect(screen.getByText("Waiting for you")).toBeInTheDocument();
    expect(screen.getByText("rejected-box")).toBeInTheDocument();

    await user.click(
      screen.getByRole("button", { name: "Show details for waiting-box" }),
    );
    expect(screen.getByText("203.0.113.7")).toBeInTheDocument();
  });

  it("lets an operator accept a waiting device and one rejected before", async () => {
    seedAuthStore({ role: "operator" });
    const user = userEvent.setup();
    renderPage();

    await user.click(await screen.findByRole("button", { name: "Accept" }));
    await user.click(
      await screen.findByRole("button", { name: "Accept device" }),
    );
    await waitFor(() => expect(accepted).toEqual(["uid-waiting"]));

    await user.click(screen.getByRole("button", { name: "Accept instead" }));
    await user.click(
      await screen.findByRole("button", { name: "Accept device" }),
    );
    await waitFor(() =>
      expect(accepted).toEqual(["uid-waiting", "uid-rejected"]),
    );
  });

  it("shows an agent refused for not proving its key with why, and nothing to accept", async () => {
    seedAuthStore({ role: "owner" });
    server.use(
      http.get("*/api/devices/tenant-registrations", () =>
        jsonWithTotal([
          mockProvisioningKeyEvent({
            id: "refused",
            device_uid: "uid-refused",
            hostname: "old-box",
            refusal: "unverified_device_key",
            attempts: 7,
            info: {
              id: "debian",
              pretty_name: "Debian",
              version: "v0.26.1",
              arch: "amd64",
              platform: "docker",
            },
          }),
        ]),
      ),
    );
    renderPage();

    expect(
      await screen.findByText(
        "Refused 7 times: agent v0.26.1 can't sign its connections.",
      ),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Device signature enforcement" }),
    ).toHaveAttribute("href", "/settings/ssh");
    expect(screen.getByText("old-box")).not.toHaveAttribute("href");
    expect(
      screen.queryByRole("button", { name: /accept|reject/i }),
    ).not.toBeInTheDocument();

    await userEvent
      .setup()
      .click(screen.getByRole("button", { name: "Show details for old-box" }));
    expect(screen.getByText("Last attempt")).toBeInTheDocument();
    expect(screen.queryByText("Registered")).not.toBeInTheDocument();
  });

  it("says so when no device registered with the tenant ID", async () => {
    seedAuthStore({ role: "owner" });
    server.use(
      http.get("*/api/devices/tenant-registrations", () => jsonWithTotal([])),
    );
    renderPage();

    expect(
      await screen.findByText(
        "No device has registered with the tenant ID yet.",
      ),
    ).toBeInTheDocument();
  });
});
