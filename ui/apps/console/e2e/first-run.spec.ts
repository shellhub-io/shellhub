import { expect, test } from "@playwright/test";
import { updateDeviceStatus } from "@/client";
import { buildRequestContext } from "./api";
import { enrollDevice } from "./devices";
import { createTeam, signInAndOpen } from "./helpers";

test("the first run moves on once the namespace's first device is accepted", async ({
  page,
}) => {
  const team = await createTeam();

  await signInAndOpen(page, team.owner.username, "/dashboard");

  const step = (title: string) =>
    page.getByRole("listitem").filter({
      has: page.getByRole("heading", { name: title, exact: true }),
    });
  const install = step("Install the agent");
  const shell = step("Open a shell on it");

  await expect(install).toHaveAttribute("aria-current", "step");

  const device = await enrollDevice(team.tenant);
  await updateDeviceStatus({
    ...buildRequestContext({ token: team.owner.token }),
    path: { uid: device.uid, status: "accept" },
  });

  await expect(shell).toHaveAttribute("aria-current", "step", {
    timeout: 15_000,
  });
  await expect(install).toContainText(`Agent running on ${device.name}`);
});
