import { expect, test } from "@playwright/test";
import { authDevice, getDevices, provisioningKeyCreate } from "@/client";
import { createTeam, signInAndOpen } from "./helpers";
import { buildShortId } from "./seed";
import { buildRequestContext } from "./api";
import { buildDeviceAuthRequest } from "./devices";

test.describe("Provisioning key plaintext", () => {
  test("a key's page reveals the plaintext an agent enrolls with", async ({
    page,
  }) => {
    const { owner, tenant } = await createTeam();
    const name = `e2e-${buildShortId()}`;
    const { data: issued } = await provisioningKeyCreate({
      ...buildRequestContext({ token: owner.token }),
      body: { name, mode: "manual" },
    });
    const { id, key, key_hint: hint } = issued;
    if (!id || !key || !hint) {
      throw new Error(
        `expected ${name} to be issued with an id, a key and a hint`,
      );
    }
    const masked = page.getByText(`${hint}••••••••••••••••`, { exact: true });
    const revealed = page.getByText(key, { exact: true });

    await signInAndOpen(
      page,
      owner.username,
      `/devices/add/fleet/${id}/activity`,
    );
    await page.getByRole("button", { name: "Details" }).click();
    await expect(masked).toBeVisible();
    await expect(revealed).toHaveCount(0);

    await page.getByRole("button", { name: "Show", exact: true }).click();

    await expect(revealed).toBeVisible();
    await expect(masked).toHaveCount(0);

    const device = buildDeviceAuthRequest();
    const { data: enrolled } = await authDevice({
      ...buildRequestContext(),
      body: { ...device, provisioning_key: await revealed.innerText() },
    });
    expect(enrolled.tenant_id).toBe(tenant);
    const { data: pending } = await getDevices({
      ...buildRequestContext({ token: owner.token }),
      query: { status: "pending" },
    });
    expect(pending.map((d) => d.name)).toContain(device.hostname);
  });
});
