import { type Locator, type Page, expect } from "@playwright/test";
import { generateKeyPairSync } from "node:crypto";
import { buildShortId } from "./seed";

export const masterPassword = "vault-master-password";

const storageOptions = {
  local: "This device only",
  server: "Sync to the ShellHub server",
};

export function buildPrivateKey({ passphrase = "" } = {}) {
  return generateKeyPairSync("rsa", {
    modulusLength: 2048,
    privateKeyEncoding: passphrase
      ? { type: "pkcs1", format: "pem", cipher: "aes-256-cbc", passphrase }
      : { type: "pkcs1", format: "pem" },
    publicKeyEncoding: { type: "spki", format: "pem" },
  }).privateKey;
}

export const uninitializedState = (page: Page) =>
  page.getByRole("button", { name: "Set Up Secure Vault" });

export async function openSetUp(page: Page) {
  await uninitializedState(page).click();
  return page.getByRole("dialog", { name: "Set up secure vault" });
}

export async function createVault(
  dialog: Locator,
  storage?: keyof typeof storageOptions,
) {
  if (storage) {
    await dialog.getByText(storageOptions[storage]).check();
  }
  await dialog
    .getByLabel("Master Password", { exact: true })
    .fill(masterPassword);
  await dialog
    .getByLabel("Confirm Password", { exact: true })
    .fill(masterPassword);
  await dialog.getByRole("button", { name: "Create Vault" }).click();
  await expect(dialog).toBeHidden();
}

export async function setUpVault(
  page: Page,
  storage?: keyof typeof storageOptions,
) {
  await createVault(await openSetUp(page), storage);
}

export const keyEntry = (page: Page, name: string) =>
  page.getByRole("button", { name: `Edit ${name}` });

export async function addKey(
  page: Page,
  { privateKey = buildPrivateKey(), passphrase = "" } = {},
) {
  const name = `e2e-key-${buildShortId()}`;
  await page.getByRole("button", { name: "Add Private Key" }).click();
  const dialog = page.getByRole("dialog", { name: "Add private key" });
  await dialog.getByLabel("Name", { exact: true }).fill(name);
  await dialog.getByRole("button", { name: "Text", exact: true }).click();
  await dialog.getByLabel("Private Key", { exact: true }).fill(privateKey);
  if (passphrase) {
    await dialog.getByLabel("Passphrase", { exact: true }).fill(passphrase);
  }
  await dialog.getByRole("button", { name: "Add key" }).click();
  await expect(keyEntry(page, name)).toBeVisible();
  return name;
}
