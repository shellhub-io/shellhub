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

function renderPage(path = "/devices/pending") {
  return render(<PendingDevices />, {
    wrapper: createTestWrapper({ initialEntries: [path] }),
  });
}

function serveRegistrationsByHostname(...hostnames: string[]) {
  const registrations = hostnames.map((hostname) =>
    mockProvisioningKeyEvent({
      id: hostname,
      device_uid: `uid-${hostname}`,
      hostname,
      device_status: "pending",
    }),
  );
  server.use(
    http.get("*/api/devices/tenant-registrations", ({ request }) => {
      const hostname =
        new URL(request.url).searchParams.get("filters[hostname][contains]") ??
        "";
      return jsonWithTotal(
        registrations.filter((r) => r.hostname.includes(hostname)),
      );
    }),
  );
}

function searchbox() {
  return screen.getByRole("searchbox", {
    name: "Search registrations by hostname",
  });
}

function findSortButton(order: string) {
  return screen.findByRole("button", { name: `Sort registrations: ${order}` });
}

function recordOrders() {
  const orders: string[] = [];
  server.use(
    http.get("*/api/devices/tenant-registrations", ({ request }) => {
      const params = new URL(request.url).searchParams;
      orders.push(`${params.get("sort_by")}:${params.get("order_by")}`);
      return jsonWithTotal([
        mockProvisioningKeyEvent({ id: "a", hostname: "alpha-box" }),
      ]);
    }),
  );
  return orders;
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

  it("narrows the registrations to the hostnames containing the search, each still decidable", async () => {
    seedAuthStore({ role: "owner" });
    serveRegistrationsByHostname("kitchen-pi", "garage-pi");
    const user = userEvent.setup();
    renderPage();

    expect(await screen.findByText("garage-pi")).toBeInTheDocument();
    await user.type(searchbox(), "kitchen");

    await waitFor(() =>
      expect(screen.queryByText("garage-pi")).not.toBeInTheDocument(),
    );
    expect(
      screen.getByRole("link", { name: "kitchen-pi" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Accept" })).toBeInTheDocument();
  });

  it("marks the searched text in each matching hostname", async () => {
    seedAuthStore({ role: "owner" });
    serveRegistrationsByHostname("kitchen-pi", "garage-pi");
    renderPage("/devices/pending?search=kitchen");

    expect(await screen.findByText("kitchen")).toBeInTheDocument();
  });

  it("counts the registrations the search matches", async () => {
    seedAuthStore({ role: "owner" });
    serveRegistrationsByHostname("kitchen-pi", "garage-pi");
    renderPage("/devices/pending?search=kitchen");

    expect(
      await screen.findByText(/1 registration with .kitchen. in the hostname/),
    ).toBeInTheDocument();
  });

  it("says when no registration matches the search, and shows them all again once cleared", async () => {
    seedAuthStore({ role: "owner" });
    serveRegistrationsByHostname("kitchen-pi", "garage-pi");
    const user = userEvent.setup();
    renderPage();

    await screen.findByText("kitchen-pi");
    const search = searchbox();
    await user.type(search, "office");
    expect(
      await screen.findByText("No registrations matching \u201coffice\u201d"),
    ).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Clear search" }));
    expect(await screen.findByText("kitchen-pi")).toBeInTheDocument();
    expect(screen.getByText("garage-pi")).toBeInTheDocument();
    expect(search).toHaveValue("");
  });

  it("opens already narrowed by the search the link carries", async () => {
    seedAuthStore({ role: "owner" });
    serveRegistrationsByHostname("kitchen-pi", "garage-pi");
    renderPage("/devices/pending?search=garage");

    expect(
      await screen.findByRole("link", { name: "garage-pi" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("link", { name: "kitchen-pi" }),
    ).not.toBeInTheDocument();
    expect(searchbox()).toHaveValue("garage");
  });

  it("reads the registrations oldest first when asked", async () => {
    seedAuthStore({ role: "owner" });
    const orders = recordOrders();
    const user = userEvent.setup();
    renderPage();

    await user.click(await findSortButton("Newest first"));
    await user.click(
      await screen.findByRole("menuitem", { name: "Oldest first" }),
    );

    await waitFor(() => expect(orders.at(-1)).toBe("created_at:asc"));
    expect(orders[0]).toBe("created_at:desc");
    expect(await findSortButton("Oldest first")).toBeInTheDocument();
  });

  it("offers newest and oldest first as the only orders", async () => {
    seedAuthStore({ role: "owner" });
    const user = userEvent.setup();
    renderPage();

    await user.click(await findSortButton("Newest first"));

    expect(
      (await screen.findAllByRole("menuitem")).map((item) => item.textContent),
    ).toEqual(["Newest first", "Oldest first"]);
  });

  it("opens in the order the link carries", async () => {
    seedAuthStore({ role: "owner" });
    const orders = recordOrders();
    renderPage("/devices/pending?sort=oldest");

    expect(await findSortButton("Oldest first")).toBeInTheDocument();
    await waitFor(() => expect(orders).toEqual(["created_at:asc"]));
  });

  it("shows the history loading, not empty, while clearing the search refetches it", async () => {
    seedAuthStore({ role: "owner" });
    let release: (() => void) | undefined;
    server.use(
      http.get("*/api/devices/tenant-registrations", async ({ request }) => {
        const hostname = new URL(request.url).searchParams.get(
          "filters[hostname][contains]",
        );
        if (hostname === "office") return jsonWithTotal([]);
        if (hostname === null) {
          await new Promise<void>((resolve) => {
            release = resolve;
          });
        }
        return jsonWithTotal([
          mockProvisioningKeyEvent({
            id: "kitchen",
            hostname: "kitchen-pi",
            device_status: "pending",
          }),
        ]);
      }),
    );
    const user = userEvent.setup();
    renderPage("/devices/pending?search=office");

    await user.click(
      await screen.findByRole("button", { name: "Clear search" }),
    );

    await waitFor(() => expect(release).toBeDefined());
    expect(screen.getByText("Loading activity…")).toBeInTheDocument();
    expect(
      screen.queryByText(/No registrations matching/),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByText("No device has registered with the tenant ID yet."),
    ).not.toBeInTheDocument();
    expect(screen.queryByText(/registrations? with/)).not.toBeInTheDocument();

    release?.();
    expect(await screen.findByText("kitchen-pi")).toBeInTheDocument();
  });
});
