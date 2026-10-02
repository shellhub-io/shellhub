import { type Page, expect, test } from "@playwright/test";
import { createAccessPolicy, listAccessPolicies } from "@/client";
import { isCommunity } from "./env";
import {
  createTeamWithOwnerInAnother,
  createTeam,
  findRow,
  signInAndOpen,
  singleNamespaceReason,
  switchNamespace,
} from "./helpers";
import { buildShortId } from "./seed";
import { buildRequestContext, createApiKey } from "./api";

const isApiKeyList = (url: URL) => url.pathname === "/api/namespaces/api-key";

async function holdApiKeyList(page: Page) {
  let release = () => {};
  const released = new Promise<void>((resolve) => (release = resolve));
  await page.route(isApiKeyList, async (route) => {
    await released;
    await route.continue();
  });
  return release;
}

async function switchToIdentity(page: Page) {
  const mode = page.getByRole("group", { name: "SSH access mode" });
  await mode.getByRole("button", { name: "Switch to identity" }).click();
  await page
    .getByRole("dialog", { name: "Switch to identity access?" })
    .getByRole("button", { name: "Switch to identity" })
    .click();
  await expect(
    mode.getByRole("button", { name: "Switch to legacy" }),
  ).toBeVisible();
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
    await page.getByRole("button", { name: "API Keys" }).click();
    await expect(findRow(page, ownKey.name)).toBeVisible();

    const releaseApiKeyList = await holdApiKeyList(page);
    const apiKeyList = page.waitForRequest((request) =>
      isApiKeyList(new URL(request.url())),
    );
    await switchNamespace(page, other.namespace);
    await page.getByRole("link", { name: "Team" }).click();
    await page.getByRole("button", { name: "API Keys" }).click();
    await apiKeyList;

    await expect(findRow(page, ownKey.name)).toHaveCount(0);
    releaseApiKeyList();
    await expect(findRow(page, otherKey.name)).toBeVisible();
  });

  test.describe("SSH access mode", () => {
    test("switching to identity seeds the owner's access policy", async ({
      page,
    }) => {
      const { owner } = await createTeam({ sshAccessMode: "legacy" });
      await signInAndOpen(page, owner.username, "/settings/ssh");

      await switchToIdentity(page);

      expect(await readAccessPolicyNames(owner.token)).toEqual([
        "Owner access",
      ]);
    });

    test("switching to identity keeps existing policies without seeding another", async ({
      page,
    }) => {
      const { owner } = await createTeam({ sshAccessMode: "legacy" });
      const name = `e2e-policy-${buildShortId()}`;
      await createAccessPolicy({
        ...buildRequestContext({ token: owner.token }),
        body: {
          name,
          subject: { type: "all-members", value: "" },
          filter: {},
          logins: ["*"],
        },
      });
      await signInAndOpen(page, owner.username, "/settings/ssh");

      await switchToIdentity(page);

      expect(await readAccessPolicyNames(owner.token)).toEqual([name]);
    });
  });
});
