import { expect, test } from "@playwright/test";
import { getDevice } from "@/client";
import { buildRequestContext } from "./api";
import { enrollDevice } from "./devices";
import { createTeamWithMember, signInAndOpen } from "./helpers";

test("an operator reaches a device registered with the tenant ID from the device list through its key's activity and accepts it", async ({
  page,
}) => {
  const team = await createTeamWithMember("operator");
  const device = await enrollDevice({ tenant: team.tenant });

  await signInAndOpen(page, team.member.username, "/devices");
  await page.getByRole("button", { name: /waiting for approval 1/i }).click();
  await page
    .getByRole("link", { name: /tenant-only registration.*1 waiting/i })
    .click();

  await expect(
    page.getByRole("heading", { name: "Tenant-only registration" }),
  ).toBeVisible();
  const row = page.getByRole("listitem").filter({ hasText: device.name });
  await row.getByRole("button", { name: "Accept" }).click();
  await page.getByRole("button", { name: "Accept device" }).click();

  await expect(row).toContainText("accepted");
  const { data } = await getDevice({
    ...buildRequestContext({ token: team.owner.token }),
    path: { uid: device.uid },
  });
  expect(data.status).toBe("accepted");
});
