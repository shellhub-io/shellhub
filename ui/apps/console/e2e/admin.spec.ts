import { expect, test } from "@playwright/test";
import {
  configureLocalAuthentication,
  getAuthenticationSettings,
  getUserInfo,
} from "@/client";
import { buildRequestContext, expectLoginStatus, loginAs } from "./api";
import { isEnterprise, isEnterpriseOrCloud } from "./env";
import { createTeam, dismissWizard, required, signInAndOpen } from "./helpers";
import { disableSaml, enableSaml } from "./saml";
import { password } from "./seed";

type Team = Awaited<ReturnType<typeof createTeam>>;

test.skip(
  !isEnterpriseOrCloud,
  "only enterprise and cloud have an admin panel",
);

test("an admin signs in as another user in a new tab", async ({
  page,
  context,
}) => {
  const target = await createTeam();
  const { id } = await loginAs(target.owner.username, password);
  const { owner: admin } = await createTeam({ admin: true });
  await signInAndOpen(page, admin.username, `/admin/users/${id}`);

  const opened = context.waitForEvent("page");
  const token = context
    .waitForEvent("request", (request) =>
      new URL(request.url()).searchParams.has("token"),
    )
    .then((request) =>
      required(
        new URL(request.url()).searchParams.get("token"),
        "a session token for the target user",
      ),
    );
  await page.getByRole("button", { name: "Login as User" }).click();
  const tab = await opened;

  await expect(tab).toHaveURL(/\/dashboard$/);
  await dismissWizard(tab);
  await expect(
    tab.getByRole("button", {
      name: `Account menu for ${target.owner.username}`,
    }),
  ).toBeVisible();
  const { data } = await getUserInfo(
    buildRequestContext({ token: await token }),
  );
  expect(data.user).toBe(target.owner.username);
});

test.describe("local authentication", () => {
  test.skip(
    !isEnterprise,
    "local authentication turns off only while SAML is on, which the specs set up only on enterprise",
  );

  let admin: Team["owner"];

  test.beforeEach(async () => {
    ({ owner: admin } = await createTeam({ admin: true }));
    await enableSaml();
  });
  test.afterEach(async () => {
    if (!admin) return;
    await configureLocalAuthentication({
      ...buildRequestContext({ token: admin.token }),
      body: { enable: true },
    });
    await disableSaml();
  });

  test("turning local authentication off withdraws password sign-in", async ({
    page,
    browser,
  }) => {
    await signInAndOpen(page, admin.username, "/admin/settings/authentication");

    const toggle = page.getByRole("switch", {
      name: "Toggle local authentication",
    });
    await expect(toggle).toBeChecked();
    await toggle.click();
    await expect(toggle).not.toBeChecked();

    const { data } = await getAuthenticationSettings(
      buildRequestContext({ token: admin.token }),
    );
    expect(data.local?.enabled).toBe(false);

    const visitorContext = await browser.newContext();
    const visitor = await visitorContext.newPage();
    await visitor.goto("/login");
    await expect(
      visitor.getByRole("button", { name: "Login with SSO" }),
    ).toBeVisible();
    await expect(visitor.getByLabel("Password", { exact: true })).toHaveCount(
      0,
    );

    await expectLoginStatus(admin.username, password, 501);
    await visitorContext.close();
  });
});
