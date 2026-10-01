import { expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { listNamespaceMembers, removeNamespaceMember } from "@/client";
import { adminUser, isEnterprise } from "./env";
import {
  signIn,
  fillLoginForm,
  dismissWizard,
  directMembershipReason,
  signUpFromInvite,
  createTeamWithMember,
} from "./helpers";
import { password, buildShortId, createUser, createNamespace } from "./seed";
import { buildRequestContext, expectStatus, invite, loginAs } from "./api";

test.describe("authentication", () => {
  test("signs in with valid credentials", async ({ page }) => {
    await signIn(page, adminUser.username, adminUser.password);

    await expect(page).toHaveURL(/\/dashboard$/);
  });

  test("rejects invalid credentials", async ({ page }) => {
    await signIn(page, "e2e-nobody", "wrong-password");

    await expect(page.getByText("Invalid login credentials")).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);
  });

  test("signs out", async ({ page }) => {
    await signIn(page, adminUser.username, adminUser.password);
    await expect(page).toHaveURL(/\/dashboard$/);

    await dismissWizard(page);
    await page
      .getByRole("button", { name: `Account menu for ${adminUser.username}` })
      .click();
    await page.getByRole("button", { name: "Logout" }).click();

    await expect(page).toHaveURL(/\/login$/);

    await page.goto("/dashboard");
    await expect(page).toHaveURL(/\/login$/);
  });
});

test.describe("session ended on the server", () => {
  test("a removed member is logged out on the next request", async ({
    page,
  }) => {
    const { owner, member, tenant } = await createTeamWithMember("observer");
    await signIn(page, member.username, password);
    await expect(page).toHaveURL(/\/dashboard$/);
    await dismissWizard(page);

    await removeNamespaceMember({
      ...buildRequestContext({ token: owner.token }),
      path: { tenant, uid: member.id },
    });
    const requestRejected = page.waitForResponse(
      (r) => r.url().includes("/api/") && r.status() === 401,
    );
    await page.getByRole("link", { name: "Devices", exact: true }).click();

    await requestRejected;
    await expect(page).toHaveURL(/\/login$/);
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/\/login$/);
    await expectStatus(
      (opts) => listNamespaceMembers({ ...opts, path: { tenant } }),
      { token: member.token },
      401,
    );
  });
});

test.describe("account lockout", () => {
  test("locks out after 3 failed attempts, then recovers", async ({ page }) => {
    test.setTimeout(120_000);

    const user = createUser("lockout");
    createNamespace(
      user.username,
      `ns-lockout-${buildShortId()}`,
      randomUUID(),
    );

    await page.goto("/login");
    for (let i = 0; i < 3; i++) {
      const resp = page.waitForResponse((r) => r.url().includes("/api/login"));
      await fillLoginForm(page, user.username, "wrong-password");
      await resp;
    }

    await expect(
      page.getByText("Too many failed login attempts"),
    ).toBeVisible();
    await expect(page.getByText(/\d+ (second|minute)/)).toBeVisible();

    await expect(page.getByText("Your timeout has finished")).toBeVisible({
      timeout: 75000,
    });

    await fillLoginForm(page, user.username, password);

    await expect(page).toHaveURL(/\/dashboard$/);
  });
});

test.describe("accept invitation", () => {
  let adminToken: string;
  let adminTenant: string;

  test.beforeAll(async () => {
    const auth = await loginAs(adminUser.username, adminUser.password);
    if (!auth.tenant) throw new Error(`${adminUser.username} has no namespace`);
    adminToken = auth.token;
    adminTenant = auth.tenant;
  });

  async function inviteUser() {
    const user = createUser("invitee");
    createNamespace(
      user.username,
      `ns-invitee-${buildShortId()}`,
      randomUUID(),
    );

    const { link } = await invite(adminToken, adminTenant, user.email);

    return { user, link };
  }

  test("existing user, logged in — accepts and switches namespace", async ({
    page,
  }) => {
    test.skip(isEnterprise, directMembershipReason);
    const { user, link } = await inviteUser();

    await signIn(page, user.username, password);
    await dismissWizard(page);

    await page.goto(link);

    await page.getByRole("button", { name: "Accept" }).click();

    const dialog = page.getByRole("dialog", { name: "Accept Invitation" });
    await dialog.getByRole("button", { name: "Accept" }).click();

    await page.getByRole("button", { name: "Go to Dashboard" }).click();
    await expect(page).toHaveURL(/\/dashboard$/, { timeout: 15000 });
    await expect(page.getByRole("tab", { selected: true })).toHaveAttribute(
      "title",
      adminUser.namespace,
    );
  });

  test("not logged in — redirected to login, then back", async ({ page }) => {
    test.skip(isEnterprise, directMembershipReason);
    const { user, link } = await inviteUser();

    await page.goto(link);
    await expect(page).toHaveURL(/\/login.*redirect/, { timeout: 10000 });

    await fillLoginForm(page, user.username, password);

    await expect(page).toHaveURL(/\/accept-invite/, { timeout: 10000 });
    await expect(page.getByRole("button", { name: "Accept" })).toBeVisible();
  });

  test("new user — sign-up form, create account + join", async ({ page }) => {
    const id = buildShortId();
    const { link } = await invite(
      adminToken,
      adminTenant,
      `e2e-signup-${id}@e2e.test`,
    );

    await signUpFromInvite(page, link, `e2e-signup-${id}`);

    await expect(page.getByText(/you.re in/i)).toBeVisible({ timeout: 15000 });
  });
});
