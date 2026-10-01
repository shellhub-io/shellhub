import { type Page, expect, test } from "@playwright/test";
import { createHmac, randomUUID } from "node:crypto";
import { getUserInfo, mfaRecover, requestResetMfa, updateUser } from "@/client";
import { isCloud, isCommunity } from "./env";
import {
  createTeam,
  dismissWizard,
  fillLoginForm,
  signIn,
  signInAndOpen,
} from "./helpers";
import {
  buildRandomEmail,
  enableMFA,
  composeExec,
  countRecoveryCodes,
  mfaSecret,
  password,
  composeLogs,
} from "./seed";
import { buildRequestContext } from "./api";

const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

function totp(secret: string, stepOffset = 0) {
  const bits = [...secret]
    .map((c) => base32Alphabet.indexOf(c).toString(2).padStart(5, "0"))
    .join("");
  const key = Buffer.from(
    (bits.match(/.{8}/g) ?? []).map((byte) => parseInt(byte, 2)),
  );
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(
    BigInt(Math.floor(Date.now() / 30_000) + stepOffset),
  );
  const hmac = createHmac("sha1", key).update(counter).digest();
  const offset = hmac[hmac.length - 1] & 0xf;
  const code = (hmac.readUInt32BE(offset) & 0x7fffffff) % 1_000_000;
  return code.toString().padStart(6, "0");
}

const buildRecoveryCode = () =>
  randomUUID().replaceAll("-", "").slice(0, 16).toUpperCase();

async function fillDigits(page: Page, code: string) {
  for (const [i, digit] of [...code].entries()) {
    await page
      .getByRole("group", { name: "Verification Code" })
      .getByRole("textbox")
      .nth(i)
      .fill(digit);
  }
}

async function createMFAUser({ recoveryCodes = [] as string[] } = {}) {
  const { owner } = await createTeam();
  const recoveryEmail = buildRandomEmail("recovery");
  enableMFA(owner.username, { recoveryEmail, recoveryCodes });
  return { ...owner, recoveryEmail };
}

async function reachCodePrompt(page: Page, username: string) {
  await signIn(page, username, password);
  await expect(page).toHaveURL(/\/mfa-login$/);
}

async function signInWithMFA(page: Page, username: string) {
  await reachCodePrompt(page, username);
  await fillDigits(page, totp(mfaSecret));
  await page.getByRole("button", { name: "Verify" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
}

async function openEnableMFA(page: Page, username: string) {
  await signInAndOpen(page, username, "/account/security");
  await page.getByRole("button", { name: "Enable MFA" }).click();
}

async function completeEnrollment(page: Page) {
  const saved = page.getByRole("checkbox", {
    name: "I have saved my recovery codes in a secure location",
  });
  await saved.press("Space");
  await expect(saved).toBeChecked();
  await page.getByRole("button", { name: "Next Step" }).click();
  await expect(page.getByRole("img", { name: "MFA QR code" })).toBeVisible();
  const secret = await page.getByLabel("Manual Entry Key").inputValue();
  await fillDigits(page, totp(secret));
  await page.getByRole("button", { name: "Verify & Enable" }).click();
  await expect(page.getByText("MFA Enabled Successfully!")).toBeVisible();
}

async function recoverWithCode(page: Page, username: string, code: string) {
  await reachCodePrompt(page, username);
  await page.getByRole("link", { name: "Use a recovery code" }).click();
  await page.getByLabel("Recovery Code").fill(code);
  await page.getByRole("button", { name: "Recover Account" }).click();
}

type MFAUser = Awaited<ReturnType<typeof createMFAUser>>;

async function readMFA({ token }: { token: string }) {
  const { data } = await getUserInfo(buildRequestContext({ token }));
  return { enabled: data.mfa, recoveryEmail: data.recovery_email };
}

async function setRecoveryEmail(token: string, recoveryEmail: string) {
  await updateUser({
    ...buildRequestContext({ token }),
    body: { recovery_email: recoveryEmail },
  });
}

async function readResetEmail(email: string) {
  const findLine = () =>
    composeLogs("server")
      .split("\n")
      .reverse()
      .find((l) => l.includes("[mail/dummy]") && l.includes(`to=${email}`));
  await expect
    .poll(findLine, { message: `MFA reset email to ${email}` })
    .toBeTruthy();
  const line = findLine() ?? "";
  const code = line.match(/following code\.(?:\\n|\s)*([A-Z2-7]{5})\b/)?.[1];
  const link = line.match(/https?:\/\/[^\s"\\)]+\/reset-mfa\?id=[\w-]+/)?.[0];
  if (!code || !link) {
    throw new Error(
      `expected a code and link in the email to ${email}: ${line}`,
    );
  }
  return { code, link };
}

async function readResetCodes(user: MFAUser) {
  const main = await readResetEmail(user.email);
  const recovery = await readResetEmail(user.recoveryEmail);
  return { main: main.code, recovery: recovery.code, link: main.link };
}

const resetCodeKeys = (userId: string) => [
  `reset-mfa={main:${userId}}`,
  `reset-mfa={recovery:${userId}}`,
];

function readResetCodeTTLs(userId: string) {
  return resetCodeKeys(userId).map((key) =>
    Number(composeExec("redis", ["valkey-cli", "TTL", key])),
  );
}

function deleteResetCodes(userId: string) {
  const deleted = composeExec("redis", [
    "valkey-cli",
    "DEL",
    ...resetCodeKeys(userId),
  ]);
  if (deleted !== "2") {
    throw new Error(
      `expected to delete 2 reset codes for ${userId}, got "${deleted}"`,
    );
  }
}

async function fillResetCodes(
  page: Page,
  codes: { main: string; recovery: string },
) {
  for (const [field, code] of [
    ["Main", codes.main],
    ["Recovery", codes.recovery],
  ]) {
    for (const [i, char] of [...code].entries()) {
      await page
        .getByLabel(`${field} email code character ${i + 1} of 5`)
        .fill(char);
    }
  }
}

async function requestReset(user: MFAUser) {
  await requestResetMfa({
    ...buildRequestContext(),
    body: { identifier: user.username },
  });
  const { link, ...codes } = await readResetCodes(user);
  const url = new URL(link);
  const userId = url.searchParams.get("id");
  if (!userId) throw new Error(`expected a user id in ${link}`);
  return { codes, userId, path: url.pathname + url.search };
}

async function openDisableMFA(page: Page, username: string) {
  await signInWithMFA(page, username);
  await page.goto("/account/security");
  await page.getByRole("button", { name: "Disable", exact: true }).click();
}

test.describe("MFA", () => {
  test.skip(isCommunity, "MFA exists only in enterprise and cloud");

  test.describe("enrollment", () => {
    test("sets a recovery email, then verifies a code to enable MFA", async ({
      page,
    }) => {
      const { owner } = await createTeam();
      const recoveryEmail = buildRandomEmail("recovery");
      await openEnableMFA(page, owner.username);
      const generated = page.waitForResponse((r) =>
        r.url().endsWith("/api/user/mfa/generate"),
      );

      await page.getByLabel("Recovery Email").fill(recoveryEmail);
      await page.getByRole("button", { name: "Save & Continue" }).click();
      const { recovery_codes: shownCodes } = (await (
        await generated
      ).json()) as { recovery_codes: string[] };
      expect(shownCodes).toHaveLength(6);
      for (const code of shownCodes) {
        await expect(page.getByText(code, { exact: true })).toBeVisible();
      }
      await completeEnrollment(page);
      await page.getByRole("button", { name: "Done" }).click();

      await expect(page.getByText("Enabled", { exact: true })).toBeVisible();
      expect(await readMFA(owner)).toEqual({ enabled: true, recoveryEmail });
      expect(countRecoveryCodes(owner.username)).toBe(6);
    });

    test("keeps the existing recovery email when confirmed", async ({
      page,
    }) => {
      const { owner } = await createTeam();
      const recoveryEmail = buildRandomEmail("recovery");
      await setRecoveryEmail(owner.token, recoveryEmail);
      await openEnableMFA(page, owner.username);

      await expect(
        page.getByText("Step 1: Confirm Recovery Email"),
      ).toBeVisible();
      await expect(page.getByText(recoveryEmail)).toBeVisible();
      await page.getByRole("button", { name: "Continue", exact: true }).click();
      await completeEnrollment(page);

      expect(await readMFA(owner)).toMatchObject({
        enabled: true,
        recoveryEmail,
      });
    });

    test("replaces the recovery email during setup", async ({ page }) => {
      const { owner } = await createTeam();
      await setRecoveryEmail(owner.token, buildRandomEmail("recovery"));
      const replacement = buildRandomEmail("replacement");
      await openEnableMFA(page, owner.username);

      await page
        .getByRole("button", { name: "Use a different recovery email" })
        .click();
      await page.getByLabel("Recovery Email").fill(replacement);
      await page.getByRole("button", { name: "Save & Continue" }).click();
      await completeEnrollment(page);

      expect(await readMFA(owner)).toMatchObject({
        enabled: true,
        recoveryEmail: replacement,
      });
    });

    test("rejects the account's own email as the recovery email", async ({
      page,
    }) => {
      const { owner } = await createTeam();
      await openEnableMFA(page, owner.username);

      const saved = page.waitForResponse(
        (r) =>
          r.url().endsWith("/api/users") && r.request().method() === "PATCH",
      );
      await page.getByLabel("Recovery Email").fill(owner.email);
      await page.getByRole("button", { name: "Save & Continue" }).click();

      expect((await saved).status()).toBe(400);
      await expect(
        page.getByText("Failed to save recovery email"),
      ).toBeVisible();
      await expect(page.getByText("Step 1: Set Recovery Email")).toBeVisible();
      expect(await readMFA(owner)).toEqual({
        enabled: false,
        recoveryEmail: "",
      });
    });
  });

  test("an MFA-enabled account continues to the code prompt", async ({
    page,
  }) => {
    const user = await createMFAUser();

    await page.goto("/login");
    const login = page.waitForResponse((r) => r.url().endsWith("/api/login"));
    await fillLoginForm(page, user.username, password);

    await expect(page).toHaveURL(/\/mfa-login$/);
    await expect(
      page.getByRole("heading", { name: "Two-Factor Authentication" }),
    ).toBeVisible();
    const response = await login;
    expect(response.status()).toBe(401);
    expect(response.headers()["x-mfa-token"]).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    );
  });

  test("signs in with a code from the authenticator", async ({ page }) => {
    const user = await createMFAUser();
    await reachCodePrompt(page, user.username);

    const verified = page.waitForResponse((r) =>
      r.url().endsWith("/api/user/mfa/auth"),
    );
    await fillDigits(page, totp(mfaSecret));
    await page.getByRole("button", { name: "Verify" }).click();

    await expect(page).toHaveURL(/\/dashboard$/);
    expect((await verified).status()).toBe(200);
  });

  test.describe("recovery codes", () => {
    test("a recovery code signs in and is consumed", async ({ page }) => {
      const code = buildRecoveryCode();
      const user = await createMFAUser({
        recoveryCodes: [code, buildRecoveryCode()],
      });

      await recoverWithCode(page, user.username, code);
      const dialog = page.getByRole("dialog", {
        name: "Recovery window active",
      });
      await expect(dialog).toBeVisible();
      await dialog.getByRole("button", { name: "Close" }).click();

      await expect(page).toHaveURL(/\/dashboard$/);
      expect(await readMFA(user)).toMatchObject({ enabled: true });
      expect(countRecoveryCodes(user.username)).toBe(1);
    });

    test("a used recovery code is rejected", async ({ page }) => {
      const code = buildRecoveryCode();
      const user = await createMFAUser({
        recoveryCodes: [code, buildRecoveryCode()],
      });
      await mfaRecover({
        ...buildRequestContext(),
        body: { identifier: user.username, recovery_code: code },
      });

      await recoverWithCode(page, user.username, code);

      await expect(
        page.getByText("Invalid recovery code or username"),
      ).toBeVisible();
      await expect(page).toHaveURL(/\/mfa-recover$/);
      expect(await readMFA(user)).toMatchObject({ enabled: true });
      expect(countRecoveryCodes(user.username)).toBe(1);
    });

    test("MFA can be turned off inside the recovery window", async ({
      page,
    }) => {
      const code = buildRecoveryCode();
      const user = await createMFAUser({ recoveryCodes: [code] });

      await recoverWithCode(page, user.username, code);
      await page
        .getByRole("dialog", { name: "Recovery window active" })
        .getByRole("button", { name: "Disable MFA" })
        .click();

      await expect(page).toHaveURL(/\/dashboard$/);
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });
  });

  test.describe("disabling from the profile", () => {
    test("with a code from the authenticator", async ({ page }) => {
      const user = await createMFAUser();
      await openDisableMFA(page, user.username);

      await fillDigits(page, totp(mfaSecret, 1));
      await page.getByRole("button", { name: "Disable MFA" }).click();

      await expect(
        page.getByRole("button", { name: "Enable MFA" }),
      ).toBeVisible();
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });

    test("with a recovery code", async ({ page }) => {
      const code = buildRecoveryCode();
      const user = await createMFAUser({ recoveryCodes: [code] });
      await openDisableMFA(page, user.username);

      await page
        .getByRole("button", { name: "Use recovery code instead" })
        .click();
      await page.getByLabel("Recovery Code").fill(code);
      await page.getByRole("button", { name: "Disable MFA" }).click();

      await expect(
        page.getByRole("button", { name: "Enable MFA" }),
      ).toBeVisible();
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });
  });

  test.describe("email reset", () => {
    test.skip(!isCloud, "only cloud sends email");

    test("disables MFA from the profile with both emailed codes", async ({
      page,
    }) => {
      const user = await createMFAUser();
      await openDisableMFA(page, user.username);

      await page
        .getByRole("button", { name: "Use recovery code instead" })
        .click();
      await page
        .getByRole("button", {
          name: "Lost recovery codes? Request email reset",
        })
        .click();
      await page
        .getByRole("button", { name: "Send Verification Codes" })
        .click();
      await expect(page.getByText("Emails Sent!")).toBeVisible();
      await fillResetCodes(page, await readResetCodes(user));
      await page.getByRole("button", { name: "Disable MFA" }).click();

      await expect(
        page.getByRole("button", { name: "Enable MFA" }),
      ).toBeVisible();
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });

    test("emails a code to each address, and both codes reset MFA", async ({
      page,
    }) => {
      const user = await createMFAUser();
      await reachCodePrompt(page, user.username);
      await page.getByRole("link", { name: "Use a recovery code" }).click();
      await page
        .getByRole("link", { name: "Lost the codes? Reset by email" })
        .click();

      await page
        .getByRole("button", { name: "Send Verification Codes" })
        .click();
      await expect(page).toHaveURL(/\/mfa-reset-verify$/);
      await fillResetCodes(page, await readResetCodes(user));
      await page.getByRole("button", { name: "Verify and Reset MFA" }).click();

      await expect(page).toHaveURL(/\/dashboard$/);
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });

    test("the emailed link resets MFA and signs in", async ({ page }) => {
      const user = await createMFAUser();
      const { codes, path } = await requestReset(user);

      await page.goto(path);
      await fillResetCodes(page, codes);
      await page.getByRole("button", { name: "Reset MFA and Login" }).click();

      await expect(page).toHaveURL(/\/dashboard$/);
      expect(await readMFA(user)).toMatchObject({ enabled: false });
    });

    test("the emailed codes stop working after 24 hours", async ({ page }) => {
      const user = await createMFAUser();
      const { codes, userId, path } = await requestReset(user);

      for (const ttl of readResetCodeTTLs(userId)) {
        expect(ttl).toBeGreaterThan(24 * 60 * 60 - 60);
        expect(ttl).toBeLessThanOrEqual(24 * 60 * 60);
      }
      deleteResetCodes(userId);
      await page.goto(path);
      await fillResetCodes(page, codes);
      await page.getByRole("button", { name: "Reset MFA and Login" }).click();

      await expect(
        page.getByText(
          "Invalid verification codes. Please check and try again.",
        ),
      ).toBeVisible();
      expect(await readMFA(user)).toMatchObject({ enabled: true });
    });
  });
});
