import { type Page, expect, test } from "@playwright/test";
import { getMembershipInvitationList, getUserInfo } from "@/client";
import { isCloud } from "./env";
import {
  confirmAccount,
  consoleAccountDeletionReason,
  createTeam,
  createTeamWithMember,
  createUser,
  deleteOwnAccount,
  emailDeliveryReason,
  fillLoginForm,
  signInAndOpen,
  signUpUser,
} from "./helpers";
import { buildUserIdentity, password } from "./seed";
import {
  buildRequestContext,
  expectLoginStatus,
  expectStatus,
  invite,
  loginAs,
} from "./api";
import { readEmailLink } from "./mail";

const openSignUpReason = "only the cloud has open sign-up";

const newPassword = `${password}-new`;

async function signUp(
  page: Page,
  { username, email }: { username: string; email: string },
) {
  await page.goto("/sign-up");
  await page.getByLabel("Name", { exact: true }).fill("E2E Signup");
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Email", { exact: true }).fill(email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Confirm Password", { exact: true }).fill(password);
  const privacy = page.getByRole("checkbox", {
    name: "I agree to the Privacy Policy.",
  });
  await privacy.press("Space");
  await expect(privacy).toBeChecked();
  await page.getByRole("button", { name: "Create Account" }).click();
  await expect(page).toHaveURL(/\/confirm-account\?username=/);
  await expect(
    page.getByRole("heading", { name: "Account Activation Required" }),
  ).toBeVisible();
}

test.describe("registration", () => {
  test.skip(!isCloud, openSignUpReason);

  test("signing up leaves the account unconfirmed", async ({ page }) => {
    const account = buildUserIdentity("signup");

    await signUp(page, account);

    await expectLoginStatus(account.username, password, 403);
  });

  test("signing up with an invited email inherits the namespace invitation", async ({
    page,
  }) => {
    const { owner, tenant } = await createTeam();
    const account = buildUserIdentity("signup");
    await invite(owner.token, tenant, account.email);

    await signUp(page, account);

    await confirmAccount(account.email);
    const { token } = await loginAs(account.username, password);
    const { data: invitations } = await getMembershipInvitationList(
      buildRequestContext({ token }),
    );
    expect(invitations.map(({ namespace }) => namespace.tenant_id)).toEqual([
      tenant,
    ]);
  });
});

test.describe("email confirmation", () => {
  test.skip(!isCloud, emailDeliveryReason);

  test("the emailed link activates the account", async ({ page }) => {
    const account = await signUpUser("signup", { confirm: false });
    const link = await readEmailLink(account.email, "/validation-account");

    await page.goto(link);

    await expect(
      page.getByText("Your account has been activated successfully"),
    ).toBeVisible();
    await expectLoginStatus(account.username, password, 200);
  });

  test("a resent email replaces the first link", async ({ page }) => {
    const account = await signUpUser("signup", { confirm: false });
    const firstLink = await readEmailLink(account.email, "/validation-account");
    const readNewestLink = () =>
      readEmailLink(account.email, "/validation-account");

    await page.goto(
      `/confirm-account?username=${encodeURIComponent(account.username)}`,
    );
    await page.getByRole("button", { name: "Resend Email" }).click();
    await expect(
      page.getByText("Confirmation email sent successfully."),
    ).toBeVisible();
    await expect.poll(readNewestLink).not.toBe(firstLink);

    await page.goto(firstLink);
    await expect(
      page.getByText("Your account activation token has expired"),
    ).toBeVisible();
    await page.goto(await readNewestLink());
    await expect(
      page.getByText("Your account has been activated successfully"),
    ).toBeVisible();
    await expectLoginStatus(account.username, password, 200);
  });
});

test.describe("password", () => {
  test("changing the password replaces the old one", async ({ page }) => {
    const user = await createUser("password");
    const before = await loginAs(user.username, password);

    await signInAndOpen(page, user.username, "/account/security");
    await page
      .getByRole("button", { name: "Change Password", exact: true })
      .click();
    const dialog = page.getByRole("dialog", { name: "Change password" });
    await dialog.getByLabel("Current Password").fill(password);
    await dialog.getByLabel("New Password", { exact: true }).fill(newPassword);
    await dialog.getByLabel("Confirm New Password").fill(newPassword);
    await dialog.getByRole("button", { name: "Change password" }).click();
    await expect(page).toHaveURL(/\/login$/);
    await expect(
      page.getByText("Password changed. Sign in with your new password."),
    ).toBeVisible();

    await expectStatus(getUserInfo, { token: before.token }, 401);
    await expectLoginStatus(user.username, newPassword, 200);
    await expectLoginStatus(user.username, password, 401);
  });

  test("the emailed reset link sets a new password", async ({ page }) => {
    test.skip(!isCloud, emailDeliveryReason);
    const user = await createUser("forgot");
    const before = await loginAs(user.username, password);

    await page.goto("/login");
    await page.getByRole("link", { name: "Forgot password?" }).click();
    await page.getByLabel("Username or email address").fill(user.username);
    await page.getByRole("button", { name: "Reset Password" }).click();
    await expect(page.getByText("Check your inbox.")).toBeVisible();

    await page.goto(await readEmailLink(user.email, "/update-password"));
    await page.getByLabel("New Password").fill(newPassword);
    await page.getByLabel("Confirm Password").fill(newPassword);
    await page.getByRole("button", { name: "Update Password" }).click();
    await expect(
      page.getByText("Password updated successfully. Please sign in."),
    ).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);

    await fillLoginForm(page, user.username, newPassword);
    await expect(page).toHaveURL(/\/dashboard$/);
    await expectStatus(getUserInfo, { token: before.token }, 401);
    await expectLoginStatus(user.username, password, 401);
  });
});

test.describe("account deletion", () => {
  test.skip(!isCloud, consoleAccountDeletionReason);

  test("deleting the account removes the user", async ({ page }) => {
    const { member } = await createTeamWithMember("operator");

    await deleteOwnAccount(page, member.username);

    await expectLoginStatus(member.username, password, 401);
  });
});
