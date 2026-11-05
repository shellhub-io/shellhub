import { type Locator, type Page, expect } from "@playwright/test";
import { randomUUID } from "node:crypto";
import {
  confirmAccount as confirmAccountSdk,
  registerUser,
  provisioningKeyCreate,
} from "@/client";
import type { AssignableRole } from "@/pages/team/helpers";
import {
  authSessionKey,
  authSessionVersion,
  signedInSession,
} from "@/stores/authSession";
import { readLatestEmail } from "./mail";
import {
  password,
  buildShortId,
  buildUserIdentity,
  createUser,
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

export async function readConfirmationCode(address: string) {
  const { body, link } = await readLatestEmail(address, "/confirm-account");
  const shown = body
    .querySelector('[aria-label="Confirmation code"]')
    ?.textContent?.trim();
  return {
    code: required(shown, "a confirmation code in the email"),
    link: link.path,
  };
}

export async function confirmAccount(address: string) {
  const { code } = await readConfirmationCode(address);
  await confirmAccountSdk({
    ...buildRequestContext(),
    body: { identifier: address, code },
  });
}

export async function submitSignUpForm(
  page: Page,
  { username, email }: { username: string; email: string },
) {
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
  await expect(page).toHaveURL(/\/confirm-account\?identifier=/);
  await expect(
    page.getByRole("heading", { name: "Confirm your email" }),
  ).toBeVisible();
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

export async function createManualProvisioningKey(token: string) {
  const { data } = await provisioningKeyCreate({
    ...buildRequestContext({ token }),
    body: { name: `e2e-${buildShortId()}`, mode: "manual" },
  });
  if (!data.key) throw new Error("expected a provisioning key secret");
  return data.key;
}

export async function createTeam({
  admin = false,
  legacy = false,
}: { admin?: boolean } & NamespaceOptions = {}) {
  const owner = createUser("owner", { admin });
  const namespace = `e2e-team-${buildShortId()}`;
  const tenant = randomUUID();
  createNamespace(owner.username, namespace, tenant, { legacy });
  const { token } = await loginAs(owner.username, password);
  const provisioningKey = legacy
    ? undefined
    : await createManualProvisioningKey(token);
  return { owner: { ...owner, token }, namespace, tenant, provisioningKey };
}

export async function createTeamWithMember(
  role: AssignableRole,
  options: NamespaceOptions = {},
) {
  const team = await createTeam(options);
  const member = createUser("member");
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

export async function chooseRadio(scope: Page | Locator, name: string) {
  const radio = scope.getByRole("radio", { name });
  await radio.locator("xpath=ancestor::label[1]").click();
  await expect(radio).toBeChecked();
}

export function findRow(page: Page, text: string) {
  return page.getByRole("row").filter({ hasText: text });
}

export async function generateKey(
  page: Page,
  {
    dialog,
    submit,
    keyLabel,
  }: { dialog: string; submit: string; keyLabel: string },
) {
  const name = `e2e-${buildShortId()}`;
  await page.getByRole("button", { name: "Generate Key" }).click();
  const generate = page.getByRole("dialog", { name: dialog });
  await generate.getByLabel("Name").fill(name);
  await generate.getByRole("button", { name: submit, exact: true }).click();
  const plaintext = generate.getByLabel(keyLabel, { exact: true });
  await expect(plaintext).toHaveText(/\S+/);
  const key = await plaintext.textContent();
  if (!key) throw new Error(`expected the plaintext of ${name}`);
  await generate.getByRole("button", { name: "Done" }).click();
  await expect(findRow(page, name)).toBeVisible();
  return { name, key };
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
  const session = await loginAs(username, password);
  await page.addInitScript(
    ({ key, value, seeded }) => {
      if (window !== window.top || sessionStorage.getItem(seeded)) return;
      sessionStorage.setItem(seeded, "1");
      localStorage.setItem(key, value);
    },
    {
      key: authSessionKey,
      seeded: `e2e-session-${randomUUID()}`,
      value: JSON.stringify({
        state: signedInSession(session),
        version: authSessionVersion,
      }),
    },
  );
  await page.goto(path);
}

export async function signOut(page: Page, username: string) {
  await page
    .getByRole("button", { name: `Account menu for ${username}` })
    .click();
  await page.getByRole("button", { name: "Logout" }).click();
  await expect(page).toHaveURL(/\/login$/);
}

/**
 * The selected tab in the console's open-views strip, which names the current namespace. Scoped
 * to that strip because pages such as team render tablists of their own.
 */
export function activeViewTab(page: Page) {
  return page
    .getByRole("tablist", { name: "Open views" })
    .getByRole("tab", { selected: true });
}

/**
 * Holds every request the page makes to url until release is called. requested waits until one
 * reaches the hold, and release lets them through and waits until one was served. Both fail the
 * test naming url when no request gets there, so a matcher that stops matching breaks the test
 * instead of skipping what it holds back.
 */
export async function holdRoute(
  page: Page,
  url: string | RegExp | ((url: URL) => boolean),
) {
  let open = () => {};
  const opened = new Promise<void>((resolve) => (open = resolve));
  let reached = false;
  let served = false;
  await page.route(url, async (route) => {
    reached = true;
    await opened;
    await route.continue();
    served = true;
  });
  const matcher =
    typeof url === "function" ? url.name || url.toString() : url.toString();
  const waitFor = (done: () => boolean, what: string) =>
    expect
      .poll(done, { message: `no request to ${matcher} was ${what}` })
      .toBe(true);
  return {
    requested: () => waitFor(() => reached, "held"),
    release: async () => {
      open();
      await waitFor(() => served, "served");
    },
  };
}

export async function switchNamespace(page: Page, namespace: string) {
  await page.getByRole("button", { name: "New tab" }).click();
  const picker = page.getByRole("dialog", { name: "New tab" });
  await picker.getByRole("combobox").fill(namespace);
  await picker.getByRole("option", { name: namespace }).click();
  await expect(activeViewTab(page)).toHaveText(new RegExp(`${namespace}$`));
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
