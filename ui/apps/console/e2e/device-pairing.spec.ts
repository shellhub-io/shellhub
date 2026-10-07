import { expect, test } from "@playwright/test";
import { createDeviceLoginCode, getDevice } from "@/client";
import { isCommunity } from "./env";
import {
  createTeamWithOwnerInAnother,
  signInAndOpen,
  singleNamespaceReason,
} from "./helpers";
import { enrollDevice } from "./devices";
import { buildRequestContext } from "./api";

async function requestLoginCode(tenant: string) {
  const device = await enrollDevice(tenant);
  const { data } = await createDeviceLoginCode(
    buildRequestContext({ token: device.token }),
  );
  if (!data.code) throw new Error(`expected a login code for ${device.name}`);
  return { code: data.code, uid: device.uid, name: device.name };
}

test.describe("accepting a device by its login code", () => {
  test.skip(isCommunity, singleNamespaceReason);

  test("the console switches to the namespace of a device in another one", async ({
    page,
  }) => {
    const { namespace, owner, other } = await createTeamWithOwnerInAnother();
    const device = await requestLoginCode(other.tenant);

    await signInAndOpen(page, owner.username, "/dashboard");
    await expect(page.getByRole("tab", { selected: true })).toHaveText(
      new RegExp(`${namespace}$`),
    );

    await page.goto(`/accept-device?code=${device.code}`);
    await expect(
      page.getByRole("heading", { name: "Accept this device?" }),
    ).toBeVisible();
    await expect(
      page.getByText(`A device is asking to join ${other.namespace}`),
    ).toBeVisible();
    await page.getByRole("button", { name: "Accept device" }).click();
    await expect(
      page.getByRole("heading", { name: "Device accepted" }),
    ).toBeVisible();
    await page.getByRole("button", { name: "View device" }).click();

    await expect(
      page.getByRole("heading", { name: device.name }),
    ).toBeVisible();
    await expect(page.getByRole("tab", { selected: true })).toHaveText(
      new RegExp(`${other.namespace}$`),
    );
    const { data } = await getDevice({
      ...buildRequestContext({ token: other.owner.token }),
      path: { uid: device.uid },
    });
    expect(data.status).toBe("accepted");
  });
});
