import { describe, it, expect } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import MemberRoleDialog from "../MemberRoleDialog";

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
    expect(request?.url).toMatch(/\/api\/namespaces\/tenant-1\/members\/user-7$/);
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
});
