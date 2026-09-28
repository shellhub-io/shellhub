import { describe, it, expect } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import PairedBy from "../PairedBy";

function renderPairedBy(ownerId?: string) {
  server.use(
    http.get("*/api/namespaces/:tenant/members", () =>
      jsonWithTotal([
        { id: "user-7", email: "alice@example.com", role: "operator" },
      ]),
    ),
  );
  render(
    <PairedBy
      tenantId="tenant-1"
      uid="device-1"
      name="laptop"
      ownerId={ownerId}
    />,
    { wrapper: createTestWrapper() },
  );
}

describe("PairedBy", () => {
  it("shows the member who paired the device", async () => {
    seedAuthStore({ role: "owner" });
    renderPairedBy("user-7");

    expect(await screen.findByText("alice@example.com")).toBeInTheDocument();
  });

  it("shows a dash for a team device", () => {
    seedAuthStore({ role: "owner" });
    renderPairedBy();

    expect(screen.getByText("—")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Make team device" }),
    ).not.toBeInTheDocument();
  });

  it("makes the device a team device once confirmed", async () => {
    seedAuthStore({ role: "administrator" });
    let cleared: string | undefined;
    server.use(
      http.delete("*/api/devices/:uid/owner", ({ params }) => {
        cleared = String(params.uid);
        return new HttpResponse(null, { status: 204 });
      }),
    );
    renderPairedBy("user-7");
    const user = userEvent.setup();

    await user.click(screen.getByRole("button", { name: "Make team device" }));
    const dialog = await screen.findByRole("dialog", {
      name: "Make team device",
    });
    expect(dialog).toHaveTextContent("cannot be undone");
    await user.click(
      within(dialog).getByRole("button", { name: "Make team device" }),
    );

    await waitFor(() => expect(cleared).toBe("device-1"));
    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
  });

  it("offers nothing to a role that cannot create provisioning keys", () => {
    seedAuthStore({ role: "operator" });
    renderPairedBy("user-7");

    expect(
      screen.queryByRole("button", { name: "Make team device" }),
    ).not.toBeInTheDocument();
  });
});
