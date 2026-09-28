import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockDevice } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import MemberRoleDialog from "../MemberRoleDialog";

function servePairedDevices(names: string[]) {
  const filters: unknown[] = [];
  server.use(
    http.get("*/api/devices", ({ request }) => {
      const filter = new URL(request.url).searchParams.get("filter") ?? "";
      filters.push(JSON.parse(atob(filter)));
      return jsonWithTotal(
        names.map((name) => mockDevice({ uid: `uid-${name}`, name })),
      );
    }),
  );
  return filters;
}

function renderDialog(role = "operator") {
  render(
    <MemberRoleDialog
      tenantId="tenant-1"
      memberId="user-7"
      email="alice@example.com"
      role={role}
    />,
    { wrapper: createTestWrapper() },
  );
}

async function openDialog() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("button", { name: "Edit role" }));
  await screen.findByRole("dialog", { name: "Edit role" });
  return user;
}

describe("MemberRoleDialog", () => {
  it("offers saving only once the role changes", async () => {
    renderDialog();
    const user = await openDialog();

    expect(screen.getByRole("button", { name: "Save role" })).toBeDisabled();
    await user.click(screen.getByRole("radio", { name: /administrator/i }));
    expect(screen.getByRole("button", { name: "Save role" })).toBeEnabled();
  });

  it("starts an owner's picker at operator", async () => {
    renderDialog("owner");
    await openDialog();

    expect(screen.getByRole("radio", { name: /operator/i })).toBeChecked();
  });

  it("sends the new role for that member and closes", async () => {
    let request: { url: string; body: unknown } | undefined;
    server.use(
      http.patch(
        "*/api/namespaces/:tenant/members/:uid",
        async ({ request: req }) => {
          request = { url: req.url, body: await req.json() };
          return HttpResponse.json({});
        },
      ),
    );
    renderDialog();
    const user = await openDialog();

    await user.click(screen.getByRole("radio", { name: /administrator/i }));
    await user.click(screen.getByRole("button", { name: "Save role" }));

    await waitFor(() =>
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument(),
    );
    expect(request?.url).toMatch(
      /\/api\/namespaces\/tenant-1\/members\/user-7$/,
    );
    expect(request?.body).toEqual({ role: "administrator" });
  });

  it("keeps the dialog open and says why when the change is refused", async () => {
    server.use(
      http.patch("*/api/namespaces/:tenant/members/:uid", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    renderDialog();
    const user = await openDialog();

    await user.click(screen.getByRole("radio", { name: /administrator/i }));
    await user.click(screen.getByRole("button", { name: "Save role" }));

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "You do not have permission to do this.",
    );
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  describe("demoting to a role that cannot accept devices", () => {
    it("lists the devices the member paired, which leave with the role", async () => {
      seedAuthStore({ role: "owner" });
      const filters = servePairedDevices(["laptop", "desktop"]);
      renderDialog();
      const user = await openDialog();

      await user.click(screen.getByRole("radio", { name: /observer/i }));

      const devices = await screen.findByRole("region", {
        name: "Paired devices",
      });
      expect(devices).toHaveTextContent(
        "2 devices this member paired will be removed",
      );
      expect(devices).toHaveTextContent("laptop");
      expect(filters).toContainEqual([
        {
          type: "property",
          params: { name: "owner_id", operator: "eq", value: "user-7" },
        },
      ]);
    });

    it("sends the devices kept as team devices", async () => {
      seedAuthStore({ role: "owner" });
      servePairedDevices(["laptop", "desktop"]);
      let body: unknown;
      server.use(
        http.patch(
          "*/api/namespaces/:tenant/members/:uid",
          async ({ request }) => {
            body = await request.json();
            return HttpResponse.json({});
          },
        ),
      );
      renderDialog();
      const user = await openDialog();

      await user.click(screen.getByRole("radio", { name: /observer/i }));
      await user.click(
        await screen.findByRole("checkbox", {
          name: "Keep laptop as a team device",
        }),
      );
      expect(
        screen.getByRole("region", { name: "Paired devices" }),
      ).toHaveTextContent("1 device this member paired will be removed");
      await user.click(screen.getByRole("button", { name: "Save role" }));

      await waitFor(() =>
        expect(body).toEqual({
          role: "observer",
          keep_devices: ["uid-laptop"],
        }),
      );
    });

    it("holds the demotion until the member's devices have loaded", async () => {
      seedAuthStore({ role: "owner" });
      server.use(http.get("*/api/devices", () => new Promise(() => {})));
      renderDialog();
      const user = await openDialog();

      await user.click(screen.getByRole("radio", { name: /observer/i }));

      expect(await screen.findByText(/loading the devices/i)).toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Save role" })).toBeDisabled();
    });

    it("offers no keeping to a caller who cannot create provisioning keys", async () => {
      seedAuthStore({ role: "operator" });
      servePairedDevices(["laptop"]);
      renderDialog("observer");
      await openDialog();

      await screen.findByRole("region", { name: "Paired devices" });
      expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
    });
  });
});
