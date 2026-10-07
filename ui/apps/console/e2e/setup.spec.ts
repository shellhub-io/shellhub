import { type Page, test as base, expect } from "@playwright/test";
import { setup } from "@/client";
import { buildRequestContext, loginAs } from "./api";
import { isCloud, requireEnv } from "./env";
import { buildUserIdentity, closeSetup, password, reopenSetup } from "./seed";

const baseOrigin = new URL(requireEnv("E2E_BASE_URL")).origin;

const test = base.extend<{ freshInstance: void }>({
  freshInstance: [
    async ({ page }, provide) => {
      await page.route(
        (url) => url.origin !== baseOrigin,
        (route) => route.abort(),
      );
      reopenSetup();
      await provide();
      closeSetup();
    },
    { auto: true },
  ],
});

async function fillSetupForm(page: Page, prefix: string) {
  const user = buildUserIdentity(prefix);
  await page.goto("/");
  await expect(page).toHaveURL(/\/setup$/);
  await page.getByLabel("Name", { exact: true }).fill("E2E Setup");
  await page.getByLabel("Username", { exact: true }).fill(user.username);
  await page.getByLabel("Email", { exact: true }).fill(user.email);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByLabel("Confirm Password", { exact: true }).fill(password);
  return user;
}

test.describe("setting up a fresh instance", () => {
  test.skip(isCloud, "the cloud has no instance setup");

  test("the new administrator is signed in on the first run", async ({
    page,
  }) => {
    const user = await fillSetupForm(page, "setup");
    await page.getByRole("button", { name: "Create and continue" }).click();

    await expect(page).toHaveURL(/\/dashboard$/);
    await expect(page.getByText(`Signed in as ${user.email}`)).toBeVisible();
    await expect(
      page.getByRole("listitem").filter({
        has: page.getByRole("heading", { name: "Install the agent" }),
      }),
    ).toHaveAttribute("aria-current", "step");
    const account = await loginAs(user.username, password);
    expect(account).toMatchObject({ user: user.username, admin: true });
  });

  test("a setup finished elsewhere while the form is open is reported as done", async ({
    page,
  }) => {
    await fillSetupForm(page, "setup");
    const other = buildUserIdentity("other");
    await setup({
      ...buildRequestContext(),
      body: {
        name: "E2E Other",
        username: other.username,
        email: other.email,
        password,
        namespace: other.username,
      },
    });

    const refusal = page.waitForResponse((response) =>
      response.url().endsWith("/api/setup"),
    );
    await page.getByRole("button", { name: "Create and continue" }).click();

    expect((await refusal).status()).toBe(409);
    await expect(
      page.getByText("Setup has already been completed."),
    ).toBeVisible();
  });
});
