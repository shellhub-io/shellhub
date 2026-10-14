import { defineConfig } from "@playwright/test";
import { edition, requireEnv } from "./e2e/env";

export default defineConfig({
  testDir: "./e2e",
  outputDir: `./e2e/tmp/${edition}/results`,
  workers: 1,
  forbidOnly: !!process.env.CI,
  reporter: [
    ["list"],
    ["html", { outputFolder: `./e2e/tmp/${edition}/report`, open: "never" }],
  ],
  use: {
    baseURL: requireEnv("E2E_BASE_URL"),
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
});
