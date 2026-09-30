import { type Page, expect, test } from "@playwright/test";
import {
  apiKeyList,
  getStats,
  listInstanceApiKeys,
  listNamespaceMembers,
} from "@/client";
import { isCommunity } from "./env";
import { createTeam, signInAndOpen } from "./helpers";
import { buildShortId } from "./seed";
import { buildRequestContext, expectStatus } from "./api";

function expectListedWithoutKey(keys: { name: string }[], name: string) {
  const listed = keys.find((k) => k.name === name);
  expect(listed).toBeDefined();
  expect(listed).not.toHaveProperty("key");
}

async function generateKey(
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
  await expect(page.getByRole("row").filter({ hasText: name })).toBeVisible();
  return { name, key };
}

test.describe("API key plaintext", () => {
  test("a namespace key shows its plaintext only at creation", async ({
    page,
  }) => {
    const { owner, tenant } = await createTeam();

    await signInAndOpen(page, owner.username, "/team");
    await page.getByRole("button", { name: "API Keys" }).click();
    const { name, key } = await generateKey(page, {
      dialog: "Generate API key",
      submit: "Generate Key",
      keyLabel: "Your API Key",
    });
    await expect(page.getByText(key)).toHaveCount(0);

    await expectStatus(
      (opts) => listNamespaceMembers({ ...opts, path: { tenant } }),
      { apiKey: key },
      200,
    );
    const { data } = await apiKeyList(
      buildRequestContext({ token: owner.token }),
    );
    expectListedWithoutKey(data, name);
  });

  test("an instance key shows its plaintext only at creation", async ({
    page,
  }) => {
    test.skip(
      isCommunity,
      "instance API keys exist only in enterprise and cloud",
    );
    const { owner: admin } = await createTeam({ admin: true });

    await signInAndOpen(page, admin.username, "/admin/instance-api-keys");
    const { name, key } = await generateKey(page, {
      dialog: "Generate instance API key",
      submit: "Generate",
      keyLabel: "Your Instance API Key",
    });
    await expect(page.getByText(key)).toHaveCount(0);

    await expectStatus(getStats, { apiKey: key }, 200);
    const { data } = await listInstanceApiKeys(
      buildRequestContext({ token: admin.token }),
    );
    expectListedWithoutKey(data, name);
  });
});
