import { type Browser, type Page, expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";
import {
  getNamespaceMembershipInvitationList,
  acceptInvite,
  apiKeyCreate,
  listAccessPolicies,
  listNamespaceMembers,
} from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import { isCommunity, isEnterprise } from "./env";
import { signIn, dismissWizard, directMembershipReason } from "./helpers";
import {
  password,
  buildShortId,
  buildRandomEmail,
  createUser,
  createNamespace,
  addMember,
  expireInvitation,
} from "./seed";
import {
  type Endpoint,
  buildRequestContext,
  expectStatus,
  invite,
  loginAs,
} from "./api";

async function createTeam() {
  const owner = createUser("owner");
  const namespace = `e2e-team-${buildShortId()}`;
  const tenant = randomUUID();
  createNamespace(owner.username, namespace, tenant);
  const { token } = await loginAs(owner.username, password);
  return { owner: { ...owner, token }, namespace, tenant };
}

async function createTeamWithMember(role: AssignableRole) {
  const team = await createTeam();
  const member = createUser("member");
  addMember(member.username, team.namespace, role);
  const { token, tenant } = await loginAs(member.username, password);
  if (tenant !== team.tenant) {
    throw new Error(`${member.username} did not land in ${team.namespace}`);
  }
  return { ...team, member: { ...member, token } };
}

async function createTeamWithMemberKey() {
  const team = await createTeamWithMember("administrator");
  const { data } = await apiKeyCreate({
    ...buildRequestContext({ token: team.member.token }),
    body: { name: `e2e-${buildShortId()}`, expires_at: -1 },
  });
  return { ...team, member: { ...team.member, apiKey: data.key } };
}

type MemberWithKey = { token: string; apiKey: string };

async function expectMemberStatus(
  endpoint: Endpoint,
  { token, apiKey }: MemberWithKey,
  status: number,
) {
  await expectStatus(endpoint, { token }, status);
  await expectStatus(endpoint, { apiKey }, status);
}

const listMembers =
  (tenant: string): Endpoint =>
    (opts) =>
      listNamespaceMembers({ ...opts, path: { tenant } });

async function rejoinAndExpectKeyRevoked({
  member,
  namespace,
  tenant,
}: Awaited<ReturnType<typeof createTeamWithMemberKey>>) {
  addMember(member.username, namespace, "administrator");
  await expectStatus(listMembers(tenant), { apiKey: member.apiKey }, 401);
}

async function expectInvitationPage(
  browser: Browser,
  link: string,
  heading: string,
) {
  const page = await browser.newPage();
  await page.goto(link);
  await expect(page.getByRole("heading", { name: heading })).toBeVisible();
  await page.close();
}

async function signInAndOpen(page: Page, username: string, path: string) {
  await signIn(page, username, password);
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
  await page.goto(path);
}

function findRow(page: Page, text: string) {
  return page.getByRole("row").filter({ hasText: text });
}

async function confirmDialog(page: Page, dialog: string, button: string) {
  await page
    .getByRole("dialog", { name: dialog })
    .getByRole("button", { name: button })
    .click();
}

test.describe("invitations", () => {
  test("a used invitation link stops working", async ({ browser }) => {
    test.skip(isEnterprise, directMembershipReason);
    const { owner, tenant } = await createTeam();
    const invitee = createUser("invitee");
    const { link } = await invite(owner.token, tenant, invitee.email);

    const { token } = await loginAs(invitee.username, password);
    await acceptInvite({ ...buildRequestContext({ token }), path: { tenant } });

    await expectInvitationPage(browser, link, "Invitation Unavailable");
  });

  test("cancelling an invitation kills its link", async ({ page, browser }) => {
    const { owner, tenant } = await createTeam();
    const email = buildRandomEmail("invitee");
    const { link } = await invite(owner.token, tenant, email);
    await expectInvitationPage(browser, link, "You've been invited");

    await signInAndOpen(page, owner.username, "/team");
    const row = findRow(page, email);
    await row.getByRole("button", { name: "Cancel invitation" }).click();
    await confirmDialog(page, "Cancel Invitation", "Cancel Invitation");
    await expect(row).toHaveCount(0);

    await expectInvitationPage(browser, link, "Invitation Unavailable");
  });

  test("an expired link dies, and regenerating issues a working one", async ({
    page,
    browser,
  }) => {
    const { owner, tenant } = await createTeam();
    const email = buildRandomEmail("invitee");
    const { link: expiredLink } = await invite(owner.token, tenant, email);
    expireInvitation(tenant);
    await expectInvitationPage(browser, expiredLink, "Invitation Unavailable");

    await signInAndOpen(page, owner.username, "/team");
    const row = findRow(page, email);
    await expect(row).toContainText("expired");
    await row
      .getByRole("button", { name: "Regenerate invitation link" })
      .click();
    await confirmDialog(page, "Regenerate Link", "Regenerate");
    await expect(row).toContainText("expires");

    const { data: invitations } = await getNamespaceMembershipInvitationList({
      ...buildRequestContext({ token: owner.token }),
      path: { tenant },
    });
    const regenerated = invitations.find((i) => i.user.email === email);
    if (!regenerated?.invite_url) {
      throw new Error(`expected a regenerated invite_url for ${email}`);
    }
    await expectInvitationPage(
      browser,
      regenerated.invite_url,
      "You've been invited",
    );
    await expectInvitationPage(browser, expiredLink, "Invitation Unavailable");
  });
});

test.describe("roles", () => {
  test("a role change applies to the member's existing token and API keys", async ({
    page,
  }) => {
    const { owner, member } = await createTeamWithMemberKey();
    await expectMemberStatus(listAccessPolicies, member, 200);

    await signInAndOpen(page, owner.username, "/team");
    const row = findRow(page, member.email);
    await row.getByRole("button", { name: "Edit role" }).click();
    const drawer = page.getByRole("dialog", { name: "Edit Role" });
    await drawer.getByRole("radio", { name: "Observer" }).press("Space");
    await drawer.getByRole("button", { name: "Save role" }).click();
    await expect(row).toContainText("observer");

    await expectMemberStatus(listAccessPolicies, member, 403);
  });
});

test.describe("losing membership", () => {
  test("removing a member ends their token and revokes their API keys", async ({
    page,
  }) => {
    const team = await createTeamWithMemberKey();
    const { owner, member, tenant } = team;
    await expectMemberStatus(listMembers(tenant), member, 200);

    await signInAndOpen(page, owner.username, "/team");
    const row = findRow(page, member.email);
    await row.getByRole("button", { name: "Remove member" }).click();
    await confirmDialog(page, "Remove Member", "Remove");
    await expect(row).toHaveCount(0);

    await expectMemberStatus(listMembers(tenant), member, 401);
    await rejoinAndExpectKeyRevoked(team);
  });

  test("leaving ends the session and the membership, and revokes API keys", async ({
    page,
  }) => {
    const team = await createTeamWithMemberKey();
    const { owner, member, tenant } = team;
    await expectMemberStatus(listMembers(tenant), member, 200);

    await signInAndOpen(page, member.username, "/settings");
    await page.getByRole("button", { name: "Leave", exact: true }).click();
    await confirmDialog(page, "Leave Namespace", "Leave");
    await expect(page).toHaveURL(/\/login$/);

    await expectMemberStatus(listMembers(tenant), member, 401);

    const { data: members } = await listNamespaceMembers({
      ...buildRequestContext({ token: owner.token }),
      path: { tenant },
    });
    expect(members.map((m) => m.username)).toEqual([owner.username]);

    await rejoinAndExpectKeyRevoked(team);
  });
});

test.describe("namespace switching", () => {
  test.skip(isCommunity, "community binds the instance to a single namespace");

  test("a member of two teams switches between them", async ({ page }) => {
    const { namespace, member } = await createTeamWithMember("observer");
    const otherTeam = await createTeam();
    addMember(member.username, otherTeam.namespace, "observer");

    await signInAndOpen(page, member.username, "/dashboard");
    const activeTab = page.getByRole("tab", { selected: true });
    await expect(activeTab).toHaveText(new RegExp(`${namespace}$`));

    await page
      .getByRole("button", { name: "Open a device, session or namespace" })
      .click();
    const palette = page.getByRole("dialog", { name: "Command palette" });
    await palette.getByRole("combobox").fill(otherTeam.namespace);
    await palette.getByRole("option", { name: otherTeam.namespace }).click();
    await dismissWizard(page);

    await expect(activeTab).toHaveText(new RegExp(`${otherTeam.namespace}$`));
  });
});
