import type { Page } from "@playwright/test";

export const directMembershipReason =
  "enterprise adds existing users directly, without an invitation link";

export async function fillLoginForm(
  page: Page,
  username: string,
  password: string,
) {
  await page.getByLabel("Username", { exact: true }).fill(username);
  await page.getByLabel("Password", { exact: true }).fill(password);
  await page.getByRole("button", { name: "Sign In" }).click();
}

export async function signIn(page: Page, username: string, password: string) {
  await page.goto("/login");
  await fillLoginForm(page, username, password);
}

export async function dismissWizard(page: Page) {
  const close = page.getByRole("button", { name: "Close wizard" });
  if (
    await close.waitFor({ state: "visible", timeout: 5000 }).then(
      () => true,
      () => false,
    )
  ) {
    await close.click();
    await close.waitFor({ state: "hidden" });
  }
}
