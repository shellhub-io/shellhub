import { expect, test, type Page } from "@playwright/test";
import { adminUser } from "./seed";

async function signIn(page: Page, username: string, password: string) {
  await page.goto("/login");
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();
}

test.describe("authentication", () => {
  test("signs in with valid credentials", async ({ page }) => {
    await signIn(page, adminUser.username, adminUser.password);

    await expect(page).toHaveURL(/\/dashboard$/);
  });

  test("rejects invalid credentials", async ({ page }) => {
    await signIn(page, "e2e-nobody", "wrong-password");

    await expect(page.getByText("Invalid login credentials")).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);
  });

  test("signs out", async ({ page }) => {
    await signIn(page, adminUser.username, adminUser.password);
    await expect(page).toHaveURL(/\/dashboard$/);

    await page.getByRole("button", { name: "Close wizard" }).click();
    await page
      .getByRole("button", { name: `Account menu for ${adminUser.username}` })
      .click();
    await page.getByRole("button", { name: "Logout" }).click();

    await expect(page).toHaveURL(/\/login$/);
  });
});
