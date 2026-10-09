import { type Page, expect } from "@playwright/test";
import { createHmac, randomUUID } from "node:crypto";
import { enableMfa, updateUser } from "@/client";
import { buildRequestContext } from "./api";
import { dismissWizard, signIn } from "./helpers";
import { buildRandomEmail, password } from "./seed";

export const mfaSecret = "JBSWY3DPEHPK3PXP";

const recoveryCodeCount = 6;

const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

export function totp(secret: string, stepOffset = 0) {
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

export async function fillDigits(page: Page, code: string) {
  for (const [i, digit] of [...code].entries()) {
    await page
      .getByRole("group", { name: "Verification Code" })
      .getByRole("textbox")
      .nth(i)
      .fill(digit);
  }
}

export async function reachCodePrompt(page: Page, username: string) {
  await signIn(page, username, password);
  await expect(page).toHaveURL(/\/mfa-login$/);
}

export async function signInWithMFA(page: Page, username: string) {
  await reachCodePrompt(page, username);
  await fillDigits(page, totp(mfaSecret));
  await page.getByRole("button", { name: "Verify" }).click();
  await expect(page).toHaveURL(/\/dashboard$/);
  await dismissWizard(page);
}

export const buildRecoveryCode = () =>
  randomUUID().replaceAll("-", "").slice(0, 16).toUpperCase();

export async function setRecoveryEmail(token: string, recoveryEmail: string) {
  await updateUser({
    ...buildRequestContext({ token }),
    body: { recovery_email: recoveryEmail },
  });
}

export async function enableMFA(
  token: string,
  {
    recoveryEmail = buildRandomEmail("recovery"),
    recoveryCodes = [] as string[],
  } = {},
) {
  await setRecoveryEmail(token, recoveryEmail);
  const codes = [
    ...recoveryCodes,
    ...Array.from(
      { length: recoveryCodeCount - recoveryCodes.length },
      buildRecoveryCode,
    ),
  ];
  const enable = async () => {
    const { response } = await enableMfa({
      ...buildRequestContext({ token }),
      throwOnError: false,
      body: {
        secret: mfaSecret,
        code: totp(mfaSecret, -1),
        recovery_codes: codes,
      },
    });
    return response?.status;
  };

  let status = await enable();
  if (status === 403) status = await enable();
  expect(status, "enabling MFA").toBe(200);
  return { recoveryEmail, recoveryCodes: codes };
}
