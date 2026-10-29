import { expect, test } from "@playwright/test";
import { createDevicePairing, getDevicePairingStatus } from "@/client";
import { mockSurveyPayload, surveyEnvironment } from "@/tests/onboardingSurvey";
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
    await phone.route("**/api/v1/client/*/environment", (route) =>
      route.fulfill({ json: surveyEnvironment(mockSurveyPayload()) }),
    );
    let surveyResponses = 0;
    await phone.route("**/api/v1/client/*/responses", (route) => {
      surveyResponses += 1;
      return route.fulfill({ json: { data: { id: "response-1" } } });
    });
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

    await expect(phonePage.getByText("Question 1 of 3")).toBeVisible();
    await expect(
      phonePage.getByRole("button", { name: "Skip survey" }),
    ).toBeHidden();
    await expect(
      phonePage.getByRole("button", { name: "View device" }),
    ).toBeHidden();
    await phonePage.getByRole("button", { name: "Next", exact: true }).click();
    await phonePage.getByRole("radio", { name: "Developer" }).press("Space");
    await phonePage.getByRole("button", { name: "Next", exact: true }).click();
    await phonePage
      .getByRole("button", { name: "Finish", exact: true })
      .click();
    const viewDevice = phonePage.getByRole("button", { name: "View device" });
    await expect(viewDevice).toBeVisible();
    expect(surveyResponses).toBe(1);
    await viewDevice.click();
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
