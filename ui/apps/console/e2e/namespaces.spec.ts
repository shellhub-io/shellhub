import { type Page, expect, test } from "@playwright/test";
import { listAccessPolicies } from "@/client";
import { isCommunity } from "./env";
import {
  createTeamWithOwnerInAnother,
  createTeam,
  findRow,
  holdRoute,
  signInAndOpen,
  singleNamespaceReason,
  switchNamespace,
} from "./helpers";
import { buildRequestContext, createApiKey } from "./api";

const isApiKeyList = (url: URL) => url.pathname === "/api/namespaces/api-key";

async function upgradeNamespace(page: Page, namespace: string) {
  const mode = page.getByRole("group", { name: "Namespace" });
  await mode.getByRole("button", { name: "Upgrade" }).click();
  const dialog = page.getByRole("dialog", {
    name: "Upgrade this namespace?",
  });
  await dialog.getByLabel(`Type "${namespace}" to confirm`).fill(namespace);
  await dialog.getByRole("button", { name: "Upgrade namespace" }).click();
  await expect(mode.getByText("New")).toBeVisible();
  await expect(mode.getByRole("button", { name: "Upgrade" })).toHaveCount(0);
}

async function readAccessPolicyNames(token: string) {
  const { data } = await listAccessPolicies(buildRequestContext({ token }));
  return data.map((policy) => policy.name);
}

test.describe("Namespaces", () => {
  test("switching drops the previous namespace's data and loads the new one's", async ({
    page,
  }) => {
    test.skip(isCommunity, singleNamespaceReason);

    const { owner, other } = await createTeamWithOwnerInAnother();
    const ownKey = await createApiKey(owner.token);
    const otherKey = await createApiKey(other.owner.token);

    await signInAndOpen(page, owner.username, "/team");
    await page.getByRole("tab", { name: "API Keys" }).click();
    await expect(findRow(page, ownKey.name)).toBeVisible();

    const apiKeyList = await holdRoute(page, isApiKeyList);
    await switchNamespace(page, other.namespace);
    await page.getByRole("link", { name: "Team" }).click();
    await page.getByRole("tab", { name: "API Keys" }).click();
    await apiKeyList.requested();

    await expect(findRow(page, ownKey.name)).toHaveCount(0);
    await apiKeyList.release();
    await expect(findRow(page, otherKey.name)).toBeVisible();
  });

  test.describe("upgrading a legacy namespace", () => {
    test("seeds the owner's access policy", async ({ page }) => {
      const { owner, namespace } = await createTeam({ legacy: true });
      await signInAndOpen(page, owner.username, "/settings/ssh");

      await upgradeNamespace(page, namespace);

      expect(await readAccessPolicyNames(owner.token)).toEqual([
        "Owner access",
      ]);
    });
  });
});
