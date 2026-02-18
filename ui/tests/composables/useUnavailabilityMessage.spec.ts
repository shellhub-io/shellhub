import { describe, it, expect, afterEach, vi } from "vitest";
import { useUnavailabilityMessage } from "@/composables/useUnavailabilityMessage";
import { envVariables } from "@/envVariables";

vi.mock("@/envVariables", () => ({
  envVariables: {
    isCloud: false,
    isEnterprise: false,
  },
}));

describe("useUnavailabilityMessage", () => {
  afterEach(() => {
    envVariables.isCloud = false;
    envVariables.isEnterprise = false;
  });

  it.each([
    {
      label: "community",
      env: { isCloud: false, isEnterprise: false },
      expected: {
        title: "Unable to Connect to API",
        descriptionFragment: "trouble connecting to your ShellHub server",
        actionText: "Retry Connection",
        icon: "mdi-server-network-off",
        supportLink: { text: "Troubleshooting Guide", url: "https://docs.shellhub.io/troubleshooting/" },
      },
    },
    {
      label: "cloud",
      env: { isCloud: true, isEnterprise: true },
      expected: {
        title: "System Temporarily Unavailable",
        descriptionFragment: "temporarily down",
        actionText: "Retry Connection",
        icon: "mdi-cloud-alert",
        supportLink: { text: "View Status Page", url: "https://status.shellhub.io/" },
      },
    },
    {
      label: "enterprise",
      env: { isCloud: false, isEnterprise: true },
      expected: {
        title: "System Temporarily Unavailable",
        descriptionFragment: "currently unavailable",
        actionText: "Try Again",
        icon: "mdi-tools",
        supportLink: { text: "Report Issue", url: "https://github.com/shellhub-io/shellhub/issues" },
      },
    },
  ])("returns the $label message", ({ env, expected }) => {
    envVariables.isCloud = env.isCloud;
    envVariables.isEnterprise = env.isEnterprise;

    const message = useUnavailabilityMessage();

    expect(message.title).toBe(expected.title);
    expect(message.description).toContain(expected.descriptionFragment);
    expect(message.actionText).toBe(expected.actionText);
    expect(message.icon).toBe(expected.icon);
    expect(message.supportLink).toEqual(expected.supportLink);
  });
});
