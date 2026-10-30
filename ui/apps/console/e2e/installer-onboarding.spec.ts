import { expect, test } from "@playwright/test";
import { createDevicePairing, getDevicePairingStatus } from "@/client";
import { isCloud } from "./env";
import { buildRequestContext, loginAs } from "./api";
import { buildDeviceAuthRequest } from "./devices";
import { required, submitSignUpForm } from "./helpers";
import { readLatestEmail } from "./mail";
import { buildShortId, buildUserIdentity, password } from "./seed";

test.describe("onboarding from the installer", () => {
  test.skip(!isCloud, "only the cloud has open sign-up");

  test("a new user signs up from the accept-device link, confirms on another device and accepts it", async ({
    page,
    browser,
  }) => {
    const { data: pairing } = await createDevicePairing({
      ...buildRequestContext(),
      body: buildDeviceAuthRequest(),
    });
    const code = required(pairing.code, "a pairing code for the new device");
    const account = buildUserIdentity("installer");
    const namespace = `installer-${buildShortId()}`;

    await page.goto(`/accept-device?code=${code}`);
    await page
      .getByRole("link", { name: "Create account", exact: true })
      .click();
    await submitSignUpForm(page, account);

    const { link } = await readLatestEmail(account.email, "/confirm-account");
    expect(link.params.get("device")).toBe(code);

    const phone = await browser.newContext();
    const phonePage = await phone.newPage();
    await phonePage.goto(link.path);

    await expect(
      phonePage.getByRole("heading", { name: "Accept this device?" }),
    ).toBeVisible();
    await phonePage
      .getByLabel("Namespace Name", { exact: true })
      .fill(namespace);
    await phonePage
      .getByRole("button", { name: "Create and accept", exact: true })
      .click();
    await expect(
      phonePage.getByRole("heading", { name: "Device accepted" }),
    ).toBeVisible();

    await phonePage.getByRole("button", { name: "View device" }).click();
    await expect(phonePage).toHaveURL(/\/devices\/[^/]+$/);
    await phone.close();

    const { tenant } = await loginAs(account.username, password);
    const { data: status } = await getDevicePairingStatus({
      ...buildRequestContext(),
      path: { code },
    });
    expect(status).toMatchObject({ status: "accepted", tenant_id: tenant });
  });
});
