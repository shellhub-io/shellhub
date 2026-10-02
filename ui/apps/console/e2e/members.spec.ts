import {
  type Browser,
  type Locator,
  type Page,
  expect,
  test,
} from "@playwright/test";
import { generateKeyPairSync, randomBytes } from "node:crypto";
import {
  getNamespaceMembershipInvitationList,
  acceptInvite,
  acceptDevicePairing,
  authDevice,
  createDevicePairing,
  getDevices,
  listAccessPolicies,
  listNamespaceMembers,
} from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import { isCloud, isCommunity, isEnterprise } from "./env";
import {
  createTeam,
  consoleAccountDeletionReason,
  createTeamWithMember,
  deleteOwnAccount,
  findRow,
  signIn,
  signInAndOpen,
  directMembershipReason,
  signUpFromInvite,
  singleNamespaceReason,
  switchNamespace,
} from "./helpers";
import {
  password,
  buildShortId,
  buildRandomEmail,
  createUser,
  addMember,
  expireInvitation,
  readUserInvitationStatus,
} from "./seed";
import {
  type Endpoint,
  buildRequestContext,
  createApiKey,
  expectStatus,
  invite,
  loginAs,
} from "./api";

async function createTeamWithMemberKey() {
  const team = await createTeamWithMember("administrator");
  const { key } = await createApiKey(team.member.token);
  return { ...team, member: { ...team.member, apiKey: key } };
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

async function confirmDialog(page: Page, dialog: string, button: string) {
  await page
    .getByRole("dialog", { name: dialog })
    .getByRole("button", { name: button })
    .click();
}

const pendingApprovalReason =
  "only enterprise holds the invitee of a non-admin for an instance admin's approval";

test.describe("invitations", () => {
  test("signing up through an invitation consumes the user invitation", async ({
    page,
    browser,
  }) => {
    test.skip(isEnterprise, pendingApprovalReason);
    const { owner, tenant } = await createTeam();
    const email = buildRandomEmail("invitee");
    const { link } = await invite(owner.token, tenant, email);
    expect(readUserInvitationStatus(email)).toBe("pending");

    await signUpFromInvite(page, link, `e2e-invitee-${buildShortId()}`);

    await expect(page.getByRole("heading", { name: "You're in" })).toBeVisible({
      timeout: 15000,
    });
    await expectInvitationPage(browser, link, "Invitation Unavailable");
    expect(readUserInvitationStatus(email)).toBe("accepted");
  });

  test("a non-admin's invitee signs in only after an instance admin approves", async ({
    page,
    browser,
  }) => {
    test.skip(!isEnterprise, pendingApprovalReason);
    const { owner, tenant } = await createTeam();
    const email = buildRandomEmail("invitee");
    const username = `e2e-invitee-${buildShortId()}`;
    const { link } = await invite(owner.token, tenant, email);

    await signUpFromInvite(page, link, username);
    await expect(
      page.getByRole("heading", { name: "Waiting for Approval" }),
    ).toBeVisible({ timeout: 15000 });
    expect(readUserInvitationStatus(email)).toBe("accepted");
    await expectInvitationPage(browser, link, "Invitation Unavailable");
    await signIn(page, username, password);
    await expect(
      page.getByText("Your account is waiting for an administrator"),
    ).toBeVisible();

    const { owner: admin } = await createTeam({ admin: true });
    await signInAndOpen(page, admin.username, "/admin/users");
    await page.getByLabel("Search users by username").fill(username);
    const row = findRow(page, email);
    await expect(row).toContainText("Awaiting Approval");
    await row
      .getByRole("button", { name: `Approve account for ${email}` })
      .click();
    await confirmDialog(page, "Approve account", "Approve account");
    await expect(row).toContainText("Confirmed");

    const { tenant: joined } = await loginAs(username, password);
    expect(joined).toBe(tenant);
  });

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

async function keepAsTeamDevices(dialog: Locator, names: string[]) {
  for (const name of names) {
    const keep = dialog.getByRole("checkbox", {
      name: `Keep ${name} as a team device`,
    });
    await keep.press("Space");
    await expect(keep).toBeChecked();
  }
}

async function changeRole(
  page: Page,
  owner: { username: string },
  member: { email: string },
  role: AssignableRole,
  keep: string[] = [],
) {
  await signInAndOpen(page, owner.username, "/team");
  const row = findRow(page, member.email);
  await row.getByRole("button", { name: "Edit role" }).click();
  const drawer = page.getByRole("dialog", { name: "Edit Role" });
  await drawer.getByRole("radio", { name: role }).press("Space");
  await keepAsTeamDevices(drawer, keep);
  await drawer.getByRole("button", { name: "Save role" }).click();
  await expect(row).toContainText(role);
}

async function removeMember(
  page: Page,
  owner: { username: string },
  member: { email: string },
  keep: string[] = [],
) {
  await signInAndOpen(page, owner.username, "/team");
  const row = findRow(page, member.email);
  await row.getByRole("button", { name: "Remove member" }).click();
  await keepAsTeamDevices(
    page.getByRole("dialog", { name: "Remove Member" }),
    keep,
  );
  await confirmDialog(page, "Remove Member", "Remove");
  await expect(row).toHaveCount(0);
}

async function leaveNamespace(page: Page, member: { username: string }) {
  await signInAndOpen(page, member.username, "/settings");
  await page.getByRole("button", { name: "Leave", exact: true }).click();
  await confirmDialog(page, "Leave Namespace", "Leave");
  await expect(page).toHaveURL(/\/login$/);
}

test.describe("roles", () => {
  test("demoting an administrator to operator restricts their token but not their API keys", async ({
    page,
  }) => {
    const { owner, member } = await createTeamWithMemberKey();
    await expectMemberStatus(listAccessPolicies, member, 200);

    await changeRole(page, owner, member, "operator");

    await expectStatus(listAccessPolicies, { token: member.token }, 403);
    await expectStatus(listAccessPolicies, { apiKey: member.apiKey }, 200);
  });

  test("demoting to observer revokes the member's API keys", async ({
    page,
  }) => {
    const { owner, member } = await createTeamWithMemberKey();
    await expectMemberStatus(listAccessPolicies, member, 200);

    await changeRole(page, owner, member, "observer");

    await expectStatus(listAccessPolicies, { token: member.token }, 403);
    await expectStatus(listAccessPolicies, { apiKey: member.apiKey }, 401);
  });
});

test.describe("losing membership", () => {
  test("removing a member ends their token and revokes their API keys", async ({
    page,
  }) => {
    const team = await createTeamWithMemberKey();
    const { owner, member, tenant } = team;
    await expectMemberStatus(listMembers(tenant), member, 200);

    await removeMember(page, owner, member);

    await expectMemberStatus(listMembers(tenant), member, 401);
    await rejoinAndExpectKeyRevoked(team);
  });

  test("leaving ends the session and the membership, and revokes API keys", async ({
    page,
  }) => {
    const team = await createTeamWithMemberKey();
    const { owner, member, tenant } = team;
    await expectMemberStatus(listMembers(tenant), member, 200);

    await leaveNamespace(page, member);

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
  test.skip(isCommunity, singleNamespaceReason);

  test("a member of two teams switches between them", async ({ page }) => {
    const { namespace, member } = await createTeamWithMember("observer");
    const otherTeam = await createTeam();
    addMember(member.username, otherTeam.namespace, "observer");

    await signInAndOpen(page, member.username, "/dashboard");
    const activeTab = page.getByRole("tab", { selected: true });
    await expect(activeTab).toHaveText(new RegExp(`${namespace}$`));

    await switchNamespace(page, otherTeam.namespace);
  });
});

function buildPairingRequest() {
  const { publicKey } = generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "pem" },
  });
  return {
    hostname: `e2e-device-${buildShortId()}`,
    identity: {
      mac: [0x02, ...randomBytes(5)]
        .map((byte) => byte.toString(16).padStart(2, "0"))
        .join(":"),
    },
    info: {
      id: "debian",
      pretty_name: "Debian GNU/Linux 12",
      version: "12",
      arch: "x86_64",
      platform: "native" as const,
    },
    public_key: publicKey,
  };
}

async function requestPairing() {
  const body = buildPairingRequest();
  const { data } = await createDevicePairing({
    ...buildRequestContext(),
    body,
  });
  if (!data.code) {
    throw new Error(`expected a pairing code for ${body.hostname}`);
  }
  return { code: data.code, name: body.hostname, body };
}

type PairedDevice = ReturnType<typeof buildPairingRequest> & {
  tenant_id: string;
};

async function pairDevice(
  member: { token: string },
  tenant: string,
): Promise<PairedDevice> {
  const { code, name, body } = await requestPairing();
  const { data } = await acceptDevicePairing({
    ...buildRequestContext({ token: member.token }),
    path: { code },
    body: { tenant_id: tenant },
  });
  if (!data.uid) throw new Error(`expected ${name} to join ${tenant}`);
  return { ...body, tenant_id: tenant };
}

async function authenticateAgent(device: PairedDevice) {
  const { response } = await authDevice({
    ...buildRequestContext(),
    throwOnError: false,
    body: device,
  });
  return response?.status;
}

async function readAcceptedDevices(owner: { token: string }) {
  const { data } = await getDevices({
    ...buildRequestContext({ token: owner.token }),
    query: { status: "accepted" },
  });
  return data.map(({ name, owner_id }) => ({ name, owner_id }));
}

test.describe("paired devices", () => {
  test("accepting a pairing makes the accepting member the owner", async ({
    page,
  }) => {
    const { owner, member, namespace } = await createTeamWithMember("operator");
    const { code, name } = await requestPairing();

    await signInAndOpen(page, member.username, `/accept-device?code=${code}`);
    await expect(
      page.getByRole("region", { name: "Accepting as" }),
    ).toContainText(member.email);
    await page.getByRole("button", { name: "Accept device" }).click();
    await expect(
      page.getByRole("heading", { name: "Device accepted" }),
    ).toBeVisible();
    await expect(page.getByText(`It's in ${namespace} now`)).toBeVisible();
    await page.getByRole("button", { name: "View device" }).click();

    await expect(page.getByRole("heading", { name })).toBeVisible();
    await expect(page.getByLabel("Paired by", { exact: true })).toContainText(
      member.email,
    );
    expect(await readAcceptedDevices(owner)).toStrictEqual([
      { name, owner_id: member.id },
    ]);
  });

  test("demoting to observer keeps the ticked device for the team and removes the rest", async ({
    page,
  }) => {
    const { owner, member, tenant } = await createTeamWithMember("operator");
    const kept = await pairDevice(member, tenant);
    const removed = await pairDevice(member, tenant);

    await changeRole(page, owner, member, "observer", [kept.hostname]);

    expect(await readAcceptedDevices(owner)).toStrictEqual([
      { name: kept.hostname, owner_id: undefined },
    ]);
    expect(await authenticateAgent(kept)).toBe(200);
    expect(await authenticateAgent(removed)).toBe(401);
  });

  test("removing a member keeps the ticked device for the team and removes the rest", async ({
    page,
  }) => {
    const { owner, member, tenant } = await createTeamWithMember("operator");
    const kept = await pairDevice(member, tenant);
    const removed = await pairDevice(member, tenant);

    await removeMember(page, owner, member, [kept.hostname]);

    expect(await readAcceptedDevices(owner)).toStrictEqual([
      { name: kept.hostname, owner_id: undefined },
    ]);
    expect(await authenticateAgent(kept)).toBe(200);
    expect(await authenticateAgent(removed)).toBe(401);
  });

  test("leaving removes the member's paired devices", async ({ page }) => {
    const { owner, member, tenant } = await createTeamWithMember("operator");
    const removed = await pairDevice(member, tenant);

    await leaveNamespace(page, member);

    expect(await readAcceptedDevices(owner)).toStrictEqual([]);
    expect(await authenticateAgent(removed)).toBe(401);
  });

  test("deleting the account removes the member's paired devices", async ({
    page,
  }) => {
    test.skip(!isCloud, consoleAccountDeletionReason);
    const { owner, member, tenant } = await createTeamWithMember("operator");
    const removed = await pairDevice(member, tenant);

    await deleteOwnAccount(page, member.username);

    expect(await readAcceptedDevices(owner)).toStrictEqual([]);
    expect(await authenticateAgent(removed)).toBe(401);
  });
});
