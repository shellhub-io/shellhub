import { type APIRequestContext, expect, test } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { adminUser, isEnterprise } from "./env";
import {
  signIn,
  fillLoginForm,
  dismissWizard,
  createUser,
  createNamespace,
} from "./helpers";
import { login, createInvitationLink } from "./api";

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

test.describe("account lockout", () => {
  test("locks out after 3 failed attempts, then recovers", async ({ page }) => {
    test.setTimeout(120_000);

    const id = randomUUID().slice(0, 8);
    const user = { username: `e2e-lockout-${id}`, password: "e2e-password" };
    createUser(user.username, user.password, `lockout-${id}@e2e.test`);
    createNamespace(user.username, `ns-lockout-${id}`, randomUUID());

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

    await fillLoginForm(page, user.username, user.password);

    await expect(page).toHaveURL(/\/dashboard$/);
  });
});

test.describe("accept invitation", () => {
  const directMembershipReason =
    "enterprise adds existing users directly, without an invitation link";
  let adminToken: string;
  let adminTenant: string;

  test.beforeAll(async ({ request }) => {
    const auth = await login(request, adminUser.username, adminUser.password);
    if (!auth.tenant) throw new Error(`${adminUser.username} has no namespace`);
    adminToken = auth.token;
    adminTenant = auth.tenant;
  });

  async function inviteUser(request: APIRequestContext) {
    const id = randomUUID().slice(0, 8);
    const user = {
      username: `e2e-invitee-${id}`,
      email: `invitee-${id}@e2e.test`,
      password: "e2e-password",
    };
    createUser(user.username, user.password, user.email);
    createNamespace(user.username, `ns-invitee-${id}`, randomUUID());

    const { link } = await createInvitationLink(
      request,
      adminToken,
      adminTenant,
      user.email,
    );

    if (!link) throw new Error("expected invitation link for existing user");

    return { user, sig: new URL(link).searchParams.get("invite")! };
  }

  test("existing user, logged in — accepts and switches namespace", async ({
    page,
    request,
  }) => {
    test.skip(isEnterprise, directMembershipReason);
    const { user, sig } = await inviteUser(request);

    await signIn(page, user.username, user.password);
    await dismissWizard(page);

    await page.goto(`/accept-invite?invite=${sig}`);

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

  test("not logged in — redirected to login, then back", async ({
    page,
    request,
  }) => {
    test.skip(isEnterprise, directMembershipReason);
    const { user, sig } = await inviteUser(request);

    await page.goto(`/accept-invite?invite=${sig}`);
    await expect(page).toHaveURL(/\/login.*redirect/, { timeout: 10000 });

    await fillLoginForm(page, user.username, user.password);

    await expect(page).toHaveURL(/\/accept-invite/, { timeout: 10000 });
    await expect(page.getByRole("button", { name: "Accept" })).toBeVisible();
  });

  test("new user — sign-up form, create account + join", async ({
    page,
    request,
  }) => {
    const id = randomUUID().slice(0, 8);
    const { link } = await createInvitationLink(
      request,
      adminToken,
      adminTenant,
      `e2e-signup-${id}@e2e.test`,
    );
    if (!link) throw new Error("expected invitation link for unknown user");
    const sig = new URL(link).searchParams.get("invite")!;

    await page.goto(`/accept-invite?invite=${sig}`);

    await expect(page.getByRole("heading", { name: /invited/i })).toBeVisible({
      timeout: 10000,
    });

    await page.getByLabel("Name", { exact: true }).fill("E2E Signup");
    await page.getByLabel("Username", { exact: true }).fill(`e2e-signup-${id}`);
    await page.getByLabel("Password", { exact: true }).fill("e2e-password");
    await page.getByLabel("Confirm password").fill("e2e-password");

    await page.getByRole("button", { name: "Join Namespace" }).click();

    await expect(page.getByText(/you.re in/i)).toBeVisible({ timeout: 15000 });
  });
});
