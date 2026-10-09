import { describe, it, expect, beforeEach } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockMembershipInvitation } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import MembersTab from "../MembersTab";

const day = 24 * 60 * 60 * 1000;

const invitationTo = (email: string, expiresAt: Date) =>
  mockMembershipInvitation({
    user: { id: `id-${email}`, email },
    expires_at: expiresAt.toISOString(),
  });

let regenerated: unknown[];

beforeEach(() => {
  regenerated = [];
  seedAuthStore({ role: "owner" });
  server.use(
    http.get("*/api/namespaces/:tenant/members", () => jsonWithTotal([])),
    http.get("*/api/namespaces/:tenant/invitations", () =>
      jsonWithTotal([
        invitationTo("expired@example.com", new Date(Date.now() - day)),
        invitationTo("valid@example.com", new Date(Date.now() + day)),
      ]),
    ),
    http.post(
      "*/api/namespaces/:tenant/invitations/links",
      async ({ request }) => {
        regenerated.push(await request.json());
        return HttpResponse.json({ link: "https://shellhub.test/new" });
      },
    ),
  );
});

function renderTab() {
  render(
    <ClipboardProvider>
      <MembersTab tenantId="tenant-1" />
    </ClipboardProvider>,
    { wrapper: createTestWrapper() },
  );
}

const findInvitationRow = (email: string) =>
  screen.findByRole("row", { name: new RegExp(email) });

describe("MembersTab", () => {
  it("marks an invitation past its expiry as expired and offers to regenerate only that one", async () => {
    renderTab();

    const expired = await findInvitationRow("expired@example.com");
    const valid = await findInvitationRow("valid@example.com");

    expect(within(expired).getByText("expired")).toBeInTheDocument();
    expect(within(valid).queryByText("expired")).not.toBeInTheDocument();
    expect(
      within(expired).getByRole("button", {
        name: "Regenerate invitation link",
      }),
    ).toBeEnabled();
    expect(
      within(valid).getByRole("button", { name: "Regenerate invitation link" }),
    ).toBeDisabled();
  });

  it("regenerates the expired invitation for the same email and role once confirmed", async () => {
    const user = userEvent.setup();
    renderTab();

    const expired = await findInvitationRow("expired@example.com");
    await user.click(
      within(expired).getByRole("button", {
        name: "Regenerate invitation link",
      }),
    );
    const dialog = await screen.findByRole("dialog", {
      name: "Regenerate link",
    });
    await user.click(
      within(dialog).getByRole("button", { name: "Regenerate link" }),
    );

    await expect
      .poll(() => regenerated)
      .toEqual([{ email: "expired@example.com", role: "observer" }]);
  });
});
