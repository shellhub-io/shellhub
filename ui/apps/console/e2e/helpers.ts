import { execFileSync } from "node:child_process";
import type { Page } from "@playwright/test";

const stackName = `shellhub-e2e-${process.env.E2E_STACK_NAME || "default"}`;

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

function serverAdmin(...args: string[]): string {
  return execFileSync(
    "docker",
    [
      "compose",
      "-p",
      stackName,
      "exec",
      "-T",
      "server",
      "/server",
      "admin",
      ...args,
    ],
    { encoding: "utf-8", stdio: ["pipe", "pipe", "pipe"], timeout: 30_000 },
  ).trim();
}

export function createUser(username: string, password: string, email: string) {
  serverAdmin("user", "create", username, password, email);
}

export function createNamespace(owner: string, name: string, tenant: string) {
  serverAdmin("namespace", "create", name, owner, tenant);
}
