import { type Page, expect } from "@playwright/test";
import { randomUUID } from "node:crypto";
import { getValidateAccount, registerUser } from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import { isCloud } from "./env";
import { readLatestEmail } from "./mail";
import {
  password,
  buildShortId,
  buildUserIdentity,
  createUserWithCli,
  createNamespace,
  addMember,
  type NamespaceOptions,
} from "./seed";
import { buildRequestContext, loginAs } from "./api";

export const directMembershipReason =
  "enterprise adds existing users directly, without an invitation link";

export const singleNamespaceReason =
  "community binds the instance to a single namespace";

export const emailDeliveryReason = "only the cloud sends email";
export const mfaReason = "MFA exists only in enterprise and cloud";

export const consoleAccountDeletionReason =
  "only the cloud deletes an account from the console";

export function required(value: string | null | undefined, what: string) {
  if (!value) throw new Error(`expected ${what}`);
  return value;
}

export async function confirmAccount(address: string) {
  const { link } = await readLatestEmail(address, "/validation-account");
  const email = link.params.get("email");
  const token = link.params.get("token");
  if (!email || !token) {
    throw new Error(`expected email and token params in ${link.path}`);
  }
  await getValidateAccount({
    ...buildRequestContext(),
    query: { email, token },
  });
}

export async function signUpUser(prefix: string, { confirm = true } = {}) {
  const user = buildUserIdentity(prefix);
  await registerUser({
    ...buildRequestContext(),
    body: { ...user, name: user.username, password, email_marketing: false },
  });
  if (confirm) await confirmAccount(user.email);
  return user;
}

export async function createUser(prefix: string, { admin = false } = {}) {
  if (!isCloud || admin) return createUserWithCli(prefix, { admin });
  return signUpUser(prefix);
}

export async function createTeam({
  admin = false,
  sshAccessMode,
}: { admin?: boolean } & NamespaceOptions = {}) {
  const owner = await createUser("owner", { admin });
  const namespace = `e2e-team-${buildShortId()}`;
  const tenant = randomUUID();
  createNamespace(owner.username, namespace, tenant, { sshAccessMode });
  const { token } = await loginAs(owner.username, password);
  return { owner: { ...owner, token }, namespace, tenant };
}

export async function createTeamWithMember(
  role: AssignableRole,
  options: NamespaceOptions = {},
) {
  const team = await createTeam(options);
  const member = await createUser("member");
  addMember(member.username, team.namespace, role);
  const { token, tenant, id } = await loginAs(member.username, password);
  if (tenant !== team.tenant) {
    throw new Error(`${member.username} did not land in ${team.namespace}`);
  }
  return { ...team, member: { ...member, token, id } };
}

export async function createTeamWithOwnerInAnother(
  options: NamespaceOptions = {},
) {
  const team = await createTeam(options);
  const other = await createTeam(options);
  addMember(team.owner.username, other.namespace, "administrator");
  return { ...team, other };
}

export function findRow(page: Page, text: string) {
  return page.getByRole("row").filter({ hasText: text });
}

export async function fillLoginForm(
  page: Page,
  username: string,
  password: string,
) {
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();
}

export async function signIn(page: Page, username: string, password: string) {
  await page.goto("/login");
  await fillLoginForm(page, username, password);
}

export async function signInAndOpen(
  page: Page,
  username: string,
  path: string,
) {
  await signIn(page, username, password);
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
  await page.goto(path);
}

export async function dismissWizard(page: Page) {
  const close = page.getByRole("button", { name: "Close wizard" });
  if (
    await close.waitFor({ state: "visible", timeout: 5000 }).then(
      () => true,
      () => false,
    )
  ) {
    await close.click();
    await close.waitFor({ state: "hidden" });
  }
}

export async function signOut(page: Page, username: string) {
  await page
    .getByRole("button", { name: `Account menu for ${username}` })
    .click();
  await page.getByRole("button", { name: "Logout" }).click();
  await expect(page).toHaveURL(/\/login$/);
}

export async function switchNamespace(page: Page, namespace: string) {
  await page
    .getByRole("button", { name: "Open a device, session or namespace" })
    .click();
  const palette = page.getByRole("dialog", { name: "Command palette" });
  await palette.getByRole("combobox").fill(namespace);
  await palette.getByRole("option", { name: namespace }).click();
  await dismissWizard(page);
  await expect(page.getByRole("tab", { selected: true })).toHaveText(
    new RegExp(`${namespace}$`),
  );
}

export async function signUpFromInvite(
  page: Page,
  link: string,
  username: string,
) {
  await page.goto(link);
  await expect(page.getByRole("heading", { name: /invited/i })).toBeVisible({
    timeout: 10000,
  });
  await page.getByLabel("Name", { exact: true }).fill("E2E Signup");
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Confirm password").fill(password);
  await page.getByRole("button", { name: "Join Namespace" }).click();
}

export async function deleteOwnAccount(page: Page, username: string) {
  await signInAndOpen(page, username, "/account/danger-zone");
  await page.getByRole("button", { name: "Delete account" }).click();
  await page
    .getByRole("dialog", { name: "Delete account" })
    .getByRole("button", { name: "Delete account" })
    .click();
  await expect(page).toHaveURL(/\/login$/);
}
