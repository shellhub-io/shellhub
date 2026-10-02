import { type Locator, type Page, expect, test } from "@playwright/test";
import { generateKeyPairSync } from "node:crypto";
import { getVault } from "@/client";
import { isCommunity } from "./env";
import {
  createTeamWithOwnerInAnother,
  createTeam,
  createTeamWithMember,
  signInAndOpen,
  signOut,
  singleNamespaceReason,
  switchNamespace,
} from "./helpers";
import { buildShortId } from "./seed";
import { expectStatus } from "./api";

const masterPassword = "vault-master-password";
const serverStorageReason = "community keeps the vault in the browser only";

const storageOptions = {
  local: "This device only",
  server: "Sync to the ShellHub server",
};

function buildPrivateKey() {
  return generateKeyPairSync("rsa", {
    modulusLength: 2048,
    privateKeyEncoding: { type: "pkcs1", format: "pem" },
    publicKeyEncoding: { type: "spki", format: "pem" },
  }).privateKey;
}

async function openVault(page: Page) {
  const team = await createTeam({ sshAccessMode: "legacy" });
  await signInAndOpen(page, team.owner.username, "/secure-vault");
  return team;
}

const uninitializedState = (page: Page) =>
  page.getByRole("button", { name: "Set Up Secure Vault" });

async function openSetUp(page: Page) {
  await uninitializedState(page).click();
  return page.getByRole("dialog", { name: "Set up secure vault" });
}

async function createVault(
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

async function setUpVault(page: Page, storage?: keyof typeof storageOptions) {
  await createVault(await openSetUp(page), storage);
}

const keyEntry = (page: Page, name: string) =>
  page.getByRole("button", { name: `Edit ${name}` });

async function addKey(page: Page) {
  const name = `e2e-key-${buildShortId()}`;
  await page.getByRole("button", { name: "Add Private Key" }).click();
  const dialog = page.getByRole("dialog", { name: "Add private key" });
  await dialog.getByLabel("Name", { exact: true }).fill(name);
  await dialog.getByRole("button", { name: "Text", exact: true }).click();
  await dialog
    .getByLabel("Private Key", { exact: true })
    .fill(buildPrivateKey());
  await dialog.getByRole("button", { name: "Add key" }).click();
  await expect(keyEntry(page, name)).toBeVisible();
  return name;
}

const lockedState = (page: Page) =>
  page.getByRole("heading", { name: "Your vault is locked" });

async function lock(page: Page) {
  await page.getByRole("button", { name: "Lock vault" }).click();
  await expect(lockedState(page)).toBeVisible();
}

async function submitUnlock(dialog: Locator, attempt: string) {
  await dialog.getByLabel("Master Password", { exact: true }).fill(attempt);
  await dialog.getByRole("button", { name: "Unlock", exact: true }).click();
}

async function unlock(page: Page, attempt: string) {
  await page.getByRole("button", { name: "Unlock Vault" }).click();
  const dialog = page.getByRole("dialog", { name: "Unlock vault" });
  await submitUnlock(dialog, attempt);
  return dialog;
}

async function expectUnlockRefused(page: Page, attempt: string) {
  const dialog = await unlock(page, attempt);
  await expect(dialog.getByRole("alert")).toHaveText(
    "Incorrect master password",
  );
  return dialog;
}

function readLocalVaultEntries(page: Page) {
  return page.evaluate(() =>
    Object.keys(localStorage).filter((key) =>
      /^shellhub-vault-(meta|data):/.test(key),
    ),
  );
}

async function resetVault(page: Page) {
  await page.getByRole("button", { name: "Reset vault" }).click();
  const dialog = page.getByRole("dialog", { name: "Reset secure vault" });
  await dialog.getByLabel('Type "RESET" to confirm').fill("RESET");
  await dialog.getByRole("button", { name: "Reset vault" }).click();
  await expect(uninitializedState(page)).toBeVisible();
}

test.describe("Secure vault", () => {
  test("a new vault opens unlocked and locks on reload", async ({ page }) => {
    await openVault(page);

    await setUpVault(page);
    await expect(
      page.getByRole("button", { name: "Add Private Key" }),
    ).toBeVisible();

    await page.reload();
    await expect(lockedState(page)).toBeVisible();
  });

  test("imports the keys stored before the vault existed", async ({ page }) => {
    await openVault(page);
    await page.evaluate(
      (data) =>
        localStorage.setItem(
          "privateKeys",
          JSON.stringify([
            {
              id: 1,
              name: "legacy-key",
              data,
              hasPassphrase: false,
              fingerprint: "legacy-fingerprint",
            },
          ]),
        ),
      buildPrivateKey(),
    );

    const dialog = await openSetUp(page);
    await expect(dialog).toContainText(
      "1 existing key will be imported and encrypted.",
    );
    await createVault(dialog);

    await expect(keyEntry(page, "legacy-key")).toBeVisible();
    await expect
      .poll(() => page.evaluate(() => localStorage.getItem("privateKeys")))
      .toBeNull();
  });

  test("unlocks with the master password and refuses a wrong one", async ({
    page,
  }) => {
    await openVault(page);
    await setUpVault(page);
    const key = await addKey(page);
    await page.reload();

    const dialog = await expectUnlockRefused(page, "wrong-master-password");

    await submitUnlock(dialog, masterPassword);
    await expect(keyEntry(page, key)).toBeVisible();
  });

  test("a changed master password replaces the old one", async ({ page }) => {
    const newMasterPassword = `${masterPassword}-new`;
    await openVault(page);
    await setUpVault(page);
    const key = await addKey(page);

    await page.getByRole("button", { name: "Change master password" }).click();
    const dialog = page.getByRole("dialog", {
      name: "Change master password",
    });
    await dialog
      .getByLabel("Current Password", { exact: true })
      .fill(masterPassword);
    await dialog
      .getByLabel("New Password", { exact: true })
      .fill(newMasterPassword);
    await dialog
      .getByLabel("Confirm New Password", { exact: true })
      .fill(newMasterPassword);
    await dialog.getByRole("button", { name: "Update Password" }).click();
    await expect(dialog).toBeHidden();
    await lock(page);

    const unlockDialog = await expectUnlockRefused(page, masterPassword);

    await submitUnlock(unlockDialog, newMasterPassword);
    await expect(keyEntry(page, key)).toBeVisible();
  });

  test("resetting destroys the vault and its keys", async ({ page }) => {
    test.skip(
      !isCommunity,
      "enterprise and cloud store a new vault on the server, reset there by the storage tests",
    );

    await openVault(page);
    await setUpVault(page);
    await addKey(page);

    await resetVault(page);

    await expect.poll(() => readLocalVaultEntries(page)).toHaveLength(0);
    await page.reload();
    await expect(uninitializedState(page)).toBeVisible();
  });

  test("each user has their own vault", async ({ page }) => {
    const { owner, member } = await createTeamWithMember("administrator", {
      sshAccessMode: "legacy",
    });

    await signInAndOpen(page, owner.username, "/secure-vault");
    await setUpVault(page);
    await addKey(page);
    await signOut(page, owner.username);

    await signInAndOpen(page, member.username, "/secure-vault");
    await expect(uninitializedState(page)).toBeVisible();
    await signOut(page, member.username);

    await signInAndOpen(page, owner.username, "/secure-vault");
    await expect(lockedState(page)).toBeVisible();
  });

  test("each namespace has its own vault", async ({ page }) => {
    test.skip(isCommunity, singleNamespaceReason);

    const { owner, namespace, other } = await createTeamWithOwnerInAnother({
      sshAccessMode: "legacy",
    });

    await signInAndOpen(page, owner.username, "/secure-vault");
    await setUpVault(page);
    const key = await addKey(page);

    await switchNamespace(page, other.namespace);
    await page.goto("/secure-vault");
    await expect(uninitializedState(page)).toBeVisible();

    await switchNamespace(page, namespace);
    await page.goto("/secure-vault");
    await unlock(page, masterPassword);
    await expect(keyEntry(page, key)).toBeVisible();
  });

  test("switching namespace locks the vault", async ({ page }) => {
    test.skip(isCommunity, singleNamespaceReason);

    const { owner, namespace, other } = await createTeamWithOwnerInAnother({
      sshAccessMode: "legacy",
    });
    await signInAndOpen(page, owner.username, "/secure-vault");
    await setUpVault(page);

    await switchNamespace(page, other.namespace);
    await switchNamespace(page, namespace);

    await page.getByRole("link", { name: "Secure Vault" }).click();
    await expect(lockedState(page)).toBeVisible();
  });

  test.describe("storage", () => {
    test.skip(isCommunity, serverStorageReason);

    test("a device-only vault stays off the server", async ({ page }) => {
      const { owner } = await openVault(page);

      await setUpVault(page, "local");

      await expect.poll(() => readLocalVaultEntries(page)).toHaveLength(2);
      await expectStatus(getVault, { token: owner.token }, 404);
    });

    test("a synced vault is stored on the server", async ({ page }) => {
      const { owner } = await openVault(page);

      await setUpVault(page, "server");

      await expect.poll(() => readLocalVaultEntries(page)).toHaveLength(0);
      await expectStatus(getVault, { token: owner.token }, 200);
    });

    test("resetting a synced vault deletes it from the server", async ({
      page,
    }) => {
      const { owner } = await openVault(page);
      await setUpVault(page, "server");
      await addKey(page);
      await expectStatus(getVault, { token: owner.token }, 200);

      await resetVault(page);

      await expectStatus(getVault, { token: owner.token }, 404);
    });

    test("moves a device-only vault to the server", async ({ page }) => {
      const { owner } = await openVault(page);
      await setUpVault(page, "local");
      const key = await addKey(page);

      await page
        .getByRole("button", { name: "Change vault storage location" })
        .click();
      const dialog = page.getByRole("dialog", {
        name: "Sync vault to the ShellHub server",
      });
      await dialog.getByRole("button", { name: "Sync vault" }).click();
      await expect(dialog).toBeHidden();

      await unlock(page, masterPassword);
      await expect(keyEntry(page, key)).toBeVisible();
      await expect.poll(() => readLocalVaultEntries(page)).toHaveLength(0);
      await expectStatus(getVault, { token: owner.token }, 200);
    });

    test("moves a synced vault to this device", async ({ page }) => {
      const { owner } = await openVault(page);
      await setUpVault(page, "server");
      const key = await addKey(page);

      await page
        .getByRole("button", { name: "Change vault storage location" })
        .click();
      const dialog = page.getByRole("dialog", {
        name: "Move vault to this device",
      });
      await dialog.getByRole("button", { name: "Move vault" }).click();
      await expect(dialog).toBeHidden();

      await unlock(page, masterPassword);
      await page
        .getByRole("dialog", { name: "Take your vault anywhere" })
        .getByRole("button", { name: "Keep it on this device" })
        .click();
      await expect(keyEntry(page, key)).toBeVisible();
      await expect.poll(() => readLocalVaultEntries(page)).toHaveLength(2);
      await expectStatus(getVault, { token: owner.token }, 404);
    });
  });
});
