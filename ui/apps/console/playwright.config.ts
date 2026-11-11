import { defineConfig } from "@playwright/test";
import { edition, instanceWide, requireEnv } from "./e2e/env";

const instanceWideTag = new RegExp(instanceWide.tag);

export default defineConfig({
  testDir: "./e2e",
  outputDir: `./e2e/tmp/${edition}/results`,
  workers: 3,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  failOnFlakyTests: !!process.env.CI,
  reporter: [
    ["list"],
    ["html", { outputFolder: `./e2e/tmp/${edition}/report`, open: "never" }],
  ],
  use: {
    baseURL: requireEnv("E2E_BASE_URL"),
    trace: process.env.CI ? "on-first-retry" : "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    {
      name: "namespaced",
      grepInvert: instanceWideTag,
      fullyParallel: true,
    },
    {
      name: "instance-wide",
      grep: instanceWideTag,
      dependencies: ["namespaced"],
      workers: 1,
    },
  ],
});
