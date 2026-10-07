import { type Page, expect, test } from "@playwright/test";
import { getLicense, sendLicense } from "@/client";
import { buildRequestContext } from "./api";
import { isEnterprise, requireEnv } from "./env";
import { createTeam, required, signInAndOpen } from "./helpers";
import { composeExec, deleteLicenses } from "./seed";

test.skip(!isEnterprise, "only enterprise installs a license");

const guardedPages = [
  { path: "/admin/instance", heading: "Instance" },
  { path: "/admin/users", heading: "Users" },
  { path: "/admin/namespaces", heading: "Namespaces" },
  { path: "/admin/instance-api-keys", heading: "Instance API Keys" },
  { path: "/admin/settings/authentication", heading: "Authentication" },
];

let adminToken: string;

const readMountedLicense = () =>
  composeExec("server", ["cat", "/etc/shellhub/license.dat"]);

test.beforeEach(async ({ page }) => {
  const { owner } = await createTeam({ admin: true });
  adminToken = owner.token;
  await signInAndOpen(page, owner.username, "/admin/license");
});

test.afterEach(() => installLicense(readMountedLicense()));

async function installLicense(contents: string) {
  await sendLicense({
    ...buildRequestContext({ token: adminToken }),
    body: { file: new Blob([contents]) },
  });
}

async function expectRedirectedToLicense(page: Page, message: string) {
  for (const { path } of guardedPages) {
    await page.goto(path);
    await expect(page, path).toHaveURL(/\/admin\/license$/);
    await expect(page.getByText(message, { exact: true })).toBeVisible();
  }
}

test("uploading a valid license shows its details", async ({ page }) => {
  deleteLicenses();
  await page.reload();
  await expect(
    page.getByText("You do not have an installed license"),
  ).toBeVisible();

  const chooser = page.waitForEvent("filechooser");
  await page
    .getByRole("button", { name: "Choose a .dat file or drag and drop" })
    .click();
  await (
    await chooser
  ).setFiles({
    name: "license.dat",
    mimeType: "application/octet-stream",
    buffer: Buffer.from(readMountedLicense()),
  });
  await page.getByRole("button", { name: "Upload license file" }).click();

  await expect(
    page
      .getByRole("status")
      .filter({ hasText: "License uploaded successfully." }),
  ).toBeVisible();
  const { data } = await getLicense(buildRequestContext({ token: adminToken }));
  await expect(
    page.getByText(required(data.customer?.email, "the license's owner")),
  ).toBeVisible();
  expect(data.expired).toBe(false);
});

test("without a license every admin page sends the admin to the license page", async ({
  page,
}) => {
  deleteLicenses();

  await expectRedirectedToLicense(page, "You do not have an installed license");
});

test("with an expired license every admin page sends the admin to the license page", async ({
  page,
}) => {
  await installLicense(
    requireEnv(
      "E2E_EXPIRED_LICENSE",
      "set it to a license that expired more than 7 days ago",
    ),
  );

  await expectRedirectedToLicense(page, "Your license has expired!");
});

test("with a valid license every admin page opens", async ({ page }) => {
  for (const { path, heading } of guardedPages) {
    await page.goto(path);
    await expect(page, path).toHaveURL(new RegExp(`${path}$`));
    await expect(
      page.getByRole("heading", { level: 1, name: heading, exact: true }),
    ).toBeVisible();
  }
});
