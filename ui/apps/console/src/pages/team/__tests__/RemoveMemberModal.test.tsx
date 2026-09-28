import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import RemoveMemberModal from "../RemoveMemberModal";

function renderModal() {
  server.use(
    http.get("*/api/devices", () =>
      jsonWithTotal([
        mockDevice({ uid: "uid-laptop", name: "laptop" }),
        mockDevice({ uid: "uid-desktop", name: "desktop" }),
      ]),
    ),
  );
  const onClose = vi.fn();
  render(
    <RemoveMemberModal
      tenantId="tenant-1"
      member={{ id: "user-7", email: "alice@example.com", role: "operator", status: "active", added_at: "" }}
      onClose={onClose}
    />,
    { wrapper: createTestWrapper() },
  );
  return onClose;
}

describe("RemoveMemberModal", () => {
  it("lists the devices that leave with the member", async () => {
    seedAuthStore({ role: "owner" });
    renderModal();

    const devices = await screen.findByRole("region", { name: "Paired devices" });
    expect(within(devices).getByText("laptop")).toBeInTheDocument();
    expect(devices).toHaveTextContent("2 devices this member paired will be removed");
    expect(screen.getByText(/API keys they created/i)).toBeInTheDocument();
  });

  it("removes the member, keeping the devices marked as team devices", async () => {
    seedAuthStore({ role: "owner" });
    let removed: URL | undefined;
    server.use(
      http.delete("*/api/namespaces/:tenant/members/:uid", ({ request }) => {
        removed = new URL(request.url);
        return HttpResponse.json({ tenant_id: "tenant-1" });
      }),
    );
    const onClose = renderModal();
    const user = userEvent.setup();

    await user.click(
      await screen.findByRole("checkbox", { name: "Keep laptop as a team device" }),
    );
    await user.click(screen.getByRole("button", { name: "Remove member" }));

    await waitFor(() => expect(onClose).toHaveBeenCalled());
    expect(removed?.pathname).toMatch(/\/api\/namespaces\/tenant-1\/members\/user-7$/);
    expect(removed?.searchParams.getAll("keep_devices")).toEqual(["uid-laptop"]);
  });

  it("stays open and says why when the removal is refused", async () => {
    seedAuthStore({ role: "owner" });
    server.use(
      http.delete("*/api/namespaces/:tenant/members/:uid", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    const onClose = renderModal();
    const user = userEvent.setup();

    await screen.findByRole("region", { name: "Paired devices" });
    await user.click(screen.getByRole("button", { name: "Remove member" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "You do not have permission to do this.",
    );
    expect(onClose).not.toHaveBeenCalled();
  });

  it("holds the removal until the member's devices have loaded", async () => {
    seedAuthStore({ role: "owner" });
    server.use(http.get("*/api/devices", () => new Promise(() => {})));
    render(
      <RemoveMemberModal
        tenantId="tenant-1"
        member={{ id: "user-7", email: "alice@example.com", role: "operator", status: "active", added_at: "" }}
        onClose={vi.fn()}
      />,
      { wrapper: createTestWrapper() },
    );

    expect(await screen.findByText(/loading the devices/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove member" })).toBeDisabled();
  });

  it("refuses the removal when the member's devices cannot be loaded", async () => {
    seedAuthStore({ role: "owner" });
    server.use(
      http.get("*/api/devices", () => HttpResponse.json({}, { status: 500 })),
    );
    render(
      <RemoveMemberModal
        tenantId="tenant-1"
        member={{ id: "user-7", email: "alice@example.com", role: "operator", status: "active", added_at: "" }}
        onClose={vi.fn()}
      />,
      { wrapper: createTestWrapper() },
    );

    expect(
      await screen.findByText(/couldn.t load the devices/i),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Remove member" })).toBeDisabled();
  });
});
