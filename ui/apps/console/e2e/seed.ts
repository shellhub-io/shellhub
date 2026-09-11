import { expect } from "@playwright/test";
import type { SetupBody } from "@/client/model";
import { adminCLI, edition, errorReason } from "./stack";

export const adminUser: SetupBody = {
  name: "E2E Admin",
  email: "admin@e2e.test",
  username: "e2e-admin",
  password: "e2e-password",
  namespace: "e2e",
};

const probeTimeout = 5000;
const apiReadyTimeout = 60_000;
const seedTimeout = 60_000;

async function waitForAPI(baseURL: string) {
  await expect
    .poll(
      () =>
        fetch(`${baseURL}/api/info`, {
          signal: AbortSignal.timeout(probeTimeout),
        }).then((response) => `HTTP ${response.status}`, errorReason),
      {
        timeout: apiReadyTimeout,
        message: `${baseURL}/api/info never became usable`,
      },
    )
    .toBe("HTTP 200");
}

export async function seedInstance(baseURL: string) {
  await waitForAPI(baseURL);

  if (edition === "cloud") {
    adminCLI([
      "user",
      "create",
      adminUser.username,
      adminUser.password,
      adminUser.email,
    ]);
    adminCLI(["namespace", "create", adminUser.namespace, adminUser.username]);

    return;
  }

  const response = await fetch(`${baseURL}/api/setup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(adminUser),
    signal: AbortSignal.timeout(seedTimeout),
  }).catch((error: unknown) => {
    throw new Error(`POST ${baseURL}/api/setup failed: ${errorReason(error)}`);
  });

  if (!response.ok) {
    throw new Error(
      `POST ${baseURL}/api/setup returned ${response.status}: ${await response.text()}`,
    );
  }
}
