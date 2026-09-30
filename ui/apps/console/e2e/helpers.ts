import { type Page, expect } from "@playwright/test";
import { randomUUID } from "node:crypto";
import type { AssignableRole } from "@/pages/team/helpers";
import {
  password,
  buildShortId,
  createUser,
  createNamespace,
  addMember,
} from "./seed";
import { loginAs } from "./api";

export const directMembershipReason =
  "enterprise adds existing users directly, without an invitation link";

export async function createTeam({ admin = false } = {}) {
  const owner = createUser("owner", { admin });
  const namespace = `e2e-team-${buildShortId()}`;
  const tenant = randomUUID();
  createNamespace(owner.username, namespace, tenant);
  const { token } = await loginAs(owner.username, password);
  return { owner: { ...owner, token }, namespace, tenant };
}

export async function createTeamWithMember(role: AssignableRole) {
  const team = await createTeam();
  const member = createUser("member");
  addMember(member.username, team.namespace, role);
  const { token, tenant, id } = await loginAs(member.username, password);
  if (tenant !== team.tenant) {
    throw new Error(`${member.username} did not land in ${team.namespace}`);
  }
  return { ...team, member: { ...member, token, id } };
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
