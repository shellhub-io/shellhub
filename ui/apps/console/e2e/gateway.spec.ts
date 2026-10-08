import { expect, test } from "@playwright/test";

test("the gateway refuses the server's internal routes", async ({
  request,
}) => {
  const metrics = await request.get("/internal/metrics");

  expect(metrics.status()).toBe(404);
});

test("the gateway serves an uncached install script that names its own address", async ({
  request,
}) => {
  const script = await request.get("/install.sh");

  expect(script.status()).toBe(200);
  expect(script.headers()["cache-control"]).toBe("no-store");
  expect(await script.text()).toContain(
    `SERVER_ADDRESS='${new URL(script.url()).origin}'`,
  );
});
