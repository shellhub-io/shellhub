import { describe, it, expect, vi, beforeEach } from "vitest";
import { getConfig, defaultConfig } from "@/env";
import { getAcceptErrorMessage, isSubscriptionBlocked } from "../acceptErrors";

const mockGetConfig = vi.mocked(getConfig);

describe("getAcceptErrorMessage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetConfig.mockReturnValue({ ...defaultConfig });
  });

  describe("402 Payment Required", () => {
    it("returns license copy for enterprise (non-cloud) on 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const msg = getAcceptErrorMessage({ status: 402 });
      expect(msg).toMatch(/license/i);
    });

    it("returns billing copy for cloud on 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage({ status: 402 });
      expect(msg).toMatch(/billing|subscription|plan/i);
    });

    it("cloud 402 message is distinct from enterprise 402 message", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const cloudMsg = getAcceptErrorMessage({ status: 402 });

      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const enterpriseMsg = getAcceptErrorMessage({ status: 402 });

      expect(cloudMsg).not.toBe(enterpriseMsg);
    });

    it("separates a subscription block from a device-limit hit on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });

      const limitMsg = getAcceptErrorMessage({ status: 402 });
      const subscriptionMsg = getAcceptErrorMessage({
        status: 402,
        message: "the namespace's subscription blocks new devices",
      });

      expect(subscriptionMsg).not.toBe(limitMsg);
      expect(limitMsg).toMatch(/device limit/i);
      expect(subscriptionMsg).not.toMatch(/device limit/i);
    });

    it("returns generic fallback for community on 402 (not billing copy)", () => {
      const msg = getAcceptErrorMessage({ status: 402 });
      expect(msg).not.toMatch(/billing|subscription|plan/i);
    });
  });

  describe("403 Forbidden", () => {
    it("returns namespace copy for 403", () => {
      const msg = getAcceptErrorMessage({ status: 403 });
      expect(msg).toMatch(/namespace|permission/i);
    });

    it("returns namespace copy even when enterprise is true (403 ignores enterprise flag)", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const msg = getAcceptErrorMessage({ status: 403 });
      expect(msg).toMatch(/namespace|permission/i);
    });
  });

  describe("409 Conflict", () => {
    it("returns rename copy for 409", () => {
      const msg = getAcceptErrorMessage({ status: 409 });
      expect(msg).toMatch(/rename|name|already exists/i);
    });
  });

  describe("unknown errors", () => {
    it("returns a generic fallback for an unrecognized status code", () => {
      const msg = getAcceptErrorMessage({ status: 500 });
      expect(msg).toBeTruthy();
      expect(typeof msg).toBe("string");
    });

    it("returns a generic fallback for a non-sdk error object", () => {
      const msg = getAcceptErrorMessage(new Error("network failure"));
      expect(msg).toBeTruthy();
      expect(typeof msg).toBe("string");
    });
  });

  describe("entityType parameter", () => {
    it("interpolates 'container' into enterprise 402 copy", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "enterprise",
      });
      const msg = getAcceptErrorMessage({ status: 402 }, "container");
      expect(msg).toMatch(/container limit/i);
      expect(msg).not.toMatch(/device/i);
    });

    it("interpolates 'container' into cloud 402 billing copy", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage({ status: 402 }, "container");
      expect(msg).toMatch(/container limit/i);
      expect(msg).not.toMatch(/device/i);
    });

    it("interpolates 'container' into 403 copy", () => {
      const msg = getAcceptErrorMessage({ status: 403 }, "container");
      expect(msg).toMatch(/containers/i);
      expect(msg).not.toMatch(/device/i);
    });

    it("interpolates 'container' into 409 copy", () => {
      const msg = getAcceptErrorMessage({ status: 409 }, "container");
      expect(msg).toMatch(/container/i);
      expect(msg).not.toMatch(/device/i);
    });

    it("interpolates 'container' into fallback copy", () => {
      const msg = getAcceptErrorMessage(new Error("boom"), "container");
      expect(msg).toMatch(/container/i);
      expect(msg).not.toMatch(/device/i);
    });
  });

  describe("canSubscribe parameter", () => {
    it("returns owner billing copy when canSubscribe is true on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage({ status: 402 }, "device", true);
      expect(msg).toMatch(/update your billing plan/i);
    });

    it("returns non-owner billing copy when canSubscribe is false on cloud 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage({ status: 402 }, "device", false);
      expect(msg).toMatch(/ask the namespace owner/i);
    });

    it("returns owner subscription copy when canSubscribe is true on cloud subscription-blocked 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage(
        { status: 402, message: "the namespace's subscription blocks new devices" },
        "device",
        true,
      );
      expect(msg).toMatch(/open billing/i);
    });

    it("returns non-owner subscription copy when canSubscribe is false on cloud subscription-blocked 402", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
      });
      const msg = getAcceptErrorMessage(
        { status: 402, message: "the namespace's subscription blocks new devices" },
        "device",
        false,
      );
      expect(msg).toMatch(/ask the namespace owner/i);
    });
  });
});

describe("isSubscriptionBlocked", () => {
  it("returns true for the subscription-blocked sentinel message", () => {
    expect(
      isSubscriptionBlocked({
        status: 402,
        message: "the namespace's subscription blocks new devices",
      }),
    ).toBe(true);
  });

  it("returns false for a regular 402", () => {
    expect(isSubscriptionBlocked({ status: 402 })).toBe(false);
  });

  it("returns false for a non-sdk error", () => {
    expect(isSubscriptionBlocked(new Error("boom"))).toBe(false);
  });
});
