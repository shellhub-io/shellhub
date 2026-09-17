import { defineConfig } from "@playwright/test";
import { requireEnv } from "./e2e/env";

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./e2e/tmp/results",
  workers: 1,
  forbidOnly: !!process.env.CI,
  reporter: [
    ["list"],
    ["html", { outputFolder: "./e2e/tmp/report", open: "never" }],
  ],
  use: {
    baseURL: requireEnv("E2E_BASE_URL"),
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
});
