import { defineConfig } from "@playwright/test";
import { baseURL } from "./e2e/stack";

export default defineConfig({
  testDir: "./e2e",
  outputDir: "./e2e/tmp/results",
  globalSetup: "./e2e/global-setup.ts",
  globalTeardown: "./e2e/global-teardown.ts",
  workers: 1,
  forbidOnly: !!process.env.CI,
  reporter: [
    ["list"],
    ["html", { outputFolder: "./e2e/tmp/report", open: "never" }],
  ],
  use: {
    baseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
});
