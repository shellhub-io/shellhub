import { defineConfig } from "@playwright/test";
import { edition, requireEnv } from "./e2e/env";

const instanceSpecs = /(setup|license|saml|admin)\.spec\.ts$/;
const terminalWithSaml = /Re-authentication with SAML/;

export default defineConfig({
  testDir: "./e2e",
  outputDir: `./e2e/tmp/${edition}/results`,
  workers: Number(process.env.E2E_WORKERS ?? 2),
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
  projects: [
    { name: "instance", testMatch: instanceSpecs, workers: 1 },
    {
      name: "instance-terminal",
      testMatch: /terminal\.spec\.ts$/,
      grep: terminalWithSaml,
      workers: 1,
      dependencies: ["instance"],
    },
    {
      name: "parallel",
      testIgnore: instanceSpecs,
      grepInvert: terminalWithSaml,
      dependencies: ["instance-terminal"],
    },
  ],
});
