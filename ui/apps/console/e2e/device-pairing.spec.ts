import { expect, test } from "@playwright/test";
import { createDeviceLoginCode, getDevice, updateDeviceStatus } from "@/client";
import { isCloud, isCommunity } from "./env";
import {
  createTeam,
  createTeamWithOwnerInAnother,
  signInAndOpen,
  singleNamespaceReason,
} from "./helpers";
import { enrollDevice } from "./devices";
import { buildRequestContext } from "./api";
import { limitNamespaceDevices } from "./seed";

async function requestLoginCode(team: Parameters<typeof enrollDevice>[0]) {
  const device = await enrollDevice(team);
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
    const device = await requestLoginCode(other);

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

test.describe("accepting a device at the namespace device limit", () => {
  test.skip(isCloud, "the cloud refuses with a billing message on a 402");

  test("the accept page names the device limit instead of a permission problem", async ({
    page,
  }) => {
    const team = await createTeam();
    const { owner, namespace } = team;
    const accepted = await enrollDevice(team);
    await updateDeviceStatus({
      ...buildRequestContext({ token: owner.token }),
      path: { uid: accepted.uid, status: "accept" },
    });
    limitNamespaceDevices(namespace, 1);
    const device = await requestLoginCode(team);

    await signInAndOpen(
      page,
      owner.username,
      `/accept-device?code=${device.code}`,
    );
    await page.getByRole("button", { name: "Accept device" }).click();

    await expect(
      page.getByText(
        "This namespace has reached its device limit. Remove a device or ask the instance administrator to raise the limit, then accept this device again.",
      ),
    ).toBeVisible();
    const { data } = await getDevice({
      ...buildRequestContext({ token: owner.token }),
      path: { uid: device.uid },
    });
    expect(data.status).toBe("pending");
  });
});
